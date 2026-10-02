-- +goose Up
ALTER TABLE pages
ADD COLUMN link TEXT NOT NULL;
