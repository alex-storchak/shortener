BEGIN;

ALTER TABLE url_storage ADD COLUMN IF NOT EXISTS user_id INTEGER;

UPDATE url_storage
SET user_id = (
    SELECT id FROM auth_user WHERE user_uuid = '00000000-0000-0000-0000-000000000000'
);

CREATE INDEX IF NOT EXISTS idx_url_storage_user_id ON url_storage(user_id);

ALTER TABLE url_storage ADD CONSTRAINT fk_url_storage_user_id FOREIGN KEY (user_id) REFERENCES auth_user(id);

COMMIT;