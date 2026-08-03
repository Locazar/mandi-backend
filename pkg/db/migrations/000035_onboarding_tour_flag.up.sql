-- Gates the one-time coach-mark feature tour across the seller-app's KYC
-- onboarding steps (FeatureFlagService.onboardingTour). Ships DISABLED — an
-- admin enables it from admin-portal's Feature Flags page when ready.
INSERT INTO feature_flags (id, flag_key, enabled, description)
VALUES ('ff_seed_onboarding_tour', 'onboarding_tour', FALSE, 'Seller app: one-time coach-mark feature tour across the KYC onboarding steps')
ON CONFLICT (flag_key) DO NOTHING;
