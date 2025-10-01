BEGIN;

DROP INDEX IF EXISTS idx_url_storage_original_url_user_id;

CREATE UNIQUE INDEX IF NOT EXISTS idx_url_storage_original_url ON public.url_storage (original_url);

-- https://github.com/Yandex-Practicum/go-autotests/pull/89/files
-- ALTER TABLE url_storage DROP CONSTRAINT IF EXISTS fk_url_storage_user_id;

DROP INDEX IF EXISTS idx_url_storage_user_id;

ALTER TABLE url_storage DROP COLUMN IF EXISTS user_id;

COMMIT;