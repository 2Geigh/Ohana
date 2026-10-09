package main

import (
	"log"

	"github.com/2Geigh/Ohana/db-init/internal/connection"
	"github.com/2Geigh/Ohana/db-init/pkg/database"
)

func init() {
	err := database.InitializeDB(&connection.DB)
	if err != nil {
		log.Fatalf("connect to database failed: %v", err)
	}

}

func main() {
	defer connection.DB.Close()

	err := database.Migrate(connection.DB)
	if err != nil {
		log.Fatalf("database migration(s) failed: %v", err)
	}
}
