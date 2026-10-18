package connection

import (
	"database/sql"
	"sync"
)

var (
	DB *sql.DB

	Queue   *sql.DB
	QueueMu sync.Mutex
)
