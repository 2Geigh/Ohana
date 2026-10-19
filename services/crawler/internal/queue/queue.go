package queue

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/2Geigh/Ohana/crawler/internal/connection"
	"github.com/pressly/goose/v3"
)

var (
	//go:embed migrations/*.sql
	embedMigrations embed.FS
)

func Migrate(db *sql.DB) error {
	log.Println("Executing migrations...")

	goose.SetBaseFS(embedMigrations)

	err := goose.SetDialect("sqlite")
	if err != nil {
		return fmt.Errorf("goose: set database dialect failed: %w", err)
	}

	err = goose.Up(connection.Queue, "migrations")
	if err != nil {
		return fmt.Errorf("goose: apply migrations failed: %w", err)
	}

	return nil
}
