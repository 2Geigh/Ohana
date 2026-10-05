-- +goose Up
ALTER TABLE "indexer_queue" DROP CONSTRAINT IF EXISTS "indexer_queue_site_id_fkey";
ALTER TABLE "indexer_queue"
ADD CONSTRAINT "indexer_queue_site_id_fkey" FOREIGN KEY ("site_id") REFERENCES "sites" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
ALTER TABLE "ranking_engine_queue" DROP CONSTRAINT IF EXISTS "ranking_engine_queue_site_id_fkey";
ALTER TABLE "ranking_engine_queue"
ADD CONSTRAINT "ranking_engine_queue_site_id_fkey" FOREIGN KEY ("site_id") REFERENCES "sites" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- +goose Down
ALTER TABLE "indexer_queue" DROP CONSTRAINT IF EXISTS "indexer_queue_site_id_fkey";
ALTER TABLE "indexer_queue"
ADD CONSTRAINT "indexer_queue_site_id_fkey" FOREIGN KEY ("site_id") REFERENCES "sites" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
ALTER TABLE "ranking_engine_queue" DROP CONSTRAINT IF EXISTS "ranking_engine_queue_site_id_fkey";
ALTER TABLE "ranking_engine_queue"
ADD CONSTRAINT "ranking_engine_queue_site_id_fkey" FOREIGN KEY ("site_id") REFERENCES "sites" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;