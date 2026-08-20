UPDATE catalog.products
SET is_active = FALSE, updated_at = NOW()
WHERE product_type IN ('GIFT', 'NFT')
  AND (telegram_gift_id IS NULL OR telegram_gift_id = '');
