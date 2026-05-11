-- Migration: Refactor order payments to use the new unified payments table
-- Date: 2026-05-10

-- 1. Drop the old foreign key from orders to order_payments
ALTER TABLE public.orders DROP CONSTRAINT IF EXISTS orders_payment_id_fkey;

-- 2. Add the new foreign key from orders to payments
ALTER TABLE public.orders ADD CONSTRAINT orders_payment_id_fkey FOREIGN KEY (payment_id) REFERENCES public.payments(id) ON DELETE CASCADE;

-- 3. Drop the redundant order_payments table
DROP TABLE IF EXISTS public.order_payments;
