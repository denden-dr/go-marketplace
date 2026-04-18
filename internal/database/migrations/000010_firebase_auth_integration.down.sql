-- 000010_firebase_auth_integration.down.sql

DROP INDEX IF EXISTS idx_users_provider_provider_id;
ALTER TABLE users DROP COLUMN IF EXISTS auth_provider;
ALTER TABLE users DROP COLUMN IF EXISTS provider_id;
ALTER TABLE users ALTER COLUMN password SET NOT NULL;
