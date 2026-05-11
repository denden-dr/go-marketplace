-- Drop payment_distributions table
DROP TABLE IF EXISTS public.payment_distributions;

-- Drop payments table
DROP TABLE IF EXISTS public.payments;

-- Remove pending_balance_after from wallets_transaction
ALTER TABLE public.wallets_transaction DROP COLUMN IF EXISTS pending_balance_after;

-- Remove pending_balance from wallets
ALTER TABLE public.wallets DROP COLUMN IF EXISTS pending_balance;
