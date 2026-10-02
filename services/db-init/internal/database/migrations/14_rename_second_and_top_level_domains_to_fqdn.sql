-- +goose Up
ALTER TABLE sites
RENAME COLUMN second_and_top_level_domain TO fqdn;

ALTER TABLE link_queue
RENAME COLUMN second_and_top_level_domain TO fqdn;