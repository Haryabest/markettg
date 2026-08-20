-- Richer copy for existing products
UPDATE catalog.products SET name = '50 Stars', description = '50 Telegram Stars сразу на ваш аккаунт. Подходят для реакций, подарков и оплаты Premium.' WHERE slug = 'stars-50';
UPDATE catalog.products SET name = '100 Stars', description = '100 Telegram Stars с мгновенной доставкой. Выгоднее, чем покупать по 50.' WHERE slug = 'stars-100';
UPDATE catalog.products SET name = '250 Stars', description = '250 Telegram Stars — запас на реакции, подарки и мини-приложения.' WHERE slug = 'stars-250';
UPDATE catalog.products SET name = '500 Stars', description = '500 Telegram Stars пакетом. Доставка сразу после оплаты, без ожидания.' WHERE slug = 'stars-500';
UPDATE catalog.products SET name = '1000 Stars', description = '1000 Telegram Stars — максимальный пакет для активных пользователей и подарков.' WHERE slug = 'stars-1000';
UPDATE catalog.products SET name = 'Premium 1 месяц', description = 'Telegram Premium на 30 дней: без рекламы, увеличенные лимиты, эксклюзивные реакции и облако 4 ГБ.' WHERE slug = 'premium-1m';
UPDATE catalog.products SET name = 'Premium 3 месяца', description = 'Telegram Premium на 90 дней. Дешевле помесячной оплаты, все привилегии Premium.' WHERE slug = 'premium-3m';
UPDATE catalog.products SET name = 'Premium 12 месяцев', description = 'Telegram Premium на год. Максимальная выгода, все функции без ограничений.' WHERE slug = 'premium-12m';
UPDATE catalog.products SET name = 'Подарок Мишка', description = 'Анимированный подарок «Мишка» в Telegram. Можно отправить другу или оставить себе.' WHERE slug = 'gift-bear';
UPDATE catalog.products SET name = 'Подарок Роза', description = 'Анимированный подарок «Роза». Классический жест, доставляется в чат мгновенно.' WHERE slug = 'gift-rose';

UPDATE catalog.categories SET name = 'Telegram Stars', description = 'Пакеты Telegram Stars с мгновенным зачислением на аккаунт' WHERE slug = 'stars';
UPDATE catalog.categories SET name = 'Telegram Premium', description = 'Подписка Telegram Premium на 1, 3, 6 и 12 месяцев' WHERE slug = 'premium';
UPDATE catalog.categories SET name = 'Подарки', description = 'Анимированные Telegram Gifts — отправим в чат сразу после оплаты' WHERE slug = 'gifts';

INSERT INTO catalog.products (slug, name, description, price_kopecks, product_type, delivery_config, popularity_score) VALUES
    ('stars-25', '25 Stars', 'Стартовый пакет: 25 Telegram Stars для реакций и небольших подарков. Зачисление сразу после оплаты.', 4900, 'STARS', '{"type":"STARS","amount":25}', 85),
    ('stars-150', '150 Stars', '150 Telegram Stars — удобный запас на неделю. Дешевле, чем несколько мелких пакетов.', 27900, 'STARS', '{"type":"STARS","amount":150}', 140),
    ('stars-1500', '1500 Stars', '1500 Telegram Stars для подарков и активных чатов. Выгодный курс за звезду.', 234900, 'STARS', '{"type":"STARS","amount":1500}', 55),
    ('stars-2500', '2500 Stars', '2500 Telegram Stars крупным пакетом. Для каналов, розыгрышей и подарков друзьям.', 379900, 'STARS', '{"type":"STARS","amount":2500}', 40),
    ('stars-5000', '5000 Stars', '5000 Telegram Stars — максимальный объём. Лучшая цена за звезду в каталоге.', 699900, 'STARS', '{"type":"STARS","amount":5000}', 30),
    ('premium-6m', 'Premium 6 месяцев', 'Telegram Premium на полгода: без рекламы, 4 ГБ облака, перевод голосовых и эксклюзивные стикеры.', 149900, 'PREMIUM', '{"type":"PREMIUM","duration_days":180}', 110),
    ('gift-heart', 'Подарок Сердце', 'Анимированное сердце в Telegram. Можно отправить в любой чат сразу после покупки.', 1990, 'GIFT', '{"type":"GIFT","gift_id":"heart"}', 95),
    ('gift-rocket', 'Подарок Ракета', 'Ракета — яркий подарок для поздравлений и запуска канала.', 3990, 'GIFT', '{"type":"GIFT","gift_id":"rocket"}', 75),
    ('gift-trophy', 'Подарок Кубок', 'Кубок победителя. Подходит для розыгрышей, челленджей и поздравлений.', 5990, 'GIFT', '{"type":"GIFT","gift_id":"trophy"}', 60),
    ('gift-cake', 'Подарок Торт', 'Праздничный торт на день рождения. Доставляется в чат как Telegram Gift.', 3490, 'GIFT', '{"type":"GIFT","gift_id":"cake"}', 80),
    ('gift-champagne', 'Подарок Шампанское', 'Бутылка шампанского — для праздников, офферов и тёплых поздравлений.', 4490, 'GIFT', '{"type":"GIFT","gift_id":"champagne"}', 58),
    ('gift-gem', 'Подарок Кристалл', 'Редкий кристалл. Выглядит эффектно в профиле и в истории подарков.', 9990, 'GIFT', '{"type":"GIFT","gift_id":"gem"}', 50),
    ('gift-bouquet', 'Подарок Букет', 'Цветочный букет в Telegram. Универсальный подарок без повода и с поводом.', 2990, 'GIFT', '{"type":"GIFT","gift_id":"bouquet"}', 72),
    ('gift-ring', 'Подарок Кольцо', 'Кольцо — премиальный Gift. Можно закрепить в профиле получателя.', 12990, 'GIFT', '{"type":"GIFT","gift_id":"ring"}', 45)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price_kopecks = EXCLUDED.price_kopecks,
    delivery_config = EXCLUDED.delivery_config,
    popularity_score = EXCLUDED.popularity_score,
    is_active = TRUE,
    updated_at = NOW();

INSERT INTO catalog.product_categories (product_id, category_id)
SELECT p.id, c.id FROM catalog.products p
JOIN catalog.categories c ON
    (p.product_type = 'STARS' AND c.slug = 'stars')
    OR (p.product_type = 'PREMIUM' AND c.slug = 'premium')
    OR (p.product_type = 'GIFT' AND c.slug = 'gifts')
ON CONFLICT DO NOTHING;

INSERT INTO catalog.promotions (title, description, discount_type, discount_value, starts_at, ends_at, is_active)
SELECT 'Скидка новичкам', 'Промокод WELCOME10 даёт 10% на первый заказ Stars, Premium или подарков.', 'PERCENT', 10, NOW() - INTERVAL '1 day', NOW() + INTERVAL '365 days', TRUE
WHERE NOT EXISTS (SELECT 1 FROM catalog.promotions WHERE title = 'Скидка новичкам');
