-- +goose Up
ALTER TABLE keywords
DROP COLUMN links;

ALTER TABLE keywords
ADD COLUMN page_id INTEGER NOT NULL;

ALTER TABLE keywords
ADD COLUMN word_occurences INTEGER NOT NULL;

ALTER TABLE keywords
ADD CONSTRAINT keywords_page_id_fkey FOREIGN KEY (page_id) REFERENCES pages (id);