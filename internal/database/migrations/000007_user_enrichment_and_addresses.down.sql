-- 000007_user_enrichment_and_addresses.down.sql

DROP INDEX IF EXISTS idx_user_addresses_user_id;
DROP TABLE IF EXISTS user_addresses;
ALTER TABLE orders DROP COLUMN IF EXISTS shipping_recipient_name;
ALTER TABLE orders DROP COLUMN IF EXISTS shipping_phone_number;
ALTER TABLE orders DROP COLUMN IF EXISTS shipping_street_address;
ALTER TABLE orders DROP COLUMN IF EXISTS shipping_city;
ALTER TABLE orders DROP COLUMN IF EXISTS shipping_province;
ALTER TABLE orders DROP COLUMN IF EXISTS shipping_postal_code;
ALTER TABLE users DROP COLUMN IF EXISTS full_name;
