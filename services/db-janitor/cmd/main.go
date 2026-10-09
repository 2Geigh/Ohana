package main

import (
	"log"
	"sync"
	"time"

	"github.com/2Geigh/Ohana/db-init/pkg/database"
	"github.com/2Geigh/Ohana/db-janitor/internal/db"
)

func init() {
	err := database.InitializeDB(&db.DB)
	if err != nil {
		log.Fatalf("connect to database failed: %v", err)
	}
}

func main() {
	var wg sync.WaitGroup

	defer db.DB.Close()

	wg.Go(func() {
		for {
			err := db.CleanupDatabase(db.DB)
			if err != nil {
				log.Fatal("cleanup error: %w", err)
			}

			time.Sleep(30 * time.Minute)
		}
	})

	wg.Wait()
}
