-- +goose Up
ALTER TABLE pages
ADD COLUMN date_last_indexed TIMESTAMP DEFAULT NULL;
