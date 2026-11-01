-- +goose Up
DROP INDEX IF EXISTS search_idx;

CREATE INDEX search_idx
ON pages USING paradedb (
    id,
    (title::pdb.unicode_words('pnorms=true')),
    (description::pdb.unicode_words('pnorms=true')),
    (page_text::pdb.unicode_words('pnorms=true')),
    date_last_crawled,
    text_language
);

-- +goose Down
DROP INDEX IF EXISTS search_idx;

CREATE INDEX search_idx
ON pages USING paradedb (
    id,
    title,
    description,
    page_text,
    date_last_crawled,
    text_language
);

