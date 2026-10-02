-- +goose Up
ALTER TABLE sites
ADD COLUMN is_indie_reference_example BOOLEAN NOT NULL DEFAULT FALSE;
