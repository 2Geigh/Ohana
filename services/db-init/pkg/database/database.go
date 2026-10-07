package database

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/2Geigh/Ohana/db-init/pkg/models"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type (
	RankedDomain struct {
		Position int     `json:"position"`
		Domain   string  `json:"domain"`
		Count    int     `json:"count"`
		Etv      float64 `json:"etv"`
	}
)

var (
	DatabaseMu sync.Mutex = sync.Mutex{}

	//go:embed migrations/*.sql
	embedMigrations embed.FS

	//go:embed data/*.json
	embedData embed.FS
)

func DequeueLink(
	db *sql.DB,
	mu *sync.Mutex,
) (
	models.Url,
	error,
) {
	var (
		id   int64
		link models.Url
	)

	mu.Lock()
	defer mu.Unlock()

	tx, err := db.Begin()
	if err != nil {
		return link, fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	err = tx.QueryRow(
		`SELECT id, hyperlink
		FROM crawler_queue
		ORDER BY date_added ASC
		LIMIT 1;`,
	).Scan(
		&id,
		&link,
	)
	if err == sql.ErrNoRows {
		return link, nil
	}
	if err != nil {
		return link, fmt.Errorf("rows: %w", err)
	}

	_, err = tx.Exec(
		`DELETE FROM crawler_queue
		WHERE id = $1`,
		id,
	)
	if err != nil {
		return link, fmt.Errorf("DELETE %s (id=%d) FROM crawler_queue failed: %w", link, id, err)
	}

	err = tx.Commit()
	if err != nil {
		return link, fmt.Errorf("commit transaction failed: %w", err)
	}

	return link, nil
}

func EnqueueLinks(
	urls []models.Url,
	db *sql.DB,
	mu *sync.Mutex,
) error {
	mu.Lock()
	defer mu.Unlock()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	for _, url := range urls {
		var (
			isBlacklisted bool
		)

		isBlacklisted, err = url.GetDomain().IsBlacklisted(db)
		if err != nil {
			return fmt.Errorf("determine url blacklist status failed: %w", err)
		}

		if isBlacklisted {
			continue
		}

		stmt, err := tx.Prepare(
			`INSERT INTO crawler_queue (
				hyperlink, fqdn
			) VALUES ($1, $2)
			ON CONFLICT (hyperlink) DO NOTHING;`,
		)
		if err != nil {
			return fmt.Errorf("prepare statement failed: %w", err)
		}

		_, err = stmt.Exec(url.SanitizeToEnqueue(), url.GetDomain().GetFQDN())
		if err != nil {
			return fmt.Errorf("execute statement failed: %w", err)
		}

		err = stmt.Close()
		if err != nil {
			return fmt.Errorf("close statement failed: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}

	return nil
}

func InitializeDB(
	db **sql.DB,
) error {
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

func InitializeDomainBlacklist(
	db *sql.DB,
	mu *sync.Mutex,
) error {
	var (
		topThousandDomains = struct {
			asBytes  []byte
			asString string
			asJson   []RankedDomain
		}{}

		err error
	)
	mu.Lock()
	defer mu.Unlock()

	topThousandDomains.asBytes, err = embedData.ReadFile("data/ranked_domains.json")
	if err != nil {
		return fmt.Errorf("read file failed: %w", err)
	}

	err = json.Unmarshal(topThousandDomains.asBytes, &topThousandDomains.asJson)
	if err != nil {
		return fmt.Errorf("unmarshal json failed: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback()

	tx.Exec(`DELETE FROM domain_blacklist *;`)

	for _, entry := range topThousandDomains.asJson {
		stmt, err := tx.Prepare(
			`INSERT INTO domain_blacklist (domain) values ($1);`,
		)
		if err != nil {
			return fmt.Errorf("prepare stmt failed: %w", err)
		}

		_, err = stmt.Exec(entry.Domain)
		if err != nil {
			return fmt.Errorf("execute stmt failed: %w", err)
		}
		stmt.Close()
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit tx failed: %w", err)
	}

	log.Println("Domain blacklist initialized successfully")

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

func ReportDatabaseHealth(db *sql.DB) {
	stats := db.Stats()
	log.Printf(`[DB STATS] InUse: %d | Idle: %d | Open: %d | WaitCount: %d`,
		stats.InUse, stats.Idle, stats.OpenConnections, stats.WaitCount)
}
