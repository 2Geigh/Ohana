package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type (
	Url string
)

func (url Url) GetDomain() Domain {
	var (
		domain string = string(url)
	)

	/*

		>>> getDomain(https://www.youtube.com/watch?v=dQw4w9WgXcQ)
		youtube.com

	*/

	if len(url) == 0 {
		return Domain(url)
	}

	_, domainWithoutProtocol, containsProtocol := strings.Cut(string(url), "://")
	if containsProtocol {
		domain = domainWithoutProtocol
	}

	_, domainWithoutWWW, containsWWW := strings.Cut(domain, "www.")
	if containsWWW {
		domain = domainWithoutWWW
	}

	domainWithoutRoutes, _, containsRoutes := strings.Cut(domain, "/")
	if containsRoutes {
		domain = domainWithoutRoutes
	}

	// We run this twice
	// To handle ? and # appearing
	// In whatever order
	for range 2 {
		domainWithoutQuery, _, containsQuery := strings.Cut(domain, "?")
		if containsQuery {
			domain = domainWithoutQuery
		}

		domainWithoutFragment, _, containsFragment := strings.Cut(domain, "#")
		if containsFragment {
			domain = domainWithoutFragment
		}
	}

	return Domain(domain)
}

func (url Url) IsTooRecentlyCrawled(
	db *sql.DB,
	oldness_threshold time.Duration,
) (
	bool,
	error,
) {
	var (
		lastCrawled          time.Time
		isTooRecentlyCrawled bool = false
	)

	stmt, err := db.Prepare(
		`SELECT date_last_crawled
		FROM pages
		WHERE link = $1;`,
	)
	if err != nil {
		return isTooRecentlyCrawled, fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(url.TrimTrailingSlash()).Scan(&lastCrawled)
	if err != nil && err != sql.ErrNoRows {
		return isTooRecentlyCrawled, fmt.Errorf("execute stmt failed: %w", err)
	}

	if time.Since(lastCrawled) < oldness_threshold {
		isTooRecentlyCrawled = true
	}

	return isTooRecentlyCrawled, nil
}

func (url Url) Sanitize() Url {
	return url.TrimTrailingSlash().TrimWWW()
}

func (url Url) TrimTrailingSlash() Url {
	return Url(
		strings.TrimSuffix(string(url), "/"),
	)
}

func (url Url) TrimWWW() Url {
	protocol, urlWithoutProtocol, hasProtocol := strings.Cut(string(url), "://")

	if hasProtocol {
		return Url(
			fmt.Sprintf(
				"%s://%s",
				protocol,
				strings.TrimPrefix(string(urlWithoutProtocol), "www.")),
		)
	}

	return Url(
		strings.TrimPrefix(string(url), "www."),
	)
}
