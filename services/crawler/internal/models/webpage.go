package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/lib/pq"
)

type (
	Webpage struct {
		ResponseBody string `json:"response_body"`

		FullDomain        Domain    `json:"full_domain"`
		Fqdn              Domain    `json:"fqdn"`
		Url               Url       `json:"Url"`
		Title             string    `json:"title"`
		Description       string    `json:"description"`
		Text              string    `json:"text"`
		Outneighbours     []Url     `json:"outneighbours"`
		Date_discovered   time.Time `json:"date_discovered"`
		Date_last_crawled time.Time `json:"date_last_crawled"`
		IsFediverseNode   bool      `json:"is_site_fediverse"`
	}
)

func (page *Webpage) Save(db *sql.DB) error {

	if len(page.FullDomain) == 0 {
		return fmt.Errorf("domain empty")
	}

	if len(page.Url.TrimTrailingSlash()) == 0 {
		return fmt.Errorf("url empty")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("start tx failed: %w", err)
	}
	defer tx.Rollback()

	var (
		siteId           int64
		isSiteInDatabase bool = true
	)
	err = tx.QueryRow(
		`SELECT id 
		FROM sites
		WHERE full_domain = $1;`,
		page.FullDomain,
	).Scan(&siteId)
	if err == sql.ErrNoRows {
		isSiteInDatabase = false
	} else if err != nil {
		return fmt.Errorf("SELECT site id failed: %w", err)
	}

	var (
		pageId           int64
		isPageInDatabase bool = true
	)
	err = tx.QueryRow(
		`SELECT id 
		FROM pages
		WHERE link = $1;`,
		page.Url.TrimTrailingSlash(),
	).Scan(&pageId)
	if err == sql.ErrNoRows {
		isPageInDatabase = false
	} else if err != nil {
		return fmt.Errorf("SELECT page id failed: %w", err)
	}

	if !isSiteInDatabase {
		stmt, err := tx.Prepare(
			`INSERT INTO sites (
				fqdn,
				full_domain,
				date_discovered,
				date_last_crawled,
				is_fediverse
			)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id;`,
		)
		if err != nil {
			return fmt.Errorf("prepare INSERT site stmt failed: %w", err)
		}

		err = stmt.QueryRow(
			page.Fqdn,
			page.FullDomain,
			time.Now(),
			time.Now(),
			page.IsFediverseNode,
		).Scan(&siteId)
		if err != nil {
			return fmt.Errorf("execute INSERT site stmt failed: %w", err)
		}
	} else {
		_, err = tx.Exec(
			`UPDATE sites
			SET
				date_last_crawled = $1,
				is_fediverse = $2
			WHERE id = $3;`,
			time.Now(), page.IsFediverseNode, siteId)
		if err != nil {
			return fmt.Errorf("update site date_last_crawled failed: %w", err)
		}
	}

	if !isPageInDatabase {
		stmt, err := tx.Prepare(
			`INSERT INTO pages (
				site_id,
				title,
				description,
				link,
				page_text,
				response_body,
				date_discovered,
				date_last_crawled,
				outlinks
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`,
		)
		if err != nil {
			return fmt.Errorf("prepare INSERT page stmt failed: %w", err)
		}

		_, err = stmt.Exec(
			siteId,
			page.Title,
			page.Description,
			page.Url.TrimTrailingSlash(),
			page.Text,
			page.ResponseBody,
			time.Now(),
			time.Now(),
			pq.Array(page.Outneighbours),
		)
		if err != nil {
			return fmt.Errorf("execute INSERT page stmt failed: %w", err)
		}
	} else {
		_, err = tx.Exec(
			`UPDATE pages
			SET 
				date_last_crawled = $1,
				outlinks = $2,
			WHERE id = $4;`,
			time.Now(), pq.Array(page.Outneighbours), pageId)
		if err != nil {
			return fmt.Errorf("update page date_last_crawled failed: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit tx failed: %w", err)
	}

	return nil
}

func (p *Webpage) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &p)
}

func (p *Webpage) EnqueueToIndexer(db *sql.DB, mu *sync.Mutex) error {
	var (
		pageId int64
		siteId int64
	)

	mu.Lock()
	defer mu.Unlock()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback()

	err = tx.QueryRow(
		`SELECT id, site_id
		FROM pages
		WHERE link = $1;`, p.Url).Scan(&pageId, &siteId)
	if err != nil {
		return fmt.Errorf("SELECT page_id and site_id failed: %w", err)
	}

	_, err = tx.Exec(
		`INSERT INTO indexer_queue (
			page_id,
			site_id
		) VALUES (
			$1, $2
		);`, pageId, siteId)
	if err != nil {
		return fmt.Errorf("INSERT INTO indexer_queue failed: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit tx failed: %w", err)
	}

	return nil
}

func (p *Webpage) Value() (driver.Value, error) {
	return json.Marshal(p)
}
