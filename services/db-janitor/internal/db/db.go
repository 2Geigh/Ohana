package db

import (
	"bufio"
	"database/sql"
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	_ "github.com/lib/pq"
)

var (
	DB *sql.DB

	//go:embed data/*.txt
	embedData embed.FS
)

// Delete all sites in the database that are of blacklisted domains
func PurgeDatabase(
	db *sql.DB,
	mu *sync.Mutex,
) error {
	mu.Lock()
	defer mu.Unlock()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`DELETE FROM sites
		WHERE EXISTS (
			SELECT 1
			FROM domain_blacklist
			WHERE sites.fqdn = domain_blacklist.domain
			OR (
				domain_blacklist.domain LIKE '*.%'
				AND right(sites.fqdn, length(substr(domain_blacklist.domain, 2))) = substr(domain_blacklist.domain, 2)
			)
		);`,
	)
	if err != nil {
		return fmt.Errorf("delete blacklisted sites failed: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get number of deleted sites failed: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit tx failed: %w", err)
	}

	log.Printf(
		"Successfully completed database purge of %d sites and their corresponding pages from blacklisted domains",
		rowsAffected,
	)

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

func ReportDatabaseHealth(db *sql.DB) {
	stats := db.Stats()
	log.Printf(`[DB STATS] InUse: %d | Idle: %d | Open: %d | WaitCount: %d`,
		stats.InUse, stats.Idle, stats.OpenConnections, stats.WaitCount)
}

func refreshLocalDomainBlacklist(
	tx *sql.Tx,
) error {
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
