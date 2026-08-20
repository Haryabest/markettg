-- NFT product type and cleanup fake seed gifts
ALTER TYPE catalog.product_type ADD VALUE IF NOT EXISTS 'NFT';

INSERT INTO catalog.categories (slug, name, description, sort_order, is_active)
VALUES (
    'nft',
    'NFT',
    'Коллекционные Telegram Gifts с уникальным номером',
    4,
    TRUE
)
ON CONFLICT (slug) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description,
    sort_order = EXCLUDED.sort_order,
    is_active = TRUE,
    updated_at = NOW();

UPDATE catalog.products
SET is_active = FALSE, updated_at = NOW()
WHERE product_type = 'GIFT'
  AND (telegram_gift_id IS NULL OR telegram_gift_id = '');
