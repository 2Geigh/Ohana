-- +goose Up
ALTER TABLE pages
ADD COLUMN text_language VARCHAR(2) NOT NULL DEFAULT 'un';
