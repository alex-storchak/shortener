BEGIN;

ALTER TABLE url_storage DROP CONSTRAINT IF EXISTS fk_url_storage_user_id;

DROP INDEX IF EXISTS idx_url_storage_user_id;

ALTER TABLE url_storage DROP COLUMN IF EXISTS user_id;

COMMIT;