package main

import (
	"database/sql"
	"log"
	"sync"
	"sync/atomic"

	"github.com/2Geigh/Ohana/crawler/internal/connection"
	"github.com/2Geigh/Ohana/crawler/internal/crawling"
	"github.com/2Geigh/Ohana/crawler/internal/queue"
	"github.com/2Geigh/Ohana/db-init/pkg/database"
	"github.com/2Geigh/Ohana/db-init/pkg/models"

	_ "github.com/ncruces/go-sqlite3/driver"
)

const (
	MAXIMUM_NUMBER_OF_CRAWLERS = 20
)

func init() {
	err := database.InitializeDB(&connection.DB)
	if err != nil {
		log.Fatalf("connect to database failed: %v", err)
	}

	log.Println("Connecting to embedded queue...")
	log.Println("Successfully connected to embedded queue.")

	connection.Queue, err = sql.Open("sqlite3", "/data/queue/queue.db")
	if err != nil {
		log.Fatalf("connect to embedded queue failed: %v", err)
	}

	log.Println("Migrating embedded queue...")

	err = queue.Migrate(connection.Queue)
	if err != nil {
		log.Fatalf("migrate embedded queue failed: %v", err)
	}

	log.Println("Successfully migrated embedded queue.")
	log.Println("Enqueing seed URLs...")

	for _, url := range crawling.SeedURLs {
		err := queue.Enqueue(connection.Queue, &connection.QueueMu, []models.Url{url})
		if err != nil {
			log.Fatalf("enqueue seed URLs failed: %v", err)
		}
	}

	log.Println("Successfully enqueued seed URLs")
}

func main() {
	var (
		wg sync.WaitGroup

		crawlIteration atomic.Int64
	)

	defer connection.DB.Close()

	wg.Go(func() {
		for {
			if crawlIteration.Load() < 10 {
				continue
			}

			err := crawling.CleanCrawlerQueue(
				connection.Queue,
				&connection.QueueMu,
				connection.DB,
			)
			if err != nil {
				log.Fatalf("clean database crawler queue failed: %v", err)
			}

			crawlIteration.Swap(0)
		}
	})

	for n := range MAXIMUM_NUMBER_OF_CRAWLERS {
		wg.Go(func() {
			connection.QueueMu.Lock()
			crawler := crawling.Crawler{
				Id:               n,
				Queue:            connection.Queue,
				CrawlerMu:        &connection.QueueMu,
				GlobalDb:         connection.DB,
				GlobalDbMu:       &database.DatabaseMu,
				NumberOfCrawlers: MAXIMUM_NUMBER_OF_CRAWLERS,
				CrawlIteration:   &crawlIteration,
			}
			connection.QueueMu.Unlock()

			crawler.Crawl()
		})
	}

	wg.Wait()
}
