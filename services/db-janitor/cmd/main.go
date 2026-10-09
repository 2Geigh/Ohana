package main

import (
	"log"
	"sync"
	"time"

	"github.com/2Geigh/Ohana/db-init/pkg/database"
	"github.com/2Geigh/Ohana/db-janitor/internal/cleanup"
	"github.com/2Geigh/Ohana/db-janitor/internal/connection"
)

func init() {
	err := database.InitializeDB(&connection.DB)
	if err != nil {
		log.Fatalf("connect to database failed: %v", err)
	}
}

func main() {
	var (
		wg sync.WaitGroup
	)

	defer connection.DB.Close()

	wg.Go(func() {
		for {
			log.Println("Refreshing domain blacklist...")
			err := database.RefreshDatabaseDomainBlacklist(
				connection.DB,
			)
			if err != nil {
				log.Fatalf("refresh domain blacklist failed: %v", err)
			}
			log.Println("Successfully completed domain blacklist refresh")

			log.Println("Commencing database purge...")
			err = cleanup.PurgeDatabase(
				connection.DB,
				&database.DatabaseMu,
			)
			if err != nil {
				log.Fatalf("Purge sites and pages from blacklisted domains failed: %v", err)
			}

			time.Sleep(30 * time.Minute)
		}
	})

	wg.Wait()
}
