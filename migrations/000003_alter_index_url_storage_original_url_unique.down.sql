DROP INDEX IF EXISTS idx_url_storage_original_url;
CREATE INDEX idx_url_storage_original_url ON url_storage(original_url);