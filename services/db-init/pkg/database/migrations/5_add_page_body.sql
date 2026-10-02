-- +goose Up
ALTER TABLE pages
ADD COLUMN body_text TEXT;

ALTER TABLE pages
ADD COLUMN response_body TEXT;