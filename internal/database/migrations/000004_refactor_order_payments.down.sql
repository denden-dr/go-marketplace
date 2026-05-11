-- Down Migration: Revert order payments refactor
-- Date: 2026-05-10

-- 1. Recreate the order_payments table
CREATE TABLE IF NOT EXISTS public.order_payments (
    id uuid NOT NULL PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    amount numeric(19,4) NOT NULL,
    payment_method character varying(50) NOT NULL,
    status character varying(50) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

-- 2. Drop the foreign key from orders to payments
ALTER TABLE public.orders DROP CONSTRAINT IF EXISTS orders_payment_id_fkey;

-- 3. Restore the old foreign key from orders to order_payments
ALTER TABLE public.orders ADD CONSTRAINT orders_payment_id_fkey FOREIGN KEY (payment_id) REFERENCES public.order_payments(id) ON DELETE CASCADE;
