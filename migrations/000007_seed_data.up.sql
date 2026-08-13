-- Seed categories
INSERT INTO catalog.categories (slug, name, description, sort_order) VALUES
    ('stars', 'Telegram Stars', 'Покупка Telegram Stars', 1),
    ('premium', 'Telegram Premium', 'Подписка Telegram Premium', 2),
    ('gifts', 'Подарки', 'Telegram Gifts', 3);

-- Seed products
INSERT INTO catalog.products (slug, name, description, price_kopecks, product_type, delivery_config, popularity_score) VALUES
    ('stars-50', '50 Stars', '50 Telegram Stars для вашего аккаунта', 9900, 'STARS', '{"type":"STARS","amount":50}', 100),
    ('stars-100', '100 Stars', '100 Telegram Stars для вашего аккаунта', 18900, 'STARS', '{"type":"STARS","amount":100}', 200),
    ('stars-250', '250 Stars', '250 Telegram Stars для вашего аккаунта', 44900, 'STARS', '{"type":"STARS","amount":250}', 150),
    ('stars-500', '500 Stars', '500 Telegram Stars для вашего аккаунта', 84900, 'STARS', '{"type":"STARS","amount":500}', 80),
    ('stars-1000', '1000 Stars', '1000 Telegram Stars для вашего аккаунта', 159900, 'STARS', '{"type":"STARS","amount":1000}', 60),
    ('premium-1m', 'Premium 1 месяц', 'Telegram Premium на 1 месяц', 29900, 'PREMIUM', '{"type":"PREMIUM","duration_days":30}', 180),
    ('premium-3m', 'Premium 3 месяца', 'Telegram Premium на 3 месяца', 79900, 'PREMIUM', '{"type":"PREMIUM","duration_days":90}', 120),
    ('premium-12m', 'Premium 12 месяцев', 'Telegram Premium на 12 месяцев', 249900, 'PREMIUM', '{"type":"PREMIUM","duration_days":365}', 90),
    ('gift-bear', 'Подарок Мишка', 'Telegram Gift — Мишка', 1500, 'GIFT', '{"type":"GIFT","gift_id":"bear"}', 70),
    ('gift-rose', 'Подарок Роза', 'Telegram Gift — Роза', 2500, 'GIFT', '{"type":"GIFT","gift_id":"rose"}', 65);

-- Link products to categories
INSERT INTO catalog.product_categories (product_id, category_id)
SELECT p.id, c.id FROM catalog.products p, catalog.categories c
WHERE (p.product_type = 'STARS' AND c.slug = 'stars')
   OR (p.product_type = 'PREMIUM' AND c.slug = 'premium')
   OR (p.product_type = 'GIFT' AND c.slug = 'gifts');

-- Seed promo code
INSERT INTO orders.promo_codes (code, discount_type, discount_value, max_uses) VALUES
    ('WELCOME10', 'PERCENT', 10, 1000);
