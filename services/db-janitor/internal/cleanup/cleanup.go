package cleanup

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq"
)

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

	result, err := tx.Exec(`
		DELETE FROM sites
		WHERE fqdn IN (
			SELECT domain
			FROM domain_blacklist
		);
	`)
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
		"Successfully completed database purge of sites and pages from blacklisted domains, directly affecting %d rows",
		rowsAffected,
	)

	return nil
}
