-- Seed data for Go Marketplace

-- 1. Users
-- Password is 'password123' hashed with bcrypt (cost 10)
-- $2a$10$8K1p/a0dxv.X1jX8R8M9.eX7W9G5Z4U.f/ZpY0mH0n0q0r0s0t0u
-- Actually, let's use a simpler one if possible, but bcrypt is better.
-- For seeding, we can just use a pre-calculated hash.

INSERT INTO public.users (id, username, email, password, full_name, is_verified) VALUES
('00000000-0000-0000-0000-000000000001', 'verified_user', 'verified@example.com', '$2a$10$SggTvHmzKlwarEm46g4vq.vDpEJZOw6ddp8BCfqwRmxu4JcktHIGO', 'Verified User', true),
('00000000-0000-0000-0000-000000000002', 'unverified_user', 'unverified@example.com', '$2a$10$SggTvHmzKlwarEm46g4vq.vDpEJZOw6ddp8BCfqwRmxu4JcktHIGO', 'Unverified User', false),
('00000000-0000-0000-0000-000000000003', 'merchant_user', 'merchant@example.com', '$2a$10$SggTvHmzKlwarEm46g4vq.vDpEJZOw6ddp8BCfqwRmxu4JcktHIGO', 'Merchant User', true)
ON CONFLICT (email) DO NOTHING;

-- 2. Merchants
INSERT INTO public.merchants (id, user_id, name, about, tax_id) VALUES
('00000000-0000-0000-0000-000000000101', '00000000-0000-0000-0000-000000000003', 'Tech Gadgets Store', 'Your one-stop shop for latest tech gadgets', 'TAX-123456')
ON CONFLICT (id) DO NOTHING;

-- 3. Products
INSERT INTO public.products (id, store_id, name, description, price, stock, height_cm, width_cm, depth_cm, weight_kg, is_onsale) VALUES
('00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000101', 'Smartphone Pro', 'High-end smartphone with amazing camera', 999.99, 50, 15.0, 7.5, 0.8, 0.2, true),
('00000000-0000-0000-0000-000000000202', '00000000-0000-0000-0000-000000000101', 'Laptop Air', 'Lightweight laptop for productivity', 1299.50, 20, 30.0, 20.0, 1.5, 1.2, false),
('00000000-0000-0000-0000-000000000203', '00000000-0000-0000-0000-000000000101', 'Wireless Earbuds', 'Noise-canceling earbuds with long battery life', 149.00, 100, 5.0, 5.0, 2.0, 0.05, true)
ON CONFLICT (id) DO NOTHING;

-- 4. Wallets
INSERT INTO public.wallets (id, user_id, wallet_number, balance, pending_balance, currency, status) VALUES
('00000000-0000-0000-0000-000000000301', '00000000-0000-0000-0000-000000000001', 'WAL-VERIFIED-001', 5000.00, 0.00, 'IDR', 'active'),
('00000000-0000-0000-0000-000000000302', '00000000-0000-0000-0000-000000000002', 'WAL-UNVERIFIED-002', 100.00, 0.00, 'IDR', 'active'),
('00000000-0000-0000-0000-000000000303', '00000000-0000-0000-0000-000000000003', 'WAL-MERCHANT-003', 10000.00, 500.00, 'IDR', 'active')
ON CONFLICT (id) DO NOTHING;

-- 5. User Addresses
INSERT INTO public.user_addresses (id, user_id, tag, recipient_name, phone_number, street_address, city, province, postal_code, is_default) VALUES
('00000000-0000-0000-0000-000000000401', '00000000-0000-0000-0000-000000000001', 'Home', 'Verified User', '081234567890', '123 Verified St', 'Jakarta', 'DKI Jakarta', '12345', true),
('00000000-0000-0000-0000-000000000402', '00000000-0000-0000-0000-000000000003', 'Work', 'Merchant User', '089876543210', '456 Merchant Ave', 'Bandung', 'West Java', '40123', true)
ON CONFLICT (id) DO NOTHING;

-- 6. Payments (Escrow Example)
INSERT INTO public.payments (id, user_id, amount, payment_type, payment_method, status, reference_id) VALUES
('00000000-0000-0000-0000-000000000501', '00000000-0000-0000-0000-000000000001', 500.00, 'order', 'wallet', 'success', '00000000-0000-0000-0000-000000000601')
ON CONFLICT (id) DO NOTHING;

-- 7. Payment Distributions (Funds waiting in escrow for merchant)
INSERT INTO public.payment_distributions (id, payment_id, recipient_id, amount) VALUES
('00000000-0000-0000-0000-000000000701', '00000000-0000-0000-0000-000000000501', '00000000-0000-0000-0000-000000000003', 500.00)
ON CONFLICT (id) DO NOTHING;
