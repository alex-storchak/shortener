BEGIN;

ALTER TABLE url_storage ADD COLUMN IF NOT EXISTS user_id INTEGER;

CREATE INDEX IF NOT EXISTS idx_url_storage_user_id ON url_storage(user_id);

ALTER TABLE url_storage ADD CONSTRAINT fk_url_storage_user_id FOREIGN KEY (user_id) REFERENCES auth_user(id);

DROP INDEX IF EXISTS public.idx_url_storage_original_url;

CREATE UNIQUE INDEX IF NOT EXISTS idx_url_storage_original_url_user_id ON url_storage (original_url, user_id);

COMMIT;