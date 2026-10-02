package models

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	parser "github.com/Cgboal/DomainParser"
)

type (
	Domain string
)

func (d Domain) GetFQDN() Domain {
	/*

		>>> getDomain(https://www.youtube.com/watch?v=dQw4w9WgXcQ)
		youtube.com

	*/

	if len(d) == 0 {
		return d
	}

	domainWithoutProtocol := d.StripProtocol()

	domainWithoutRoutes, _, _ := strings.Cut(string(domainWithoutProtocol), "/")

	parser := parser.NewDomainParser()
	return Domain(parser.GetFQDN(string(domainWithoutRoutes)))
}

func (d Domain) HasBeenRequestedTooRecently(politeness_interval time.Duration, databaseMu *sync.Mutex, db *sql.DB) (bool, error) {
	var (
		fqdn        = d.GetFQDN()
		lastCrawled time.Time

		// better to be too
		// polite than not nice enough
		// to an API
		hasBeenCrawledTooRecently bool = true
	)

	databaseMu.Lock()
	defer databaseMu.Unlock()

	stmt, err := db.Prepare(
		`SELECT date_last_crawled
		FROM sites
		WHERE fqdn = $1;`,
	)
	if err != nil {
		return hasBeenCrawledTooRecently, fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(fqdn).Scan(&lastCrawled)
	if err == sql.ErrNoRows {
		hasBeenCrawledTooRecently = false
	} else if err != nil {
		return hasBeenCrawledTooRecently, fmt.Errorf("execute stmt failed: %w", err)
	}

	if time.Since(lastCrawled) > politeness_interval {
		hasBeenCrawledTooRecently = false
	}

	return hasBeenCrawledTooRecently, nil
}

func (d Domain) IsBlacklisted(db *sql.DB) (bool, error) {
	var (
		fqdn = d.GetFQDN()

		exists bool
	)

	err := db.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM domain_blacklist WHERE domain = $1);`,
		fqdn,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("query failed: %w", err)
	}

	return exists, nil
}

func (d Domain) StripProtocol() Domain {
	_, domainWithoutProtocol, includesProtocol := strings.Cut(string(d), "://")

	if !includesProtocol {
		return d
	}

	return Domain(domainWithoutProtocol)
}
