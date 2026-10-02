package main

import (
	"log"
	"sync"
	"sync/atomic"

	"github.com/2Geigh/Ohana/crawler/internal/crawling"
	"github.com/2Geigh/Ohana/db-init/pkg/database"
)

const (
	NUMBER_OF_CRAWLERS = 10
)

func init() {
	err := database.InitializeDB()
	if err != nil {
		log.Fatalf("connect to database failed: %v", err)
	}
}

func main() {
	var (
		wg sync.WaitGroup

		startCrawler   func()
		crawlIteration atomic.Int64
	)

	defer database.DB.Close()

	startCrawler = func() {
		wg.Add(1)

		go func() {
			crawling.Crawl(&wg, &crawlIteration)

			// Crawler replaces itself when it returns
			startCrawler()
		}()
	}

	for range NUMBER_OF_CRAWLERS {
		startCrawler()
	}

	go crawling.CleanCrawlerQueue(database.DB, &database.DatabaseMu, &crawlIteration)

	wg.Wait()
}
