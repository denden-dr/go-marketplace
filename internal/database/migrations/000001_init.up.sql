-- PostgreSQL database dump
-- Dumped from database version 18.3
-- Dumped by pg_dump version 18.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

-- Extension
CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;

-- Functions
CREATE OR REPLACE FUNCTION public.products_search_vector_trigger() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  new.search_vector := to_tsvector('english', new.name || ' ' || COALESCE(new.description, ''));
  RETURN new;
END
$$;

SET default_tablespace = '';
SET default_table_access_method = heap;

-- Tables
CREATE TABLE IF NOT EXISTS public.users (
    id uuid NOT NULL PRIMARY KEY,
    username character varying(255) NOT NULL,
    email character varying(255) NOT NULL UNIQUE,
    password character varying(255),
    full_name character varying(255) DEFAULT ''::character varying NOT NULL,
    auth_provider character varying(20) DEFAULT 'local'::character varying NOT NULL,
    provider_id character varying(255),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public.merchants (
    id uuid DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    name character varying(255) NOT NULL,
    about text,
    tax_id character varying(100),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public.products (
    id uuid DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    store_id uuid NOT NULL REFERENCES public.merchants(id) ON DELETE CASCADE,
    name character varying(255) NOT NULL,
    description text,
    price numeric(15,2) NOT NULL,
    stock integer DEFAULT 0 NOT NULL,
    height_cm numeric(10,2) DEFAULT 0 NOT NULL,
    width_cm numeric(10,2) DEFAULT 0 NOT NULL,
    depth_cm numeric(10,2) DEFAULT 0 NOT NULL,
    weight_kg numeric(10,2) DEFAULT 0 NOT NULL,
    is_onsale boolean DEFAULT false,
    search_vector tsvector,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public.wallets (
    id uuid NOT NULL PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES public.users(id),
    wallet_number character varying(100) NOT NULL UNIQUE,
    balance numeric(15,2) DEFAULT 0.00,
    currency character varying(10) DEFAULT 'IDR'::character varying,
    status character varying(50) DEFAULT 'active'::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public.wallets_transaction (
    id uuid NOT NULL PRIMARY KEY,
    wallet_id uuid NOT NULL REFERENCES public.wallets(id),
    amount numeric(15,2) NOT NULL,
    direction character varying(10) NOT NULL,
    type character varying(50) NOT NULL,
    status character varying(50) NOT NULL,
    reference_id character varying(255),
    balance_after numeric(15,2) NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public.order_payments (
    id uuid NOT NULL PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    amount numeric(19,4) NOT NULL,
    payment_method character varying(50) NOT NULL,
    status character varying(50) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public.orders (
    id uuid NOT NULL PRIMARY KEY,
    payment_id uuid NOT NULL REFERENCES public.order_payments(id) ON DELETE CASCADE,
    merchant_id uuid NOT NULL REFERENCES public.merchants(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    status character varying(50) NOT NULL,
    total_amount numeric(19,4) NOT NULL,
    is_appealed boolean DEFAULT false,
    shipping_recipient_name character varying(255) DEFAULT ''::character varying NOT NULL,
    shipping_phone_number character varying(20) DEFAULT ''::character varying NOT NULL,
    shipping_street_address text DEFAULT ''::text NOT NULL,
    shipping_city character varying(100) DEFAULT ''::character varying NOT NULL,
    shipping_province character varying(100) DEFAULT ''::character varying NOT NULL,
    shipping_postal_code character varying(10) DEFAULT ''::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public.order_items (
    id uuid NOT NULL PRIMARY KEY,
    order_id uuid NOT NULL REFERENCES public.orders(id) ON DELETE CASCADE,
    product_id uuid NOT NULL REFERENCES public.products(id) ON DELETE CASCADE,
    quantity integer NOT NULL,
    price numeric(19,4) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public.cart_items (
    id uuid NOT NULL PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    product_id uuid NOT NULL REFERENCES public.products(id) ON DELETE CASCADE,
    quantity integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, product_id)
);

CREATE TABLE IF NOT EXISTS public.refresh_tokens (
    id uuid NOT NULL PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    token_hash character varying(255) NOT NULL,
    family_id uuid NOT NULL,
    is_revoked boolean DEFAULT false,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public.user_addresses (
    id uuid NOT NULL PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    tag character varying(50) NOT NULL,
    recipient_name character varying(255) NOT NULL,
    phone_number character varying(20) NOT NULL,
    street_address text NOT NULL,
    city character varying(100) NOT NULL,
    province character varying(100) NOT NULL,
    postal_code character varying(10) NOT NULL,
    is_default boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public.cancellation_appeals (
    id uuid NOT NULL PRIMARY KEY,
    order_id uuid NOT NULL REFERENCES public.orders(id) ON DELETE CASCADE,
    reason text NOT NULL,
    status character varying(50) DEFAULT 'pending'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_users_provider_provider_id ON public.users (auth_provider, provider_id) WHERE (provider_id IS NOT NULL);
CREATE INDEX IF NOT EXISTS idx_cart_items_user_id ON public.cart_items (user_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON public.order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_orders_merchant_id ON public.orders (merchant_id);
CREATE INDEX IF NOT EXISTS idx_orders_payment_id ON public.orders (payment_id);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON public.orders (user_id);
CREATE INDEX IF NOT EXISTS idx_products_name_trgm ON public.products USING gin (name public.gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_products_search_vector ON public.products USING gin (search_vector);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_family_id ON public.refresh_tokens (family_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON public.refresh_tokens (token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON public.refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_user_addresses_user_id ON public.user_addresses (user_id);
CREATE INDEX IF NOT EXISTS idx_wallets_transaction_wallet_id ON public.wallets_transaction (wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON public.wallets (user_id);

-- Triggers
DROP TRIGGER IF EXISTS products_search_vector_update ON public.products;
CREATE TRIGGER products_search_vector_update BEFORE INSERT OR UPDATE ON public.products FOR EACH ROW EXECUTE FUNCTION public.products_search_vector_trigger();
