package main

import (
	"database/sql"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/2Geigh/Ohana/crawler/internal/connection"
	"github.com/2Geigh/Ohana/crawler/internal/crawling"
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

	log.Println("Enqueing seed URLs...")

	for _, url := range crawling.SeedURLs {
		err := database.EnqueueLinks(
			[]models.Url{url},
			connection.Queue,
			&connection.QueueMu,
		)
		if err != nil {
			log.Fatalf("enqueue seed URLs failed: %v", err)
		}
	}

	log.Println("Successfully enqueued seed URLs")
}

func main() {
	var (
		wg sync.WaitGroup

		crawlerMu sync.Mutex

		crawlerQueues    = make(crawling.Queues)
		numberOfCrawlers atomic.Int64
		crawlIteration   atomic.Int64
		crawlerId        int64 = 0
	)

	defer connection.DB.Close()

	wg.Go(func() {
		for {
			if crawlIteration.Load() < 5 {
				continue
			}

			err := crawling.CleanCrawlerQueue(
				connection.DB,
				&database.DatabaseMu,
			)
			if err != nil {
				log.Fatalf("clean database crawler queue failed: %v", err)
			}

			crawlIteration.Swap(0)
		}
	})

	for {

		newLink, err := database.DequeueLink(connection.DB, &database.DatabaseMu)
		if err != nil {
			log.Printf("dequeue links from database crawler queue failed: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		fqdn := newLink.GetDomain().GetFQDN()

		crawlerMu.Lock()
		_, isDomainOwnedByACrawler := crawlerQueues[fqdn]
		crawlerMu.Unlock()

		if isDomainOwnedByACrawler {
			crawlerMu.Lock()
			crawlerQueues[fqdn] = append(crawlerQueues[fqdn], newLink)
			crawlerMu.Unlock()
			continue
		}

		if strings.TrimSpace(string(fqdn)) != "" {
			crawlerMu.Lock()
			crawlerQueues[fqdn] = []models.Url{newLink}
			crawlerMu.Unlock()
		}

		if numberOfCrawlers.Load() >= MAXIMUM_NUMBER_OF_CRAWLERS {
			continue
		}

		wg.Go(func() {
			numberOfCrawlers.Add(1)
			defer numberOfCrawlers.Add(-1)

			crawlerId += 1

			crawler := crawling.Crawler{
				Id:               crawlerId,
				Queues:           &crawlerQueues,
				Fqdn:             fqdn,
				CrawlerMu:        &crawlerMu,
				Db:               connection.DB,
				DbMu:             &database.DatabaseMu,
				NumberOfCrawlers: &numberOfCrawlers,
				CrawlIteration:   &crawlIteration,
			}

			crawler.Crawl()

			crawlerMu.Lock()
			delete(crawlerQueues, fqdn)
			crawlerMu.Unlock()
		})
	}

	wg.Wait()
}
