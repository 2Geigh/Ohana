-- +goose Up
-- https://www.paradedb.com/docs/start/create-your-first-index
CREATE INDEX search_idx
ON pages USING paradedb(
        id,
        title,
        description,
        page_text,
        date_last_crawled,
        text_language
);
