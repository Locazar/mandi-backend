-- Database Seeding Script for Localzar Mandi Backend
-- This script populates test data for development and manual testing

-- Clear existing test data (optional - comment out for production)
-- DELETE FROM "orders" WHERE id > 1000;
-- DELETE FROM "admins" WHERE id > 100;
-- DELETE FROM "shops" WHERE id > 100;

-- ============================================
-- 1. TEST ADMIN/SELLER ACCOUNTS
-- ============================================

INSERT INTO "admins" (
  phone, email, first_name, last_name, password_hash, is_verified, 
  created_at, updated_at
) VALUES
-- Test Account 1: Verified seller with complete profile
(
  '9876543210', 'seller1@test.com', 'Raj', 'Kumar', 
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36gZvWFm', 
  true, NOW(), NOW()
),
-- Test Account 2: Unverified seller (needs documents)
(
  '9876543211', 'seller2@test.com', 'Priya', 'Singh',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36gZvWFm',
  false, NOW(), NOW()
),
-- Test Account 3: New seller (just created)
(
  '9876543212', 'seller3@test.com', 'Arjun', 'Patel',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36gZvWFm',
  false, NOW(), NOW()
) ON CONFLICT DO NOTHING;

-- Get IDs for use in next inserts
CREATE TEMP TABLE admin_ids AS
SELECT id, phone FROM "admins" 
WHERE phone IN ('9876543210', '9876543211', '9876543212')
ORDER BY phone;

-- ============================================
-- 2. SHOPS (Shop Details)
-- ============================================

INSERT INTO "shops" (
  admin_id, shop_name, owner_name, email, phone, 
  shop_description, shop_type, address_line1, address_line2, 
  city, state, country, pincode, latitude, longitude,
  bank_account_number, bank_ifsc, pan_number,
  shop_status, shop_verification_status,
  photo_shop_verification, business_doc_verification, 
  identity_doc_verification, address_proof_verification,
  created_at, updated_at
) 
SELECT
  a.id, 
  CASE WHEN a.phone = '9876543210' THEN 'Fresh & Organic Vegetables'
       WHEN a.phone = '9876543211' THEN 'Spice Hub Delhi'
       ELSE 'Premium Dairy Products' END,
  a.phone,
  'shop@test.com',
  a.phone,
  'Premium quality produce and vegetables from local farms',
  'produce',
  '123 Market Street',
  'Near Railway Station',
  CASE WHEN a.phone = '9876543210' THEN 'Delhi'
       WHEN a.phone = '9876543211' THEN 'Mumbai'
       ELSE 'Bangalore' END,
  CASE WHEN a.phone = '9876543210' THEN 'Delhi'
       WHEN a.phone = '9876543211' THEN 'Maharashtra'
       ELSE 'Karnataka' END,
  'India',
  CASE WHEN a.phone = '9876543210' THEN '110001'
       WHEN a.phone = '9876543211' THEN '400001'
       ELSE '560001' END,
  CASE WHEN a.phone = '9876543210' THEN 28.7041
       WHEN a.phone = '9876543211' THEN 19.0760
       ELSE 12.9716 END,
  CASE WHEN a.phone = '9876543210' THEN 77.1025
       WHEN a.phone = '9876543211' THEN 72.8777
       ELSE 77.5946 END,
  '1234567890123456',
  'SBIN0001234',
  'ABCDE1234F',
  'active',
  CASE WHEN a.phone = '9876543210' THEN true ELSE false END,
  CASE WHEN a.phone = '9876543210' THEN true ELSE false END,
  CASE WHEN a.phone = '9876543210' THEN true ELSE false END,
  CASE WHEN a.phone = '9876543210' THEN true ELSE false END,
  CASE WHEN a.phone = '9876543210' THEN true ELSE false END,
  NOW(), NOW()
FROM admin_ids a
ON CONFLICT (admin_id) DO NOTHING;

-- ============================================
-- 3. CATEGORIES & SUB-CATEGORIES
-- ============================================

INSERT INTO "categories" (name, description, image_url, is_active, created_at, updated_at) VALUES
('Vegetables', 'Fresh vegetables and produce', 'vegetables.jpg', true, NOW(), NOW()),
('Fruits', 'Fresh fruits and berries', 'fruits.jpg', true, NOW(), NOW()),
('Spices', 'Herbs, spices and condiments', 'spices.jpg', true, NOW(), NOW()),
('Dairy', 'Milk, cheese and dairy products', 'dairy.jpg', true, NOW(), NOW()),
('Grains', 'Rice, wheat and pulses', 'grains.jpg', true, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ============================================
-- 4. PRODUCTS (Items)
-- ============================================

CREATE TEMP TABLE shop_id_temp AS
SELECT id FROM "shops" LIMIT 1;

INSERT INTO "product_items" (
  shop_id, category_id, name, description, 
  price, quantity, unit, 
  is_active, created_at, updated_at
)
SELECT
  (SELECT id FROM shop_id_temp LIMIT 1),
  (SELECT id FROM "categories" WHERE name = 'Vegetables' LIMIT 1),
  'Fresh Tomatoes',
  'Locally sourced premium red tomatoes',
  80,
  100,
  'kg',
  true,
  NOW(),
  NOW()
UNION ALL
SELECT
  (SELECT id FROM shop_id_temp LIMIT 1),
  (SELECT id FROM "categories" WHERE name = 'Vegetables' LIMIT 1),
  'Organic Spinach',
  'Fresh organic spinach bundles',
  60,
  50,
  'bundle',
  true,
  NOW(),
  NOW()
UNION ALL
SELECT
  (SELECT id FROM shop_id_temp LIMIT 1),
  (SELECT id FROM "categories" WHERE name = 'Fruits' LIMIT 1),
  'Bananas',
  'Premium yellow bananas',
  45,
  150,
  'kg',
  true,
  NOW(),
  NOW()
UNION ALL
SELECT
  (SELECT id FROM shop_id_temp LIMIT 1),
  (SELECT id FROM "categories" WHERE name = 'Spices' LIMIT 1),
  'Turmeric Powder',
  '100% pure turmeric powder',
  250,
  25,
  'kg',
  true,
  NOW(),
  NOW()
UNION ALL
SELECT
  (SELECT id FROM shop_id_temp LIMIT 1),
  (SELECT id FROM "categories" WHERE name = 'Dairy' LIMIT 1),
  'Whole Milk',
  'Fresh pasteurized whole milk',
  65,
  200,
  'liter',
  true,
  NOW(),
  NOW()
ON CONFLICT DO NOTHING;

-- ============================================
-- 5. ORDERS (for testing order flows)
-- ============================================

INSERT INTO "orders" (
  shop_id, customer_id, total_amount, order_status,
  created_at, updated_at
)
SELECT
  (SELECT id FROM shop_id_temp LIMIT 1),
  1,
  500,
  'pending',
  NOW() - INTERVAL '2 hours',
  NOW() - INTERVAL '2 hours'
UNION ALL
SELECT
  (SELECT id FROM shop_id_temp LIMIT 1),
  2,
  1200,
  'confirmed',
  NOW() - INTERVAL '1 day',
  NOW() - INTERVAL '1 day'
UNION ALL
SELECT
  (SELECT id FROM shop_id_temp LIMIT 1),
  3,
  750,
  'completed',
  NOW() - INTERVAL '2 days',
  NOW() - INTERVAL '2 days'
ON CONFLICT DO NOTHING;

-- ============================================
-- 6. VERIFY SEEDING
-- ============================================

SELECT '=== SEED DATA SUMMARY ===' as info;
SELECT COUNT(*) as admin_count FROM "admins" WHERE phone LIKE '9876543%';
SELECT COUNT(*) as shop_count FROM "shops";
SELECT COUNT(*) as product_count FROM "product_items";
SELECT COUNT(*) as order_count FROM "orders";
SELECT COUNT(*) as category_count FROM "categories";

-- Cleanup
DROP TABLE IF EXISTS admin_ids;
DROP TABLE IF EXISTS shop_id_temp;
