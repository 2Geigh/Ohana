-- +goose Up
ALTER TABLE pages
ADD COLUMN outlinks TEXT[] NOT NULL;

ALTER TABLE pages
ADD COLUMN backlinks TEXT[];