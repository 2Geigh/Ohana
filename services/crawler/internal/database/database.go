package database

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/2Geigh/Ohana/crawler/internal/models"
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
	DB *sql.DB = nil

	DatabaseMu sync.Mutex

	//go:embed migrations/*.sql
	embedMigrations embed.FS

	//go:embed data/*.json
	embedData embed.FS
)

func DequeueLinks(db *sql.DB, mu *sync.Mutex) ([]models.Url, error) {
	var (
		queueRows []models.Url
	)

	mu.Lock()
	defer mu.Unlock()

	tx, err := db.Begin()
	if err != nil {
		return queueRows, fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	// Get the first contiguous set of links with the same domain
	// Ex: If the first five rows of the queue
	//	   are all from google.com, return the
	//     first five rows.
	//
	//	   Otherwise, just return the first row.

	rows, err := tx.Query(
		`WITH first_value AS (
			SELECT fqdn AS value
			FROM crawler_queue
			ORDER BY id
			LIMIT 1
		),

		rows_to_dequeue AS MATERIALIZED (
			SELECT t.id
			FROM crawler_queue AS t
			CROSS JOIN first_value AS f
			WHERE t.fqdn IS NOT DISTINCT FROM f.value
			ORDER BY t.id
			LIMIT 1000
		),

		deleted AS (
			DELETE FROM crawler_queue AS t
			USING rows_to_dequeue AS d
			WHERE t.id = d.id
			RETURNING t.id, t.hyperlink
		)
			
		SELECT hyperlink
		FROM deleted
		ORDER BY id;`,
	)
	if err != nil {
		return queueRows, fmt.Errorf("tx query failed: %w", err)
	}

	for rows.Next() {
		var (
			hyperlink models.Url
		)

		err = rows.Scan(&hyperlink)
		if err != nil {
			return queueRows, fmt.Errorf("scan row to local variable failed: %w", err)
		}

		queueRows = append(queueRows, hyperlink)
	}

	err = rows.Close()
	if err != nil {
		return queueRows, fmt.Errorf("close rows failed: %w", err)
	}

	err = rows.Err()
	if err != nil {
		return queueRows, fmt.Errorf("rows: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return queueRows, fmt.Errorf("commit transaction failed: %w", err)
	}

	return queueRows, nil
}

func EnqueueLinks(urls []models.Url, db *sql.DB, mu *sync.Mutex) error {
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

		_, err = stmt.Exec(url, url.GetDomain().GetFQDN())
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

func InitializeDomainBlacklist(db *sql.DB) error {
	var (
		topThousandDomains = struct {
			asBytes  []byte
			asString string
			asJson   []RankedDomain
		}{}

		err error
	)

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

func InitializeDB() error {
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

	fmt.Println(dsn)

	if username == "" {
		log.Println("Warning: DB_USERNAME is empty. Connection might fail.")
	}

	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open database connection failed: %w", err)
	}

	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("verify database connection failed: %w", err)
	}
	log.Println("Database connection successful.")

	log.Println("Executing migrations...")
	err = migrate(DB)
	if err != nil {
		return fmt.Errorf("database migration(s) failed: %w", err)
	}

	return nil
}

func ReportDatabaseHealth() {
	// for {
	stats := DB.Stats()
	log.Printf(`[DB STATS] InUse: %d | Idle: %d | Open: %d | WaitCount: %d`,
		stats.InUse, stats.Idle, stats.OpenConnections, stats.WaitCount)

	// time.Sleep(5 * time.Second)
	// }
}

func migrate(db *sql.DB) error {

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
