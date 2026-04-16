-- Backfill existing NULL data to 0
UPDATE products SET height_cm = 0 WHERE height_cm IS NULL;
UPDATE products SET width_cm = 0 WHERE width_cm IS NULL;
UPDATE products SET depth_cm = 0 WHERE depth_cm IS NULL;
UPDATE products SET weight_kg = 0 WHERE weight_kg IS NULL;

-- Set defaults
ALTER TABLE products ALTER COLUMN height_cm SET DEFAULT 0;
ALTER TABLE products ALTER COLUMN width_cm SET DEFAULT 0;
ALTER TABLE products ALTER COLUMN depth_cm SET DEFAULT 0;
ALTER TABLE products ALTER COLUMN weight_kg SET DEFAULT 0;

-- Add NOT NULL constraints
ALTER TABLE products ALTER COLUMN height_cm SET NOT NULL;
ALTER TABLE products ALTER COLUMN width_cm SET NOT NULL;
ALTER TABLE products ALTER COLUMN depth_cm SET NOT NULL;
ALTER TABLE products ALTER COLUMN weight_kg SET NOT NULL;
