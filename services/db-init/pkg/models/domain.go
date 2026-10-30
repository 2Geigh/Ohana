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

	var (
		domainWithoutAmpersand string
		domainWithoutPercent   string
	)
	for range 2 {
		domainWithoutAmpersand, _, _ = strings.Cut(domainWithoutRoutes, "&")
		domainWithoutPercent, _, _ = strings.Cut(domainWithoutAmpersand, "%")
	}

	parser := parser.NewDomainParser()
	return Domain(parser.GetFQDN(string(domainWithoutPercent)))
}

func (d Domain) HasBeenRequestedTooRecently(
	politeness_interval time.Duration,
	queue *sql.DB,
	mu *sync.Mutex,
) (
	bool,
	error,
) {
	var (
		// timestapm in Unix time
		lastCrawled int64

		// better to be too
		// polite than not nice enough
		// to an API
		hasBeenRequestedTooRecently bool = true
	)

	mu.Lock()
	defer mu.Unlock()

	err := queue.QueryRow(
		`SELECT time_last_requested
		FROM domains
		WHERE fqdn = ?;`,
		d.GetFQDN(),
	).Scan(&lastCrawled)

	if err == sql.ErrNoRows {
		hasBeenRequestedTooRecently = false
		return hasBeenRequestedTooRecently, nil
	} else if err != nil {
		return hasBeenRequestedTooRecently, fmt.Errorf("execute stmt failed: %w", err)
	}

	hasBeenRequestedTooRecently = time.Since(
		time.Unix(lastCrawled, 0),
	) < politeness_interval

	return hasBeenRequestedTooRecently, nil
}

func (d Domain) IsBlacklisted(
	db *sql.DB,
	dbMu *sync.Mutex,
) (
	bool,
	error,
) {
	var (
		fqdn                    = d.GetFQDN()
		domainLevels   []string = strings.Split(string(fqdn), ".")
		topLevelDomain          = "*." + domainLevels[len(domainLevels)-1]
		exists         bool
	)

	err := db.QueryRow(
		`SELECT EXISTS (
			SELECT 1
			FROM domain_blacklist
			WHERE domain = $1
				OR domain_blacklist.domain = $2
		);`,
		fqdn,
		topLevelDomain,
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
