package main

import (
	"context"
	"database/sql"
	"runtime"
	"sync"

	"github.com/2Geigh/Ohana/indexer/internal/database"
	// Import pgvector, onnxruntime, goquery, etc.
)

type PageTask struct {
	PageID int
	SiteID int
	URL    string
}

func init() {
	database.InitializeDB()
}

func main() {
	ctx := context.Background()

	// Configure Postgres connection pool

	// Initialize ONNX Runtime session for embeddings here (load once)
	// Initialize NLP/Stopword configurations here (load once)

	taskChannel := make(chan PageTask, 100)
	var wg sync.WaitGroup

	// Determine concurrent worker count (e.g., matching CPU cores)
	workerCount := runtime.NumCPU()

	// Spawn workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go indexerWorker(ctx, i, database.DB, taskChannel, &wg)
	}

	// Queue manager: Fetch from DB and feed channel
	go func() {
		for {
			// Select batch of URLs from indexer_queue
			// rows, _ := dbpool.Query(ctx, "SELECT indexer_queue.page_id, indexer_queue.site_id, pages.link FROM indexer_queue LEFT JOIN pages ON pages.id = indexer_queue.page_id LIMIT 100")

			// Iterate rows and send to taskChannel
			// If queue is empty, time.Sleep(1 * time.Second)
		}
	}()

	wg.Wait()
}

func indexerWorker(ctx context.Context, id int, db *sql.DB, tasks <-chan PageTask, wg *sync.WaitGroup) {
	defer wg.Done()

	// for task := range tasks {
	// 1. Fetch HTML from DB
	// 2. Parse HTML text via goquery / x/net/html
	// 3. Detect language
	// 4. Extract keywords (Stopword removal + Tokenization)
	// 5. Generate Embedding via ONNX Runtime
	// 6. Execute Postgres Transaction (UPDATE pages, DELETE queue, INSERT keywords)
	// }
}
