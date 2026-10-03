package main

import (
	"log"
	"sync"
	"sync/atomic"

	"github.com/2Geigh/Ohana/crawler/internal/crawling"
	"github.com/2Geigh/Ohana/db-init/pkg/database"
	"github.com/2Geigh/Ohana/db-init/pkg/models"
)

const (
	NUMBER_OF_CRAWLERS = 15
)

func init() {
	err := database.InitializeDB(database.DB)
	if err != nil {
		log.Fatalf("connect to database failed: %v", err)
	}

	for _, url := range crawling.SeedURLs {
		err := database.EnqueueLinks(
			[]models.Url{url},
			database.DB,
			&database.DatabaseMu,
		)
		if err != nil {
			log.Fatalf("enqueue seed URLs failed: %v", err)
		}
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
