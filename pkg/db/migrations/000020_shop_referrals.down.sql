DROP TABLE IF EXISTS shop_referrals;
DROP INDEX IF EXISTS idx_admins_referral_coupon_id;
ALTER TABLE admins DROP COLUMN IF EXISTS referral_coupon_id;
