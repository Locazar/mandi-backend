-- 000001_baseline.down.sql
-- Drop all tables in reverse dependency order (referencers before targets).

-- DPDP / audit
DROP TABLE IF EXISTS audit_logs                   CASCADE;
DROP TABLE IF EXISTS user_consents                CASCADE;

-- Auth sessions
DROP TABLE IF EXISTS otp_sessions_email           CASCADE;
DROP TABLE IF EXISTS otp_sessions                 CASCADE;
DROP TABLE IF EXISTS user_refresh_sessions        CASCADE;
DROP TABLE IF EXISTS admin_refresh_sessions       CASCADE;

-- Mobile auth
DROP TABLE IF EXISTS login_audit_logs             CASCADE;
DROP TABLE IF EXISTS otp_requests                 CASCADE;
DROP TABLE IF EXISTS mobile_users                 CASCADE;

-- Jobs
DROP TABLE IF EXISTS jobs                         CASCADE;
DROP TABLE IF EXISTS job_category_locations       CASCADE;
DROP TABLE IF EXISTS job_category_filters         CASCADE;
DROP TABLE IF EXISTS job_filters                  CASCADE;
DROP TABLE IF EXISTS job_locations                CASCADE;
DROP TABLE IF EXISTS job_sub_categories           CASCADE;
DROP TABLE IF EXISTS job_categories               CASCADE;

-- FCM / notifications
DROP TABLE IF EXISTS seller_alert_logs            CASCADE;
DROP TABLE IF EXISTS alert_templates              CASCADE;
DROP TABLE IF EXISTS alert_actions                CASCADE;
DROP TABLE IF EXISTS alerts                       CASCADE;
DROP TABLE IF EXISTS fcm_tokens                   CASCADE;
DROP TABLE IF EXISTS notification_device_tokens   CASCADE;
DROP TABLE IF EXISTS notifications                CASCADE;

-- Banners / ads
DROP TABLE IF EXISTS banners                      CASCADE;
DROP TABLE IF EXISTS advertisements               CASCADE;

-- Shop social / verification
DROP TABLE IF EXISTS shop_verification_histories  CASCADE;
DROP TABLE IF EXISTS shop_verifications           CASCADE;
DROP TABLE IF EXISTS shop_socials                 CASCADE;
DROP TABLE IF EXISTS shop_times                   CASCADE;
DROP TABLE IF EXISTS shop_departments             CASCADE;
DROP TABLE IF EXISTS shop_offers                  CASCADE;

-- Orders / returns
DROP TABLE IF EXISTS order_returns                CASCADE;
DROP TABLE IF EXISTS order_lines                  CASCADE;
DROP TABLE IF EXISTS shop_orders                  CASCADE;

-- Cart / wallet / transactions
DROP TABLE IF EXISTS cart_items                   CASCADE;
DROP TABLE IF EXISTS carts                        CASCADE;
DROP TABLE IF EXISTS transactions                 CASCADE;
DROP TABLE IF EXISTS wallets                      CASCADE;

-- Coupon uses before coupons
DROP TABLE IF EXISTS coupon_uses                  CASCADE;
DROP TABLE IF EXISTS coupons                      CASCADE;

-- Wish lists
DROP TABLE IF EXISTS wish_lists                   CASCADE;

-- Subscriptions
DROP TABLE IF EXISTS user_subscriptions           CASCADE;
DROP TABLE IF EXISTS subscription_orders          CASCADE;
DROP TABLE IF EXISTS subscription_plans           CASCADE;

-- Promotions
DROP TABLE IF EXISTS promotions                   CASCADE;
DROP TABLE IF EXISTS promotions_types             CASCADE;
DROP TABLE IF EXISTS promotion_categories         CASCADE;

-- Offers
DROP TABLE IF EXISTS offer_products               CASCADE;
DROP TABLE IF EXISTS offer_categories             CASCADE;
DROP TABLE IF EXISTS offers                       CASCADE;

-- Product configs / images / views / filters
DROP TABLE IF EXISTS product_configurations       CASCADE;
DROP TABLE IF EXISTS product_item_images          CASCADE;
DROP TABLE IF EXISTS product_images               CASCADE;
DROP TABLE IF EXISTS product_item_views           CASCADE;
DROP TABLE IF EXISTS product_item_filter_types    CASCADE;
DROP TABLE IF EXISTS product_items                CASCADE;
DROP TABLE IF EXISTS products                     CASCADE;

-- Category images / attributes
DROP TABLE IF EXISTS sub_type_attribute_options   CASCADE;
DROP TABLE IF EXISTS sub_type_attributes          CASCADE;
DROP TABLE IF EXISTS category_images              CASCADE;

-- Variation options / variations
DROP TABLE IF EXISTS variation_options            CASCADE;
DROP TABLE IF EXISTS variations                   CASCADE;

-- Taxonomy
DROP TABLE IF EXISTS sub_categories               CASCADE;
DROP TABLE IF EXISTS categories                   CASCADE;
DROP TABLE IF EXISTS departments                  CASCADE;
DROP TABLE IF EXISTS brands                       CASCADE;

-- Shop details
DROP TABLE IF EXISTS shop_details                 CASCADE;

-- Payment methods
DROP TABLE IF EXISTS payment_methods              CASCADE;

-- User addresses / addresses
DROP TABLE IF EXISTS user_addresses               CASCADE;
DROP TABLE IF EXISTS addresses                    CASCADE;

-- Core entities
DROP TABLE IF EXISTS admins                       CASCADE;
DROP TABLE IF EXISTS users                        CASCADE;
DROP TABLE IF EXISTS countries                    CASCADE;
