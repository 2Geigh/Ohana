package models

import (
	"database/sql"
	"fmt"
	"sync"
)

type (
	LocalQueue struct {
		Mu    sync.Mutex
		Links []Url
	}
)

func (q *LocalQueue) Dequeue() Url {

	if len(q.Links) == 0 {
		return Url("")
	}

	toReturn := q.Links[0]

	if len(q.Links) == 1 {
		q.Links = []Url{}
	} else {
		q.Links = q.Links[1:]
	}

	return toReturn
}

func (q *LocalQueue) Enqueue(urls []Url, db *sql.DB) error {
	q.Mu.Lock()
	defer q.Mu.Unlock()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	for _, url := range urls {
		stmt, err := tx.Prepare(
			`INSERT INTO crawler_queue (hyperlink) VALUES ($1);`,
		)
		if err != nil {
			return fmt.Errorf("prepare statement failed: %w", err)
		}

		_, err = stmt.Exec(url)
		if err != nil {
			return fmt.Errorf("execute statement failed: %w", err)
		}

		err = stmt.Close()
		if err != nil {
			return fmt.Errorf("close statement failed: %w", err)
		}
	}

	// SAVE URL TO QUEUE TABLE IF NOT ALREADY PRESENT

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}

	return nil
}
