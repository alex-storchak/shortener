BEGIN;

CREATE TABLE IF NOT EXISTS url_storage (
    uuid         SERIAL PRIMARY KEY,
    short_id     VARCHAR(255) NOT NULL UNIQUE,
    original_url TEXT         NOT NULL,
    created_at   TIMESTAMP DEFAULT NOW()
);

COMMIT;