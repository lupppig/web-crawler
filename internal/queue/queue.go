package queue

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"
)

type Queue struct {
	queueLock *sync.Mutex
	url       []string
	size      int
	totalSize int
	queued    sync.Map
}

func NewQueue() *Queue {
	return &Queue{
		url:       make([]string, 0),
		queueLock: new(sync.Mutex),
		queued:    sync.Map{},
	}
}

func (q *Queue) Enqueue(data string) {
	normalizedURL := strings.TrimSuffix(data, "/")
	hashed := hashURL(normalizedURL)
	if _, exists := q.queued.LoadOrStore(hashed, true); exists {
		return
	}

	q.queueLock.Lock()
	q.url = append(q.url, data)
	q.size++
	q.totalSize++
	q.queueLock.Unlock()
}

func (q *Queue) Dequeue() string {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()

	if len(q.url) == 0 {
		return ""
	}

	url := q.url[0]
	q.url = q.url[1:]
	q.size--
	return url
}

func (q *Queue) IsEmpty() bool {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	return q.size == 0
}

func (q *Queue) Size() int {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	return q.size
}

func (q *Queue) TotalSize() int {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	return q.totalSize
}

func hashURL(url string) string {
	sha := sha256.New()
	sha.Write([]byte(url))
	hashed := sha.Sum(nil)
	return hex.EncodeToString(hashed)
}
