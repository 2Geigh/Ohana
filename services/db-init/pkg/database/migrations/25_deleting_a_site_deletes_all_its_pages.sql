-- +goose Up
ALTER TABLE "pages" DROP CONSTRAINT IF EXISTS "pages_site_id_fkey";
ALTER TABLE "pages"
ADD CONSTRAINT "pages_site_id_fkey" FOREIGN KEY ("site_id") REFERENCES "sites" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- +goose Down
ALTER TABLE "pages" DROP CONSTRAINT IF EXISTS "pages_site_id_fkey";
ALTER TABLE "pages"
ADD CONSTRAINT "pages_site_id_fkey" FOREIGN KEY ("site_id") REFERENCES "sites" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;