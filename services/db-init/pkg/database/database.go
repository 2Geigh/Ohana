package database

import (
	"bufio"
	"database/sql"
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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

	//go:embed data/*.txt
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

		isBlacklisted, err = url.GetDomain().IsBlacklisted(db, &DatabaseMu)
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

func refreshPornographicDomains(
	tx *sql.Tx,
) error {
	log.Println("Starting refresh of remote pornographic domain blacklist!")

	const (
		// Source repository: https://github.com/Bon-Appetit/porn-domains
		pornDomainListUrl string = "https://test@raw.githubusercontent.com/Bon-Appetit/porn-domains/refs/heads/main/block.570f9ed0d2.dd0b4e.txt"
	)

	pornDomainList := struct {
		asBytes []byte
		asText  string
	}{}

	log.Println("GETting remote pornographic domain blacklist...")
	resp, err := http.Get(pornDomainListUrl)
	if err != nil {
		return fmt.Errorf("GET porn domain list failed: %w", err)
	}
	log.Println("GET remote pornographic domain blacklist complete")

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", resp.Status)
	}

	log.Println("Reading remote pornographic domain blacklist...")
	pornDomainList.asBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read GET response failed: %w", err)
	}
	defer resp.Body.Close()
	log.Println("Finished reading remote pornographic domain blacklist")

	pornDomainList.asText = string(pornDomainList.asBytes)

	stmt, err := tx.Prepare(
		`INSERT INTO domain_blacklist (domain) VALUES ($1);`,
	)
	if err != nil {
		return fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	scanner := bufio.NewScanner(strings.NewReader(pornDomainList.asText))

	log.Println("Scanning acquired pornographic domain blacklist...")
	for scanner.Scan() {
		var (
			domain string = strings.TrimSpace(string(scanner.Text()))
		)

		if strings.TrimSpace(domain) == "" {
			continue
		}

		_, err = stmt.Exec(domain)
		if err != nil {
			return fmt.Errorf("execute stmt with %s failed: %w", domain, err)
		}
	}
	err = scanner.Err()
	if err != nil {
		return fmt.Errorf("scanner error")
	}
	log.Println("Finished scanning acquired pornographic domain blacklist.")

	return nil
}

func refreshLocalDomainBlacklist(tx *sql.Tx) error {
	log.Println("Starting refresh of local domain blacklist!")

	log.Println("Opening domain_blacklist.txt ...")
	file, err := embedData.Open("data/domain_blacklist.txt")
	if err != nil {
		return fmt.Errorf("read file failed: %w", err)
	}
	defer file.Close()
	log.Println("Opened domain_blacklist.txt")

	log.Println("Create scanner of domain_blacklist.txt...")
	scanner := bufio.NewScanner(file)
	log.Println("Created scanner of domain_blacklist.txt")

	stmt, err := tx.Prepare(
		`INSERT INTO domain_blacklist (domain) values ($1);`,
	)
	if err != nil {
		return fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	log.Println("Scanning domain_blacklist.txt ...")
	for scanner.Scan() {
		// log.Println("Preparing statement for", scanner.Text(), "...")
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}

		_, err = stmt.Exec(scanner.Text())
		if err != nil {
			return fmt.Errorf("execute stmt failed: %w", err)
		}
	}
	log.Println("Finished scanning domain_blacklist.txt")

	err = scanner.Err()
	if err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

func RefreshDatabaseDomainBlacklist(
	db *sql.DB,
) error {
	var (
		wg  sync.WaitGroup
		err error
	)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback()

	tx.Exec(`DELETE FROM domain_blacklist *;`)

	wg.Go(func() {
		e := refreshLocalDomainBlacklist(tx)
		if e != nil {
			err = fmt.Errorf("refresh local domain blacklist failed: %w", e)
		}
	})

	wg.Go(func() {
		e := refreshPornographicDomains(tx)
		if e != nil {
			err = fmt.Errorf("refresh remote pornographic domain blacklist failed: %w", e)
		}
	})

	wg.Wait()
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit tx failed: %w", err)
	}

	log.Println("Domain blacklist completed successfully")
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
