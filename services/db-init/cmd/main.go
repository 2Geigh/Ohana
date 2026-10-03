package main

import (
	"log"

	"github.com/2Geigh/Ohana/db-init/pkg/database"
)

func main() {
	err := database.InitializeDB(database.DB)
	if err != nil {
		log.Fatalf("connect to database failed: %v", err)
	}
	defer database.DB.Close()

	err = database.Migrate(database.DB)
	if err != nil {
		log.Fatalf("database migration(s) failed: %v", err)
	}

	err = database.InitializeDomainBlacklist(database.DB)
	if err != nil {
		log.Fatalf("initialize domain blacklist failed: %v", err)
	}
}
