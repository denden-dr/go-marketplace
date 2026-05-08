DROP TRIGGER IF EXISTS products_search_vector_update ON public.products;
DROP FUNCTION IF EXISTS public.products_search_vector_trigger();

DROP TABLE IF EXISTS public.cancellation_appeals CASCADE;
DROP TABLE IF EXISTS public.user_addresses CASCADE;
DROP TABLE IF EXISTS public.refresh_tokens CASCADE;
DROP TABLE IF EXISTS public.cart_items CASCADE;
DROP TABLE IF EXISTS public.order_items CASCADE;
DROP TABLE IF EXISTS public.orders CASCADE;
DROP TABLE IF EXISTS public.order_payments CASCADE;
DROP TABLE IF EXISTS public.wallets_transaction CASCADE;
DROP TABLE IF EXISTS public.wallets CASCADE;
DROP TABLE IF EXISTS public.products CASCADE;
DROP TABLE IF EXISTS public.merchants CASCADE;
DROP TABLE IF EXISTS public.users CASCADE;

DROP EXTENSION IF EXISTS pg_trgm;
