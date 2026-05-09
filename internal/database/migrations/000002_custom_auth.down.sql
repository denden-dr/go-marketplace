-- Drop verification_codes table
DROP TABLE IF EXISTS verification_codes;

-- Remove is_verified from users
ALTER TABLE users DROP COLUMN IF EXISTS is_verified;

-- Remove metadata from sessions and rename back to refresh_tokens
ALTER TABLE sessions DROP COLUMN IF EXISTS ip_address;
ALTER TABLE sessions DROP COLUMN IF EXISTS user_agent;
ALTER TABLE sessions RENAME TO refresh_tokens;
