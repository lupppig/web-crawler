package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/kehl-gopher/crawler/internal/crawler"
	"github.com/kehl-gopher/crawler/internal/storage"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("no .env file found, using environment variables")
	}

	dbCred := os.Getenv("DBCred")
	if dbCred == "" {
		fmt.Println("DBCred environment variable is required")
		os.Exit(1)
	}

	store := storage.NewStorage("crawledContent")
	if err := store.Connect(dbCred); err != nil {
		fmt.Printf("failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close(context.Background())

	seedURL := "https://nexford.edu/"
	crawlerInstance := crawler.NewCrawler(seedURL, 10, store)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down crawler...")
		cancel()
	}()

	startTime := time.Now()
	fmt.Println("Starting crawler...")
	crawlerInstance.Start(ctx)

	duration := time.Since(startTime)
	stats := crawlerInstance.Stats()

	fmt.Printf("\n--- Crawl Statistics ---\n")
	fmt.Printf("Total Duration: %s\n", duration)
	fmt.Printf("Total Pages Crawled: %d\n", stats.TotalPages)
	fmt.Printf("Successful Requests: %d\n", stats.SuccessCount)
	fmt.Printf("Failed Requests: %d\n", stats.FailureCount)

	if stats.TotalPages > 0 {
		avgTime := duration.Seconds() / float64(stats.TotalPages)
		fmt.Printf("Average Time Per Page: %.4f seconds\n", avgTime)
	}
}
