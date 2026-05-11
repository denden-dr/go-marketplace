-- Add pending_balance to wallets
ALTER TABLE public.wallets ADD COLUMN pending_balance numeric(15,2) DEFAULT 0.00;

-- Add pending_balance_after to wallets_transaction
ALTER TABLE public.wallets_transaction ADD COLUMN pending_balance_after numeric(15,2) DEFAULT 0.00;

-- Create payments table
CREATE TABLE IF NOT EXISTS public.payments (
    id uuid NOT NULL PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES public.users(id),
    external_id character varying(255),
    amount numeric(15,2) NOT NULL,
    payment_type character varying(50) NOT NULL,
    payment_method character varying(50) NOT NULL,
    status character varying(50) NOT NULL,
    reference_id uuid,
    snap_token character varying(255),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payments_user_id ON public.payments (user_id);
CREATE INDEX IF NOT EXISTS idx_payments_external_id ON public.payments (external_id);
CREATE INDEX IF NOT EXISTS idx_payments_reference_id ON public.payments (reference_id);

-- Create payment_distributions table
CREATE TABLE IF NOT EXISTS public.payment_distributions (
    id uuid NOT NULL PRIMARY KEY,
    payment_id uuid NOT NULL REFERENCES public.payments(id) ON DELETE CASCADE,
    recipient_id uuid NOT NULL REFERENCES public.users(id),
    amount numeric(15,2) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payment_distributions_payment_id ON public.payment_distributions (payment_id);
