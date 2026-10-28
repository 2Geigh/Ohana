package queue

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/2Geigh/Ohana/crawler/internal/connection"
	"github.com/2Geigh/Ohana/db-init/pkg/models"
	"github.com/pressly/goose/v3"
)

var (
	//go:embed migrations/*.sql
	embedMigrations embed.FS
)

func Dequeue(queue *sql.DB, mu *sync.Mutex) (models.Url, error) {
	var (
		id   int64
		link models.Url
	)

	mu.Lock()
	defer mu.Unlock()

	tx, err := queue.Begin()
	if err != nil {
		return link, fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback()

	err = tx.QueryRow(
		`SELECT id, hyperlink
		FROM crawler_queue
		ORDER BY date_added ASC
		LIMIT 1;`,
	).Scan(&id, &link)
	if err != nil {
		return link, fmt.Errorf("SELECT (id, hyperlink) FROM crawler_queue failed: %w", err)
	}

	_, err = tx.Exec(
		`DELETE FROM crawler_queue
		WHERE id = ?;`,
		id,
	)
	if err != nil {
		return link, fmt.Errorf("DELETE FROM crawler_queue failed: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return link, fmt.Errorf("commit tx failed: %w", err)
	}

	return link, nil
}

func UpdateTimeLastRequested(url models.Url, queue *sql.DB, mu *sync.Mutex, isUsingCheckpoints bool) error {
	mu.Lock()
	defer mu.Unlock()

	result, err := queue.Exec(
		`UPDATE domains
		SET time_last_requested = ?
		WHERE fqdn = ?;`,
		time.Now().Unix(),
		url.GetDomain().GetFQDN(),
	)

	if err != nil {
		return fmt.Errorf("UPDATE domains failed: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check updated rows: %w", err)
	}

	if isUsingCheckpoints {
		log.Printf("updated last-requested timestamp for %q: rows=%d",
			url.GetDomain().GetFQDN(), n)
	}

	return nil
}

func Enqueue(queue *sql.DB, mu *sync.Mutex, links []models.Url) error {
	mu.Lock()
	defer mu.Unlock()

	tx, err := queue.Begin()
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback()

	for _, link := range links {
		var (
			domainId int64
		)
		result, err := tx.Exec(
			`INSERT INTO domains
				(fqdn)
			VALUES
				(?)
			ON CONFLICT (fqdn)
				DO NOTHING;
			`,
			link.GetDomain().GetFQDN(),
		)
		if err != nil {
			return fmt.Errorf("INSERT %s INTO domains failed: %w", link.GetDomain().GetFQDN(), err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf(
				"get result.RowsAfffected of INSERT %s INTO domains failed: %w",
				link.GetDomain().GetFQDN(),
				err,
			)
		}

		if rowsAffected > 0 {
			domainId, err = result.LastInsertId()
			if err != nil {
				return fmt.Errorf(
					"get result.LastInsertId of INSERT %s INTO domains failed: %w",
					link.GetDomain().GetFQDN(),
					err,
				)
			}
		} else {
			err = tx.QueryRow(
				`SELECT id
				FROM domains
				WHERE fqdn = ?
				LIMIT 1;`,
				link.GetDomain().GetFQDN(),
			).Scan(&domainId)
			if err != nil {
				return fmt.Errorf(
					"SELECT %s's id from domains failed: %w",
					link.GetDomain().GetFQDN(),
					err,
				)
			}
		}

		_, err = tx.Exec(
			`INSERT INTO crawler_queue
				(domain_id, hyperlink)
			VALUES
				(?, ?)
			ON CONFLICT (hyperlink)
				DO NOTHING;`,
			domainId,
			link.SanitizeToEnqueue(),
		)
		if err != nil {
			return fmt.Errorf("INSERT %s INTO crawler_queue failed: %w", link.SanitizeToEnqueue(), err)
		}

	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit tx failed: %w", err)
	}

	return nil
}

func Migrate(queue *sql.DB) error {
	log.Println("Executing migrations...")

	goose.SetBaseFS(embedMigrations)

	err := goose.SetDialect("sqlite")
	if err != nil {
		return fmt.Errorf("goose: set database dialect failed: %w", err)
	}

	err = goose.Up(connection.Queue, "migrations")
	if err != nil {
		return fmt.Errorf("goose: apply migrations failed: %w", err)
	}

	return nil
}
