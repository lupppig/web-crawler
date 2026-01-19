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

	// Handle graceful shutdown
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
	fmt.Printf("Crawler finished in %s\n", duration)
	fmt.Printf("Pages crawled: %d\n", crawlerInstance.Stats())
}
