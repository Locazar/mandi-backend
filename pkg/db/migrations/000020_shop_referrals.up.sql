-- Sales Executive referral program: a platform user (admin) can carry a
-- unique referral coupon code; sellers optionally enter it during onboarding
-- and, on a match, get attached to that executive via shop_referrals.

ALTER TABLE admins ADD COLUMN IF NOT EXISTS referral_coupon_id VARCHAR(50);

-- Partial unique index: only enforces uniqueness for admins that actually
-- carry a referral code, so the many admins with an empty/NULL value never
-- collide with each other.
CREATE UNIQUE INDEX IF NOT EXISTS idx_admins_referral_coupon_id
    ON admins (referral_coupon_id)
    WHERE referral_coupon_id IS NOT NULL AND referral_coupon_id <> '';

CREATE TABLE IF NOT EXISTS shop_referrals (
    id VARCHAR(32) PRIMARY KEY,
    referral_coupon_id VARCHAR(50) NOT NULL,
    platform_user_id VARCHAR(32) NOT NULL,
    shop_id VARCHAR(32) NOT NULL,
    seller_admin_id VARCHAR(32),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- One active attachment per shop.
CREATE UNIQUE INDEX IF NOT EXISTS idx_shop_referrals_shop_id
    ON shop_referrals (shop_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_shop_referrals_platform_user_id ON shop_referrals (platform_user_id);
CREATE INDEX IF NOT EXISTS idx_shop_referrals_referral_coupon_id ON shop_referrals (referral_coupon_id);
