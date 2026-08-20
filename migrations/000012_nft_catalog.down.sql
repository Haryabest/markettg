UPDATE catalog.products
SET is_active = TRUE, updated_at = NOW()
WHERE product_type = 'GIFT'
  AND slug IN (
    'gift-bear', 'gift-rose', 'gift-heart', 'gift-rocket', 'gift-trophy',
    'gift-cake', 'gift-champagne', 'gift-gem', 'gift-bouquet', 'gift-ring'
  )
  AND telegram_gift_id IS NULL;

DELETE FROM catalog.categories WHERE slug = 'nft';

-- PostgreSQL does not support removing enum values safely.
