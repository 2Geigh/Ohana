package db

import (
	"bufio"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"strings"
	"sync"

	_ "github.com/lib/pq"
)

var (
	DB *sql.DB

	//go:embed data/*
	embedData embed.FS
)

func CleanupDatabase(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback()

	log.Println("Refreshing domain blacklist...")

	err = refreshDatabaseDomainBlacklist(tx)
	if err != nil {
		log.Printf("WARNING: refresh domain blacklist failed: %v", err)
	}

	log.Println("Successfully completed domain blacklist refresh")
	log.Println("Commencing database purge...")

	err = purgeDatabase(tx)
	if err != nil {
		log.Printf("WARNING: Purge sites and pages from blacklisted domains failed: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit tx failed: %w", err)
	}

	return nil
}

// Delete all sites in the database that are of blacklisted domains
func purgeDatabase(tx *sql.Tx) error {
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

	log.Printf(
		"Successfully completed database purge of %d sites and their corresponding pages from blacklisted domains",
		rowsAffected,
	)
	return nil
}

func refreshDatabaseDomainBlacklist(tx *sql.Tx) error {
	var (
		wg  sync.WaitGroup
		err error
	)

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

	log.Println("Domain blacklist completed successfully")
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

func refreshPornographicDomains(tx *sql.Tx) error {
	log.Println("Starting refresh of remote pornographic domain blacklist!")

	log.Println("Opening Bon-Appetit/porn-domains blocklist...")
	matches, err := fs.Glob(embedData, "data/porn-domains/block.*.txt")
	if err != nil {
		return fmt.Errorf("couldn't find blocklist: %w", err)
	}
	if len(matches) < 1 {
		return fmt.Errorf("blockfile not found: %w", err)
	}

	blocklist, err := embedData.ReadFile(matches[0])
	if err != nil {
		return fmt.Errorf("read blockfile failed: %w", err)
	}

	stmt, err := tx.Prepare(
		`INSERT INTO domain_blacklist (domain) VALUES ($1);`,
	)
	if err != nil {
		return fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	scanner := bufio.NewScanner(strings.NewReader(string(blocklist)))

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
