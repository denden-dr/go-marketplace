-- 000007_user_enrichment_and_addresses.up.sql

-- Add full_name to users
ALTER TABLE users ADD COLUMN full_name VARCHAR(255) NOT NULL DEFAULT '';

-- Add Address attribute to orders
ALTER TABLE orders ADD COLUMN shipping_recipient_name VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE orders ADD COLUMN shipping_phone_number VARCHAR(20) NOT NULL DEFAULT '';
ALTER TABLE orders ADD COLUMN shipping_street_address TEXT NOT NULL DEFAULT '';
ALTER TABLE orders ADD COLUMN shipping_city VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE orders ADD COLUMN shipping_province VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE orders ADD COLUMN shipping_postal_code VARCHAR(10) NOT NULL DEFAULT '';

-- Create user_addresses table
CREATE TABLE IF NOT EXISTS user_addresses (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tag VARCHAR(50) NOT NULL,
    recipient_name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    street_address TEXT NOT NULL,
    city VARCHAR(100) NOT NULL,
    province VARCHAR(100) NOT NULL,
    postal_code VARCHAR(10) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster lookup by user
CREATE INDEX IF NOT EXISTS idx_user_addresses_user_id ON user_addresses(user_id);
