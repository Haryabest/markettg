ALTER TABLE catalog.products
    ADD COLUMN IF NOT EXISTS telegram_gift_id VARCHAR(128);

ALTER TABLE catalog.products
    DROP CONSTRAINT IF EXISTS products_telegram_gift_id_key;

ALTER TABLE catalog.products
    ADD CONSTRAINT products_telegram_gift_id_key UNIQUE (telegram_gift_id);
