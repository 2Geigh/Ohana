-- +goose Up
CREATE TABLE IF NOT EXISTS crawler_queue(
        id INTEGER PRIMARY KEY,
        hyperlink TEXT UNIQUE NOT NULL,
        date_added TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        fqdn TEXT NOT NULL
);
