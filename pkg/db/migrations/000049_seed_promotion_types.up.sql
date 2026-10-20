-- Add the type column if it doesn't exist yet (safe to run multiple times).
ALTER TABLE promotions_types ADD COLUMN IF NOT EXISTS type VARCHAR(100) NOT NULL DEFAULT '';

-- Seed default promotion categories.
-- IDs match the Flutter app's hard-coded fallback pill values so the
-- category-by-id query returns results even when the API categories list is
-- displayed from the fallback.
INSERT INTO promotion_categories (id, name, is_active, icon_path, created_at, updated_at)
VALUES
    ('price_based', 'Price-Based',  TRUE, 'discount',    NOW(), NOW()),
    ('value_add',   'Value-Add',    TRUE, 'card_giftcard', NOW(), NOW()),
    ('loyalty',     'Loyalty',      TRUE, 'favorite',    NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Seed default promotion types for each category.
-- icon_path values match the icon-name → IconData switch in the Flutter app
-- so the correct icon renders without an image URL.
INSERT INTO promotions_types (id, name, is_active, promotion_category_id, promotion_offer_details, icon_path, type, created_at, updated_at)
VALUES
    -- ── Price-Based ────────────────────────────────────────────────────────
    ('prmt_pct_discount',  'Percentage Discount', TRUE, 'price_based', '{"discount_rate":10}',   'percent',         'percentage',     NOW(), NOW()),
    ('prmt_fixed_amt',     'Fixed Amount Off',    TRUE, 'price_based', '{"discount_rate":50}',   'currency_rupee',  'fixed_amount',   NOW(), NOW()),
    ('prmt_flash_sale',    'Flash Sale',          TRUE, 'price_based', '{"discount_rate":20}',   'flash_on',        'flash_sale',     NOW(), NOW()),
    ('prmt_bundle_deal',   'Bundle Deal',         TRUE, 'price_based', '{"discount_rate":15}',   'layers',          'bundle',         NOW(), NOW()),
    -- ── Value-Add ─────────────────────────────────────────────────────────
    ('prmt_bogo',          'Buy 1 Get 1',         TRUE, 'value_add',   '{"bogo_buy_quantity":1,"bogo_get_quantity":1}', 'redeem', 'bogo', NOW(), NOW()),
    ('prmt_free_gift',     'Free Gift',           TRUE, 'value_add',   '{}',                     'celebration',     'free_gift',      NOW(), NOW()),
    ('prmt_free_shipping', 'Free Shipping',       TRUE, 'value_add',   '{}',                     'local_shipping',  'free_shipping',  NOW(), NOW()),
    -- ── Loyalty ───────────────────────────────────────────────────────────
    ('prmt_new_customer',  'New Customer',        TRUE, 'loyalty',     '{"discount_rate":10}',   'waving_hand',     'new_customer',   NOW(), NOW()),
    ('prmt_repeat',        'Repeat Customer',     TRUE, 'loyalty',     '{"discount_rate":5}',    'stars',           'repeat',         NOW(), NOW()),
    ('prmt_referral',      'Referral Bonus',      TRUE, 'loyalty',     '{"discount_rate":10}',   'people',          'referral',       NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
