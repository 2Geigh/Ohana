-- +goose Up

ALTER TABLE pages
ADD visits INTEGER DEFAULT 0;

UPDATE pages
SET visits = 0;

ALTER TABLE pages
ALTER COLUMN visits SET NOT NULL;
