package cleanup

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq"
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
