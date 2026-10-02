-- +goose Up
ALTER TABLE link_queue
ADD COLUMN second_and_top_level_domain TEXT NOT NULL;
