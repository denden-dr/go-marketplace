CREATE EXTENSION IF NOT EXISTS pg_trgm;

ALTER TABLE products ADD COLUMN search_vector tsvector;

-- Create GIN indexes
CREATE INDEX idx_products_search_vector ON products USING GIN(search_vector);
CREATE INDEX idx_products_name_trgm ON products USING GIN(name gin_trgm_ops);

-- Backfill existing data
UPDATE products 
SET search_vector = to_tsvector('english', name || ' ' || COALESCE(description, ''));

-- Create trigger function
CREATE OR REPLACE FUNCTION products_search_vector_trigger() RETURNS trigger AS $$
BEGIN
  new.search_vector := to_tsvector('english', new.name || ' ' || COALESCE(new.description, ''));
  RETURN new;
END
$$ LANGUAGE plpgsql;

-- Create trigger
CREATE TRIGGER products_search_vector_update
BEFORE INSERT OR UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION products_search_vector_trigger();
