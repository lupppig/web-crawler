package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kehl-gopher/crawler/internal/parser"
	"github.com/kehl-gopher/crawler/internal/queue"
	"github.com/kehl-gopher/crawler/internal/storage"
	"github.com/temoto/robotstxt"
)

type Crawler struct {
	seed          string
	queue         *queue.Queue
	storage       *storage.Storage
	visited       sync.Map
	workers       int
	pageCount     int32
	activeWorkers int32
}

func NewCrawler(seed string, workers int, storage *storage.Storage) *Crawler {
	return &Crawler{
		seed:    seed,
		queue:   queue.NewQueue(),
		storage: storage,
		workers: workers,
	}
}

func (c *Crawler) Start(ctx context.Context) {
	c.queue.Enqueue(c.seed)
	c.markVisited(c.seed)

	var wg sync.WaitGroup

	for i := 0; i < c.workers; i++ {
		wg.Add(1)
		go c.worker(ctx, &wg)
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	idleCount := 0

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return
		case <-ticker.C:
			if c.queue.IsEmpty() && atomic.LoadInt32(&c.activeWorkers) == 0 {
				idleCount++
				if idleCount > 5 {
					return
				}
			} else {
				idleCount = 0
			}
		}
	}
}

func (c *Crawler) worker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			urlStr := c.queue.Dequeue()
			if urlStr == "" {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			atomic.AddInt32(&c.activeWorkers, 1)
			c.processURL(ctx, urlStr)
			atomic.AddInt32(&c.activeWorkers, -1)
		}
	}
}

func (c *Crawler) processURL(ctx context.Context, urlStr string) {
	if checked, err := c.checkRobotstxt(urlStr); err != nil || checked == "" {
		fmt.Printf("Skipping %s: %v\n", urlStr, err)
		return
	}

	resp, err := c.sendRequest(urlStr)
	if err != nil {
		fmt.Printf("Error fetching %s: %v\n", urlStr, err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	pageData, err := parser.Parse(bodyBytes, urlStr)
	if err != nil {
		return
	}

	content := storage.CrawlContent{
		Title:   pageData.Title,
		Body:    pageData.Body,
		Path:    urlStr,
		AddedAt: time.Now(),
	}

	if err := c.storage.AddContent(ctx, content); err != nil {
		fmt.Printf("Error saving %s: %v\n", urlStr, err)
	}

	atomic.AddInt32(&c.pageCount, 1)

	for _, link := range pageData.Links {
		if c.shouldVisit(link) {
			c.markVisited(link)
			c.queue.Enqueue(link)
		}
	}
}

func (c *Crawler) sendRequest(urlStr string) (*http.Response, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MyCrawler/1.0)")
	return client.Do(req)
}

func (c *Crawler) checkRobotstxt(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	robotsURL := u.Scheme + "://" + u.Host + "/robots.txt"
	resp, err := c.sendRequest(robotsURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	rob, err := robotstxt.FromResponse(resp)
	if err != nil {
		return "", err
	}

	group := rob.FindGroup("*")
	if group.Test(uri) {
		return uri, nil
	}
	return "", fmt.Errorf("robots.txt disallowed")
}

func (c *Crawler) shouldVisit(urlStr string) bool {
	normalized := strings.TrimSuffix(urlStr, "/")
	if _, ok := c.visited.Load(normalized); ok {
		return false
	}
	return true
}

func (c *Crawler) markVisited(urlStr string) {
	normalized := strings.TrimSuffix(urlStr, "/")
	c.visited.Store(normalized, true)
}

func (c *Crawler) Stats() int32 {
	return atomic.LoadInt32(&c.pageCount)
}
