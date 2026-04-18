-- 000010_firebase_auth_integration.up.sql

ALTER TABLE users ADD COLUMN auth_provider VARCHAR(20) DEFAULT 'local' NOT NULL;
ALTER TABLE users ADD COLUMN provider_id VARCHAR(255);
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;

CREATE UNIQUE INDEX idx_users_provider_provider_id ON users (auth_provider, provider_id) WHERE provider_id IS NOT NULL;
