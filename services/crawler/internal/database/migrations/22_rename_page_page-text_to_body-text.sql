-- +goose Up
ALTER TABLE pages
RENAME COLUMN body_text TO page_text;
