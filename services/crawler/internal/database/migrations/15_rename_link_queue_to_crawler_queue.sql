-- +goose Up
ALTER TABLE link_queue
RENAME TO crawler_queue;
