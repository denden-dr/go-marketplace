-- Rename refresh_tokens to sessions and add metadata
ALTER TABLE refresh_tokens RENAME TO sessions;
ALTER TABLE sessions ADD COLUMN ip_address VARCHAR(45);
ALTER TABLE sessions ADD COLUMN user_agent TEXT;

-- Add is_verified to users
ALTER TABLE users ADD COLUMN is_verified BOOLEAN DEFAULT FALSE;

-- Create verification_codes table
CREATE TABLE IF NOT EXISTS verification_codes (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
