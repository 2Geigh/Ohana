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
	var (
		wg sync.WaitGroup
	)

	defer db.DB.Close()

	wg.Go(func() {
		for {
			log.Println("Refreshing domain blacklist...")
			err := db.RefreshDatabaseDomainBlacklist(
				db.DB,
			)
			if err != nil {
				log.Fatalf("refresh domain blacklist failed: %v", err)
			}
			log.Println("Successfully completed domain blacklist refresh")

			log.Println("Commencing database purge...")
			err = db.PurgeDatabase(
				db.DB,
				&database.DatabaseMu,
			)
			if err != nil {
				log.Fatalf("Purge sites and pages from blacklisted domains failed: %v", err)
			}

			time.Sleep(30 * time.Minute)
		}
	})

	wg.Go(func() {
		for {
			time.Sleep(10 * time.Second)
			db.ReportDatabaseHealth(db.DB)
		}
	})

	wg.Wait()
}
