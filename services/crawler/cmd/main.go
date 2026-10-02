package main

import (
	"database/sql"
	"log"
	"sync"

	"github.com/2Geigh/Ohana/crawler/internal/crawling"
	"github.com/2Geigh/Ohana/db-init/pkg/database"
)

const (
	NUMBER_OF_CRAWLERS = 10
)

var (
	DB         *sql.DB = nil
	DatabaseMu sync.Mutex
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

		startCrawler func()
	)

	defer database.DB.Close()

	startCrawler = func() {
		wg.Add(1)

		go func() {
			crawling.Crawl(&wg)

			// Crawler replaces itself when it returns
			startCrawler()
		}()
	}

	for range NUMBER_OF_CRAWLERS {
		startCrawler()
	}

	wg.Wait()
}
