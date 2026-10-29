package connection

import (
	"database/sql"
	"sync"
)

const (
	QUEUE_SQLITE_CONNECTION_PATH = "/var/lib/ohana/crawler/queue.db"
)

var (
	DB *sql.DB

	Queue   *sql.DB
	QueueMu sync.Mutex
)
