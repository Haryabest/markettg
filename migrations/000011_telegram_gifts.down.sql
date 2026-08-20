ALTER TABLE catalog.products DROP CONSTRAINT IF EXISTS products_telegram_gift_id_key;
ALTER TABLE catalog.products DROP COLUMN IF EXISTS telegram_gift_id;
