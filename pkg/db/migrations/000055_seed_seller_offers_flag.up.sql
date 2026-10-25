INSERT INTO feature_flags (id, flag_key, enabled, description, created_at, updated_at)
VALUES (
    'ff_seller_offers_default01',
    'seller_offers',
    false,
    'Show the Offers & Promotions section in the seller app. Enable to give sellers access to the full offers/promotions feature.',
    NOW(),
    NOW()
)
ON CONFLICT (flag_key) DO NOTHING;
