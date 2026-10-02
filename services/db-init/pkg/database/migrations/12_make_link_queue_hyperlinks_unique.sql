-- +goose Up
DELETE FROM link_queue;

ALTER TABLE link_queue
ADD CONSTRAINT link_queue_hyperlink_unique UNIQUE (hyperlink);

ALTER TABLE link_queue
ALTER COLUMN hyperlink
SET NOT NULL;
