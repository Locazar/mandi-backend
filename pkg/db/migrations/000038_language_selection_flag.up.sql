-- Gates the seller-app onboarding language picker
-- (FeatureFlagService.languageSelection). Ships DISABLED — an admin enables it
-- from admin-portal's Feature Flags page when ready.
INSERT INTO feature_flags (id, flag_key, enabled, description)
VALUES ('ff_seed_language_selection', 'language_selection', FALSE, 'Seller app: language selection step during onboarding')
ON CONFLICT (flag_key) DO NOTHING;
