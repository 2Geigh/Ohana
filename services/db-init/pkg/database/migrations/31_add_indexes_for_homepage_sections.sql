-- +goose NO TRANSACTION

-- +goose Up
CREATE INDEX CONCURRENTLY IF NOT EXISTS
    pages_recently_added_homepage_idx
ON pages (date_discovered DESC)
WHERE title IS NOT NULL
  AND TRIM(title) <> ''
  AND description IS NOT NULL
  AND TRIM(description) <> '';

CREATE INDEX CONCURRENTLY IF NOT EXISTS
    pages_most_visited_homepage_idx
ON pages (visits DESC, date_discovered ASC);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS pages_recently_added_homepage_idx;
DROP INDEX CONCURRENTLY IF EXISTS pages_most_visited_homepage_idx;

