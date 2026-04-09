CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price NUMERIC(15, 2) NOT NULL,
    stock INTEGER NOT NULL DEFAULT 0,
    height_cm NUMERIC(10, 2),
    width_cm NUMERIC(10, 2),
    depth_cm NUMERIC(10, 2),
    weight_kg NUMERIC(10, 2),
    is_onsale BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
