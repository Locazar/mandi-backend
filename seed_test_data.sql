-- Database Seeding Script for Localzar Mandi Backend
-- Updated for current schema (2026-05-25)
-- Focuses on sellers and shops (core entities)

-- ============================================
-- 1. TEST SELLER ACCOUNTS (admins)
-- ============================================

INSERT INTO admins (
  full_name, email, mobile, password, user_name,
  city, state, country, pincode,
  bank_account_number, bank_ifsc, pan, aadhar,
  verified_seller, status, agree_to_terms,
  latitude, longitude,
  created_at, updated_at
) VALUES
-- Seller 1: Verified with complete profile
(
  'Raj Kumar',
  'seller1@test.com',
  '9876543210',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36gZvWFm', -- password: password
  'seller_raj_001',
  'Delhi',
  'Delhi',
  'India',
  '110001',
  '1234567890123456',
  'SBIN0001234',
  'ABCDE1234F',
  'AAAA1234B567',
  true,
  'active',
  true,
  28.7041,
  77.1025,
  NOW(),
  NOW()
),
-- Seller 2: Verified, different city
(
  'Priya Singh',
  'seller2@test.com',
  '9876543211',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36gZvWFm',
  'seller_priya_002',
  'Mumbai',
  'Maharashtra',
  'India',
  '400001',
  '9876543210123456',
  'HDFC0001234',
  'BCDEF5678G',
  'BBBB5678C890',
  true,
  'active',
  true,
  19.0760,
  72.8777,
  NOW(),
  NOW()
),
-- Seller 3: Unverified (for verification testing)
(
  'Arjun Patel',
  'seller3@test.com',
  '9876543212',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36gZvWFm',
  'seller_arjun_003',
  'Bangalore',
  'Karnataka',
  'India',
  '560001',
  '5678901234567890',
  'ICIC0001234',
  'CDEFG9012H',
  'CCCC9012D345',
  false,
  'pending',
  false,
  12.9716,
  77.5946,
  NOW(),
  NOW()
),
-- Seller 4: Active seller
(
  'Neha Sharma',
  'seller4@test.com',
  '9876543213',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36gZvWFm',
  'seller_neha_004',
  'Hyderabad',
  'Telangana',
  'India',
  '500001',
  '2345678901234567',
  'AXIS0001234',
  'DEFGH1234I',
  'DDDD1234E567',
  true,
  'active',
  true,
  17.3850,
  78.4867,
  NOW(),
  NOW()
) ON CONFLICT (mobile) DO NOTHING;

-- ============================================
-- 2. SHOP DETAILS
-- ============================================

INSERT INTO shop_details (
  admin_id, shop_name, owner_name, email, phone,
  shop_description, shop_type,
  address_line1, address_line2, city, state, country, pincode,
  latitude, longitude,
  bank_account_number, bank_ifsc, pan_number,
  shop_status, shop_verification_status,
  photo_shop_verification, business_doc_verification,
  identity_doc_verification, address_proof_verification,
  created_at, updated_at
) VALUES
-- Shop 1: Fresh & Organic Vegetables
(
  (SELECT id FROM admins WHERE mobile = '9876543210' AND full_name = 'Raj Kumar' LIMIT 1),
  'Fresh & Organic Vegetables',
  'Raj Kumar',
  'shop1@test.com',
  '9876543210',
  'Premium quality organic vegetables directly from local farms. Specializing in seasonal produce with guaranteed freshness.',
  'produce',
  '123 Market Street',
  'Near Railway Station',
  'Delhi',
  'Delhi',
  'India',
  '110001',
  28.7041,
  77.1025,
  '1234567890123456',
  'SBIN0001234',
  'ABCDE1234F',
  'active',
  true,
  true,
  true,
  true,
  true,
  NOW(),
  NOW()
),
-- Shop 2: Spice Hub
(
  (SELECT id FROM admins WHERE mobile = '9876543211' AND full_name = 'Priya Singh' LIMIT 1),
  'Spice Hub Delhi',
  'Priya Singh',
  'shop2@test.com',
  '9876543211',
  'Authentic Indian spices, herbs and condiments. Pure, freshly ground spices with no additives.',
  'spices',
  '456 Spice Market',
  'Central Delhi',
  'Mumbai',
  'Maharashtra',
  'India',
  '400001',
  19.0760,
  72.8777,
  '9876543210123456',
  'HDFC0001234',
  'BCDEF5678G',
  'active',
  true,
  true,
  true,
  true,
  true,
  NOW(),
  NOW()
),
-- Shop 3: Premium Dairy
(
  (SELECT id FROM admins WHERE mobile = '9876543212' AND full_name = 'Arjun Patel' LIMIT 1),
  'Premium Dairy Products',
  'Arjun Patel',
  'shop3@test.com',
  '9876543212',
  'Fresh milk, cheese, yogurt and dairy products. Farm to table in 24 hours.',
  'dairy',
  '789 Dairy Lane',
  'Near Market',
  'Bangalore',
  'Karnataka',
  'India',
  '560001',
  12.9716,
  77.5946,
  '5678901234567890',
  'ICIC0001234',
  'CDEFG9012H',
  'active',
  false,
  false,
  false,
  false,
  false,
  NOW(),
  NOW()
),
-- Shop 4: Fresh Fruits
(
  (SELECT id FROM admins WHERE mobile = '9876543213' AND full_name = 'Neha Sharma' LIMIT 1),
  'Fresh Fruits Market',
  'Neha Sharma',
  'shop4@test.com',
  '9876543213',
  'Tropical and seasonal fruits. Sourced directly from orchards. Best prices guaranteed.',
  'fruits',
  '321 Fruit Garden',
  'Beside Park',
  'Hyderabad',
  'Telangana',
  'India',
  '500001',
  17.3850,
  78.4867,
  '2345678901234567',
  'AXIS0001234',
  'DEFGH1234I',
  'active',
  true,
  true,
  true,
  true,
  true,
  NOW(),
  NOW()
) ON CONFLICT DO NOTHING;

-- ============================================
-- 3. VERIFY SEEDING
-- ============================================

SELECT '=== SEED DATA SUMMARY ===' as info;
SELECT '✓ Test sellers and shops created' as status;

SELECT COUNT(*) as test_sellers_created FROM admins WHERE mobile LIKE '9876543%';
SELECT COUNT(*) as test_shops_created FROM shop_details WHERE shop_name LIKE '%Fresh%' OR shop_name LIKE '%Spice%' OR shop_name LIKE '%Dairy%' OR shop_name LIKE '%Fruits%';

-- List created sellers
SELECT '--- Test Sellers ---' as info;
SELECT
  id,
  full_name,
  mobile,
  status,
  verified_seller,
  'password' as test_password
FROM admins
WHERE mobile LIKE '9876543%'
ORDER BY mobile;

-- List created shops
SELECT '--- Test Shops ---' as info;
SELECT
  s.id,
  s.shop_name,
  a.full_name as seller_name,
  s.shop_status,
  s.shop_verification_status
FROM shop_details s
JOIN admins a ON s.admin_id = a.id
WHERE a.mobile LIKE '9876543%'
ORDER BY s.id;

SELECT '--- Usage Notes ---' as info;
SELECT 'Use these test credentials to log in as a seller:' as note;
SELECT 'Phone: 9876543210, Password: password' as credentials UNION ALL
SELECT 'Phone: 9876543211, Password: password' UNION ALL
SELECT 'Phone: 9876543212, Password: password' UNION ALL
SELECT 'Phone: 9876543213, Password: password';

SELECT '--- Templates ---' as info;
SELECT COUNT(*) as alert_templates_available FROM alert_templates;
SELECT 'Templates: Complete Your Setup, Welcome Offer, Seller Guide, Help Center' as template_list;
