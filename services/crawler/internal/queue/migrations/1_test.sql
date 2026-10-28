-- +goose Up

CREATE TABLE IF NOT EXISTS domains(
        id INTEGER PRIMARY KEY,
        fqdn TEXT NOT NULL UNIQUE, -- fully-qualified domain name
        time_last_requested INT NOT NULL DEFAULT 0 -- timestamp in Unix time format
);

CREATE TABLE IF NOT EXISTS crawler_queue(
        id INTEGER PRIMARY KEY,
        date_added TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        domain_id INT NOT NULL REFERENCES domains(id), 
        hyperlink TEXT UNIQUE NOT NULL
);

PRAGMA foreign_keys = ON;

