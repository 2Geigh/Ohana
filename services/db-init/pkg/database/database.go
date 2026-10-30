package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"sync"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

var (
	DatabaseMu sync.Mutex = sync.Mutex{}

	//go:embed migrations/*.sql
	embedMigrations embed.FS
)

func InitializeDB(db **sql.DB) error {
	log.Println("Connecting to Postgresql...")

	var (
		username string = os.Getenv("DB_USERNAME")
		password string = os.Getenv("DB_PASSWORD")
		dbHost   string = os.Getenv("DB_HOST")
		dbPort   string = os.Getenv("DB_CONTAINER_PORT")
		dbName   string = os.Getenv("DB_NAME")

		dsn string = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", username, password, dbHost, dbPort, dbName)

		err error
	)

	if username == "" {
		log.Println("Warning: DB_USERNAME is empty. Connection might fail.")
	}

	*db, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open database connection failed: %w", err)
	}

	err = (*db).Ping()
	if err != nil {
		return fmt.Errorf("verify database connection failed: %w", err)
	}
	log.Println("Database connection successful.")

	return nil
}

func Migrate(
	db *sql.DB,
) error {
	log.Println("Executing migrations...")

	goose.SetBaseFS(embedMigrations)

	err := goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("Goose: set database dialect failed: %w", err)
	}

	err = goose.Up(db, "migrations")
	if err != nil {
		return fmt.Errorf("Goose: apply migrations failed: %w", err)
	}

	return nil
}
