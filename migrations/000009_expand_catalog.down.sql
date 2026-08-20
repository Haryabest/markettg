DELETE FROM catalog.promotions WHERE title = 'Скидка новичкам';

DELETE FROM catalog.products WHERE slug IN (
    'stars-25', 'stars-150', 'stars-1500', 'stars-2500', 'stars-5000',
    'premium-6m',
    'gift-heart', 'gift-rocket', 'gift-trophy', 'gift-cake',
    'gift-champagne', 'gift-gem', 'gift-bouquet', 'gift-ring'
);
