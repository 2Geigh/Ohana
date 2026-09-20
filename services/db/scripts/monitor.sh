#!/bin/sh

exec watch -n 0.75 '
psql -d ohana -c "
SELECT count(*) AS sites_count FROM sites;
SELECT count(*) AS pages_count FROM pages;
SELECT count(*) AS crawler_queue_count FROM crawler_queue;
SELECT count(*) AS indexer_queue_count FROM indexer_queue;
SELECT count(*) AS ranker_queue_count FROM ranking_engine_queue;
"
'