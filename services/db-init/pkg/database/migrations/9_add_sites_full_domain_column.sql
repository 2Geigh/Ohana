-- +goose Up
ALTER TABLE sites
ADD COLUMN full_domain TEXT NOT NULL;
