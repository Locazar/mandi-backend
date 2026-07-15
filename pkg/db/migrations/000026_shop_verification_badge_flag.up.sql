-- Gates the shop-verification status icon on the seller-app's Add Products
-- page app bar (FeatureFlagService.shopVerificationBadge). That icon was
-- previously always shown, so this ships enabled to preserve current
-- behavior — an admin can flip it off from admin-portal's Feature Flags page.
INSERT INTO feature_flags (id, flag_key, enabled, description)
VALUES ('ff_seed_shop_verification_badge', 'shop_verification_badge', TRUE, 'Seller app: shop verification status icon on the Add Products page app bar')
ON CONFLICT (flag_key) DO NOTHING;
