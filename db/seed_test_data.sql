-- ============================================================
-- Locazar Mandi Backend — Comprehensive Seed
-- Targets the GORM AutoMigrate schema (bigint serial PKs)
-- All test credentials: password = Test@1234
-- ============================================================

-- ============================================================
-- CLEAN: clear test data (CASCADE handles dependent tables)
-- ============================================================
TRUNCATE TABLE
  admins, users,
  departments, categories, sub_categories,
  shop_details, products, product_items,
  coupons, offers, alert_templates,
  product_item_filter_types
RESTART IDENTITY CASCADE;

-- ============================================================
-- 1. DEPARTMENTS
-- ============================================================
INSERT INTO departments (name, sort_order, is_active, image_url, icon) VALUES
  ('Vegetables & Produce', 1, true, '', 'vegetable'),
  ('Dairy & Eggs',         2, true, '', 'dairy'),
  ('Grocery & Staples',    3, true, '', 'grocery'),
  ('Spices & Masala',      4, true, '', 'spice'),
  ('Fresh Fruits',          5, true, '', 'fruit');

-- ============================================================
-- 2. CATEGORIES
-- ============================================================
INSERT INTO categories (department_id, name, sort_order, is_active, image_url, icon) VALUES
  -- Vegetables & Produce
  ((SELECT id FROM departments WHERE name='Vegetables & Produce'), 'Leafy Greens',    1, true, '', 'leaf'),
  ((SELECT id FROM departments WHERE name='Vegetables & Produce'), 'Root Vegetables', 2, true, '', 'root'),
  ((SELECT id FROM departments WHERE name='Vegetables & Produce'), 'Seasonal Veggies',3, true, '', 'seasonal'),
  -- Dairy & Eggs
  ((SELECT id FROM departments WHERE name='Dairy & Eggs'),         'Milk & Cream',    1, true, '', 'milk'),
  ((SELECT id FROM departments WHERE name='Dairy & Eggs'),         'Cheese & Paneer', 2, true, '', 'cheese'),
  -- Grocery & Staples
  ((SELECT id FROM departments WHERE name='Grocery & Staples'),    'Rice & Grains',   1, true, '', 'rice'),
  ((SELECT id FROM departments WHERE name='Grocery & Staples'),    'Lentils & Pulses',2, true, '', 'dal'),
  -- Spices & Masala
  ((SELECT id FROM departments WHERE name='Spices & Masala'),      'Whole Spices',    1, true, '', 'whole'),
  ((SELECT id FROM departments WHERE name='Spices & Masala'),      'Ground Masala',   2, true, '', 'ground'),
  -- Fresh Fruits
  ((SELECT id FROM departments WHERE name='Fresh Fruits'),         'Citrus Fruits',   1, true, '', 'citrus'),
  ((SELECT id FROM departments WHERE name='Fresh Fruits'),         'Tropical Fruits', 2, true, '', 'tropical');

-- ============================================================
-- 3. SUB-CATEGORIES
-- ============================================================
INSERT INTO sub_categories (department_id, category_id, name, sort_order, is_active, image_url, icon) VALUES
  -- Leafy Greens
  ((SELECT id FROM departments WHERE name='Vegetables & Produce'),
   (SELECT id FROM categories WHERE name='Leafy Greens'),
   'Spinach', 1, true, '', 'spinach'),
  ((SELECT id FROM departments WHERE name='Vegetables & Produce'),
   (SELECT id FROM categories WHERE name='Leafy Greens'),
   'Fenugreek', 2, true, '', 'fenugreek'),
  -- Root Vegetables
  ((SELECT id FROM departments WHERE name='Vegetables & Produce'),
   (SELECT id FROM categories WHERE name='Root Vegetables'),
   'Onion & Garlic', 1, true, '', 'onion'),
  ((SELECT id FROM departments WHERE name='Vegetables & Produce'),
   (SELECT id FROM categories WHERE name='Root Vegetables'),
   'Potato', 2, true, '', 'potato'),
  -- Seasonal Veggies
  ((SELECT id FROM departments WHERE name='Vegetables & Produce'),
   (SELECT id FROM categories WHERE name='Seasonal Veggies'),
   'Tomato', 1, true, '', 'tomato'),
  ((SELECT id FROM departments WHERE name='Vegetables & Produce'),
   (SELECT id FROM categories WHERE name='Seasonal Veggies'),
   'Capsicum', 2, true, '', 'capsicum'),
  -- Dairy
  ((SELECT id FROM departments WHERE name='Dairy & Eggs'),
   (SELECT id FROM categories WHERE name='Milk & Cream'),
   'Cow Milk', 1, true, '', 'milk'),
  ((SELECT id FROM departments WHERE name='Dairy & Eggs'),
   (SELECT id FROM categories WHERE name='Cheese & Paneer'),
   'Paneer', 1, true, '', 'paneer'),
  -- Grocery
  ((SELECT id FROM departments WHERE name='Grocery & Staples'),
   (SELECT id FROM categories WHERE name='Rice & Grains'),
   'Basmati Rice', 1, true, '', 'rice'),
  ((SELECT id FROM departments WHERE name='Grocery & Staples'),
   (SELECT id FROM categories WHERE name='Lentils & Pulses'),
   'Toor Dal', 1, true, '', 'dal'),
  -- Spices
  ((SELECT id FROM departments WHERE name='Spices & Masala'),
   (SELECT id FROM categories WHERE name='Whole Spices'),
   'Cumin & Coriander', 1, true, '', 'cumin'),
  ((SELECT id FROM departments WHERE name='Spices & Masala'),
   (SELECT id FROM categories WHERE name='Ground Masala'),
   'Turmeric & Chilli', 1, true, '', 'turmeric'),
  -- Fruits
  ((SELECT id FROM departments WHERE name='Fresh Fruits'),
   (SELECT id FROM categories WHERE name='Citrus Fruits'),
   'Lemon & Lime', 1, true, '', 'lemon'),
  ((SELECT id FROM departments WHERE name='Fresh Fruits'),
   (SELECT id FROM categories WHERE name='Tropical Fruits'),
   'Mango', 1, true, '', 'mango');

-- ============================================================
-- 4. SELLERS (admins)
--    bcrypt('Test@1234', cost=10)
-- ============================================================
INSERT INTO admins (
  full_name, email, password,
  address_line1, address_line2, city, state, country, pincode,
  mobile, latitude, longitude,
  payment_status,
  bank_account_number, bank_ifsc, pan,
  aadhaar_last4, aadhaar_verified,
  verified_seller, status, user_name,
  created_at, updated_at
) VALUES
(
  'Raj Kumar', 'seller.raj@locazar.dev',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36gZvWFm',
  '12 Azad Mandi Road', 'Near SBI ATM', 'Delhi', 'Delhi', 'India', '110006',
  '9810001001', 28.7041, 77.1025,
  true,
  '1234567890123456', 'SBIN0001234', 'ABCDE1234F',
  '3456', true,
  true, 'active', 'seller_raj',
  NOW(), NOW()
),
(
  'Priya Singh', 'seller.priya@locazar.dev',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36gZvWFm',
  '55 Crawford Market', 'Ground Floor', 'Mumbai', 'Maharashtra', 'India', '400001',
  '9820002002', 18.9465, 72.8350,
  true,
  '9876543210123456', 'HDFC0001234', 'PQRST5678G',
  '8765', true,
  true, 'active', 'seller_priya',
  NOW(), NOW()
),
(
  'Arjun Patel', 'seller.arjun@locazar.dev',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36gZvWFm',
  '78 KR Market', 'Shop No 12', 'Bangalore', 'Karnataka', 'India', '560002',
  '9830003003', 12.9716, 77.5946,
  false,
  NULL, NULL, NULL,
  NULL, false,
  false, 'inactive', 'seller_arjun',
  NOW(), NOW()
),
(
  'Neha Sharma', 'seller.neha@locazar.dev',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36gZvWFm',
  '3 Rythu Bazaar Lane', 'Beside Old Clock', 'Hyderabad', 'Telangana', 'India', '500001',
  '9840004004', 17.3850, 78.4867,
  true,
  '2345678901234567', 'AXIS0001234', 'MNOPQ9012I',
  '5566', true,
  true, 'active', 'seller_neha',
  NOW(), NOW()
);

-- ============================================================
-- 5. SHOPS
-- ============================================================
INSERT INTO shop_details (
  admin_id, shop_name, owner_name, email, phone,
  address_line1, address_line2, city, state, country, pincode,
  latitude, longitude,
  shop_description, document_type, document_value, pan_number,
  shop_type, shop_status,
  bank_account_number, bank_ifsc,
  shop_verification_status, shop_verification_remarks,
  photo_shop_verification, business_doc_verification,
  identity_doc_verification, address_proof_verification,
  has_offers, created_at, updated_at
) VALUES
(
  (SELECT id FROM admins WHERE mobile='9810001001'),
  'Raj Fresh Vegetables', 'Raj Kumar', 'shop.raj@locazar.dev', '9810001001',
  '12 Azad Mandi Road', 'Near SBI ATM', 'Delhi', 'Delhi', 'India', '110006',
  28.7041, 77.1025,
  'Premium quality fresh vegetables sourced daily from local farms. Seasonal specials every day.',
  'gst', '27ABCDE1234F2Z5', 'ABCDE1234F',
  'retail', 'active',
  '1234567890123456', 'SBIN0001234',
  true, 'All documents verified by Locazar team',
  true, true, true, true,
  true, NOW(), NOW()
),
(
  (SELECT id FROM admins WHERE mobile='9820002002'),
  'Priya Spice & Dairy Hub', 'Priya Singh', 'shop.priya@locazar.dev', '9820002002',
  '55 Crawford Market', 'Ground Floor', 'Mumbai', 'Maharashtra', 'India', '400001',
  18.9465, 72.8350,
  'Authentic whole spices, freshly ground masala blends, and farm-fresh dairy products.',
  'pan', 'PQRST5678G', 'PQRST5678G',
  'retail', 'active',
  '9876543210123456', 'HDFC0001234',
  true, 'Verified — all checks passed',
  true, true, true, true,
  false, NOW(), NOW()
),
(
  (SELECT id FROM admins WHERE mobile='9830003003'),
  'Arjun Farm Dairy', 'Arjun Patel', 'shop.arjun@locazar.dev', '9830003003',
  '78 KR Market', 'Shop No 12', 'Bangalore', 'Karnataka', 'India', '560002',
  12.9716, 77.5946,
  'Fresh cow milk, curd, paneer and organic eggs delivered from our own farm within 24 hours.',
  NULL, NULL, NULL,
  'wholesale', 'inactive',
  NULL, NULL,
  false, 'Pending document verification',
  false, false, false, false,
  false, NOW(), NOW()
),
(
  (SELECT id FROM admins WHERE mobile='9840004004'),
  'Neha Grocers & Staples', 'Neha Sharma', 'shop.neha@locazar.dev', '9840004004',
  '3 Rythu Bazaar Lane', 'Beside Old Clock', 'Hyderabad', 'Telangana', 'India', '500001',
  17.3850, 78.4867,
  'Best quality basmati rice, dals, and grocery staples at wholesale prices. Bulk orders welcome.',
  'gst', '36MNOPQ9012I3Z9', 'MNOPQ9012I',
  'retail', 'active',
  '2345678901234567', 'AXIS0001234',
  true, 'Verified and active',
  true, true, true, true,
  false, NOW(), NOW()
);

-- ============================================================
-- 6. PRODUCTS
-- ============================================================
INSERT INTO products (name, description, category_id, department_id, image, shop_id, created_at, updated_at) VALUES
(
  'Fresh Spinach Bunch',
  'Crisp farm-fresh spinach bunches. Harvested daily, rich in iron and vitamins.',
  (SELECT id FROM categories WHERE name='Leafy Greens'),
  (SELECT id FROM departments WHERE name='Vegetables & Produce'),
  'https://placehold.co/400x300?text=Spinach',
  (SELECT id FROM shop_details WHERE shop_name='Raj Fresh Vegetables'),
  NOW(), NOW()
),
(
  'Red Onion',
  'Fresh Nashik red onions. Perfect for daily cooking. Mild flavour and firm texture.',
  (SELECT id FROM categories WHERE name='Root Vegetables'),
  (SELECT id FROM departments WHERE name='Vegetables & Produce'),
  'https://placehold.co/400x300?text=Onion',
  (SELECT id FROM shop_details WHERE shop_name='Raj Fresh Vegetables'),
  NOW(), NOW()
),
(
  'Desi Tomato',
  'Vine-ripened country tomatoes. Juicy and flavourful for curries and chutneys.',
  (SELECT id FROM categories WHERE name='Seasonal Veggies'),
  (SELECT id FROM departments WHERE name='Vegetables & Produce'),
  'https://placehold.co/400x300?text=Tomato',
  (SELECT id FROM shop_details WHERE shop_name='Raj Fresh Vegetables'),
  NOW(), NOW()
),
(
  'Pure Whole Cumin',
  'Sun-dried cumin seeds from Rajasthan. Aromatic and freshly packed.',
  (SELECT id FROM categories WHERE name='Whole Spices'),
  (SELECT id FROM departments WHERE name='Spices & Masala'),
  'https://placehold.co/400x300?text=Cumin',
  (SELECT id FROM shop_details WHERE shop_name='Priya Spice & Dairy Hub'),
  NOW(), NOW()
),
(
  'Premium Turmeric Powder',
  'High-curcumin turmeric, stone-ground. Bright golden colour with intense aroma.',
  (SELECT id FROM categories WHERE name='Ground Masala'),
  (SELECT id FROM departments WHERE name='Spices & Masala'),
  'https://placehold.co/400x300?text=Turmeric',
  (SELECT id FROM shop_details WHERE shop_name='Priya Spice & Dairy Hub'),
  NOW(), NOW()
),
(
  'Premium Basmati Rice',
  'Aged 2-year basmati from the Himalayan foothills. Extra-long grain, naturally fragrant.',
  (SELECT id FROM categories WHERE name='Rice & Grains'),
  (SELECT id FROM departments WHERE name='Grocery & Staples'),
  'https://placehold.co/400x300?text=Basmati+Rice',
  (SELECT id FROM shop_details WHERE shop_name='Neha Grocers & Staples'),
  NOW(), NOW()
),
(
  'Organic Toor Dal',
  'Unpolished organic toor dal. High protein, no additives. Sourced from Vidarbha farmers.',
  (SELECT id FROM categories WHERE name='Lentils & Pulses'),
  (SELECT id FROM departments WHERE name='Grocery & Staples'),
  'https://placehold.co/400x300?text=Toor+Dal',
  (SELECT id FROM shop_details WHERE shop_name='Neha Grocers & Staples'),
  NOW(), NOW()
);

-- ============================================================
-- 7. PRODUCT ITEMS (SKUs — admin_id stored as jsonb string)
-- ============================================================
INSERT INTO product_items (
  sub_category_name, sub_category_id, category_id, department_id,
  dynamic_fields, admin_id, shop_id, stock, created_at, updated_at
) VALUES
-- Spinach 250g
(
  'Spinach',
  (SELECT id FROM sub_categories WHERE name='Spinach'),
  (SELECT id FROM categories WHERE name='Leafy Greens'),
  (SELECT id FROM departments WHERE name='Vegetables & Produce'),
  '{"price": 1800, "currency": "INR", "weight": "250g", "unit": "bunch", "origin": "Haryana"}',
  to_jsonb((SELECT id FROM admins WHERE mobile='9810001001')::text),
  (SELECT id FROM shop_details WHERE shop_name='Raj Fresh Vegetables'),
  true, NOW(), NOW()
),
-- Spinach 500g
(
  'Spinach',
  (SELECT id FROM sub_categories WHERE name='Spinach'),
  (SELECT id FROM categories WHERE name='Leafy Greens'),
  (SELECT id FROM departments WHERE name='Vegetables & Produce'),
  '{"price": 3200, "currency": "INR", "weight": "500g", "unit": "bunch", "origin": "Haryana"}',
  to_jsonb((SELECT id FROM admins WHERE mobile='9810001001')::text),
  (SELECT id FROM shop_details WHERE shop_name='Raj Fresh Vegetables'),
  true, NOW(), NOW()
),
-- Onion 1kg
(
  'Onion & Garlic',
  (SELECT id FROM sub_categories WHERE name='Onion & Garlic'),
  (SELECT id FROM categories WHERE name='Root Vegetables'),
  (SELECT id FROM departments WHERE name='Vegetables & Produce'),
  '{"price": 4500, "currency": "INR", "weight": "1kg", "unit": "bag", "origin": "Nashik"}',
  to_jsonb((SELECT id FROM admins WHERE mobile='9810001001')::text),
  (SELECT id FROM shop_details WHERE shop_name='Raj Fresh Vegetables'),
  true, NOW(), NOW()
),
-- Onion 5kg
(
  'Onion & Garlic',
  (SELECT id FROM sub_categories WHERE name='Onion & Garlic'),
  (SELECT id FROM categories WHERE name='Root Vegetables'),
  (SELECT id FROM departments WHERE name='Vegetables & Produce'),
  '{"price": 20000, "currency": "INR", "weight": "5kg", "unit": "bag", "origin": "Nashik"}',
  to_jsonb((SELECT id FROM admins WHERE mobile='9810001001')::text),
  (SELECT id FROM shop_details WHERE shop_name='Raj Fresh Vegetables'),
  true, NOW(), NOW()
),
-- Tomato 1kg
(
  'Tomato',
  (SELECT id FROM sub_categories WHERE name='Tomato'),
  (SELECT id FROM categories WHERE name='Seasonal Veggies'),
  (SELECT id FROM departments WHERE name='Vegetables & Produce'),
  '{"price": 3500, "currency": "INR", "weight": "1kg", "unit": "loose", "origin": "Pune"}',
  to_jsonb((SELECT id FROM admins WHERE mobile='9810001001')::text),
  (SELECT id FROM shop_details WHERE shop_name='Raj Fresh Vegetables'),
  true, NOW(), NOW()
),
-- Cumin 100g
(
  'Cumin & Coriander',
  (SELECT id FROM sub_categories WHERE name='Cumin & Coriander'),
  (SELECT id FROM categories WHERE name='Whole Spices'),
  (SELECT id FROM departments WHERE name='Spices & Masala'),
  '{"price": 5000, "currency": "INR", "weight": "100g", "unit": "packet", "brand": "Farm Direct"}',
  to_jsonb((SELECT id FROM admins WHERE mobile='9820002002')::text),
  (SELECT id FROM shop_details WHERE shop_name='Priya Spice & Dairy Hub'),
  true, NOW(), NOW()
),
-- Turmeric 200g
(
  'Turmeric & Chilli',
  (SELECT id FROM sub_categories WHERE name='Turmeric & Chilli'),
  (SELECT id FROM categories WHERE name='Ground Masala'),
  (SELECT id FROM departments WHERE name='Spices & Masala'),
  '{"price": 7500, "currency": "INR", "weight": "200g", "unit": "pack", "curcumin_pct": "5.2"}',
  to_jsonb((SELECT id FROM admins WHERE mobile='9820002002')::text),
  (SELECT id FROM shop_details WHERE shop_name='Priya Spice & Dairy Hub'),
  true, NOW(), NOW()
),
-- Basmati Rice 1kg
(
  'Basmati Rice',
  (SELECT id FROM sub_categories WHERE name='Basmati Rice'),
  (SELECT id FROM categories WHERE name='Rice & Grains'),
  (SELECT id FROM departments WHERE name='Grocery & Staples'),
  '{"price": 18000, "currency": "INR", "weight": "1kg", "unit": "bag", "age": "2 year"}',
  to_jsonb((SELECT id FROM admins WHERE mobile='9840004004')::text),
  (SELECT id FROM shop_details WHERE shop_name='Neha Grocers & Staples'),
  true, NOW(), NOW()
),
-- Basmati Rice 5kg
(
  'Basmati Rice',
  (SELECT id FROM sub_categories WHERE name='Basmati Rice'),
  (SELECT id FROM categories WHERE name='Rice & Grains'),
  (SELECT id FROM departments WHERE name='Grocery & Staples'),
  '{"price": 85000, "currency": "INR", "weight": "5kg", "unit": "bag", "age": "2 year"}',
  to_jsonb((SELECT id FROM admins WHERE mobile='9840004004')::text),
  (SELECT id FROM shop_details WHERE shop_name='Neha Grocers & Staples'),
  true, NOW(), NOW()
),
-- Toor Dal 500g
(
  'Toor Dal',
  (SELECT id FROM sub_categories WHERE name='Toor Dal'),
  (SELECT id FROM categories WHERE name='Lentils & Pulses'),
  (SELECT id FROM departments WHERE name='Grocery & Staples'),
  '{"price": 12000, "currency": "INR", "weight": "500g", "unit": "packet", "type": "unpolished"}',
  to_jsonb((SELECT id FROM admins WHERE mobile='9840004004')::text),
  (SELECT id FROM shop_details WHERE shop_name='Neha Grocers & Staples'),
  true, NOW(), NOW()
);

-- ============================================================
-- 8. CUSTOMERS (users)
-- ============================================================
-- Columns match User struct: email_verified, phone_verified, date_of_birth (no legacy verified/age).
INSERT INTO users (first_name, last_name, email, phone, password, email_verified, phone_verified, block_status, date_of_birth, created_at, updated_at) VALUES
  ('Amit',   'Verma',   'amit.verma@example.com',   '9911100001',
   '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36gZvWFm',
   true, true, false, '1990-04-15', NOW(), NOW()),
  ('Sunita', 'Rao',     'sunita.rao@example.com',   '9911100002',
   '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36gZvWFm',
   true, true, false, '1985-11-22', NOW(), NOW()),
  ('Farhan', 'Qureshi', 'farhan.q@example.com',     '9911100003',
   '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36gZvWFm',
   false, true, false, '1995-07-08', NOW(), NOW()),
  ('Deepa',  'Nair',    'deepa.nair@example.com',   '9911100004',
   '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36gZvWFm',
   true, true, false, '1992-03-30', NOW(), NOW()),
  ('Test',   'Blocked', 'blocked.user@example.com', '9911100099',
   '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36gZvWFm',
   true, true, true, '1988-01-01', NOW(), NOW());

-- ============================================================
-- 9. COUPONS
-- minimum_cart_price is in paise (INR * 100)
-- ============================================================
INSERT INTO coupons (coupon_name, coupon_code, expire_date, description, discount_rate, minimum_cart_price, block_status, created_at, updated_at) VALUES
  ('WELCOME10', 'WELCOME10',
   NOW() + INTERVAL '90 days',
   'Get 10% off on your first order. Valid for all new customers.',
   10, 10000, false, NOW(), NOW()),
  ('FREESHIP', 'FREESHIP',
   NOW() + INTERVAL '30 days',
   'Free delivery on orders above 299.',
   0, 29900, false, NOW(), NOW()),
  ('MONSOON20', 'MONSOON20',
   NOW() + INTERVAL '15 days',
   'Monsoon special - flat 20% off on vegetables and fruits.',
   20, 19900, false, NOW(), NOW());

-- ============================================================
-- 10. OFFERS
-- discount_rate is a percentage (0-100) stored as bigint
-- ============================================================
INSERT INTO offers (name, description, discount_rate, offer_type, start_date, end_date, image, thumbnail, sort_order, is_active, created_at, updated_at) VALUES
  (
    'Fresh Daily Deal',
    'Save 15% on all vegetables ordered before 10 AM',
    15, 'percentage',
    NOW(), NOW() + INTERVAL '30 days',
    'https://placehold.co/800x400?text=Fresh+Daily+Deal',
    'https://placehold.co/200x200?text=Fresh+Daily',
    1, true, NOW(), NOW()
  ),
  (
    'Bulk Saver',
    'Buy 5kg or more of any staple for flat 5% off',
    5, 'percentage',
    NOW(), NOW() + INTERVAL '60 days',
    'https://placehold.co/800x400?text=Bulk+Saver',
    'https://placehold.co/200x200?text=Bulk+Saver',
    2, true, NOW(), NOW()
  );

-- ============================================================
-- 11. ALERT TEMPLATES
-- ============================================================
INSERT INTO alert_templates (
  key, title, description, type, priority,
  is_active, enabled, conditions, actions,
  frequency_config, validity,
  template_type, display_type, content_schema, flow_key
) VALUES
(
  'complete_profile',
  'Complete Your Seller Profile',
  'Add your bank details, GST number and shop photos to start receiving orders.',
  'warning', 10, true, true,
  '{"rule": "profile_incomplete", "fields": ["bank_account_number", "pan", "shop_image_url"]}',
  '[{"label": "Complete Now", "action_url": "/seller/profile", "action_type": "navigate"}]',
  '{"type": "once"}', NULL,
  'onboarding', 'bottom_sheet',
  '{"cta": "Complete Now", "icon": "profile_incomplete"}',
  'onboarding_profile'
),
(
  'verify_shop',
  'Shop Verification Pending',
  'Your shop is under review. Upload clear photos of your shop front and business documents to speed up verification.',
  'info', 8, true, true,
  '{"rule": "shop_unverified"}',
  '[{"label": "Upload Docs", "action_url": "/seller/verification", "action_type": "navigate"}, {"label": "Dismiss", "action_url": "", "action_type": "dismiss"}]',
  '{"type": "daily", "limit": 1}', NULL,
  'onboarding', 'banner',
  '{"cta": "Upload Docs", "icon": "document_upload"}',
  'onboarding_verification'
),
(
  'add_first_product',
  'Add Your First Product',
  'You have not listed any products yet. Start adding products so customers can discover your shop.',
  'info', 7, true, true,
  '{"rule": "no_products"}',
  '[{"label": "Add Product", "action_url": "/seller/products/new", "action_type": "navigate"}]',
  '{"type": "daily", "limit": 1}', NULL,
  'onboarding', 'bottom_sheet',
  '{"cta": "Add Product", "icon": "add_product"}',
  'onboarding_catalog'
),
(
  'subscription_expiring',
  'Your Subscription Expires Soon',
  'Your plan expires in 7 days. Renew now to keep your shop visible to customers.',
  'critical', 9, true, true,
  '{"rule": "subscription_expiring_soon", "days_before": 7}',
  '[{"label": "Renew Plan", "action_url": "/seller/subscription", "action_type": "navigate"}, {"label": "Remind Later", "action_url": "", "action_type": "dismiss"}]',
  '{"type": "daily", "limit": 1}',
  '{"days_before_expiry": 7}',
  'subscription', 'modal',
  '{"cta": "Renew Plan", "icon": "subscription_alert"}',
  'subscription_renewal'
),
(
  'welcome_seller',
  'Welcome to Locazar!',
  'You are all set. Explore the seller dashboard to manage your shop, add products, and grow your business.',
  'info', 5, true, true,
  '{"rule": "new_seller", "hours_after_signup": 1}',
  '[{"label": "Explore Dashboard", "action_url": "/seller/dashboard", "action_type": "navigate"}]',
  '{"type": "once"}', NULL,
  'announcement', 'bottom_sheet',
  '{"cta": "Explore Dashboard", "icon": "welcome"}',
  'welcome_flow'
),
(
  'low_stock_warning',
  'Some Items Are Low on Stock',
  'Customers are browsing your shop but some popular items are marked out of stock. Update your inventory.',
  'warning', 6, true, true,
  '{"rule": "has_out_of_stock_items"}',
  '[{"label": "Update Stock", "action_url": "/seller/products", "action_type": "navigate"}]',
  '{"type": "weekly", "limit": 2}', NULL,
  'operational', 'banner',
  '{"cta": "Update Stock", "icon": "inventory"}',
  'inventory_management'
),
(
  'seller_guide',
  'Seller Success Guide',
  'Read our 5-step guide to getting your first 10 orders on Locazar.',
  'info', 3, true, true,
  '{"rule": "orders_less_than", "count": 10}',
  '[{"label": "Read Guide", "action_url": "/seller/guide", "action_type": "navigate"}, {"label": "Skip", "action_url": "", "action_type": "dismiss"}]',
  '{"type": "once"}', NULL,
  'education', 'bottom_sheet',
  '{"cta": "Read Guide", "icon": "guide"}',
  'education_guide'
);

-- ============================================================
-- 12. PRODUCT ITEM FILTER TYPES (global)
-- ============================================================
INSERT INTO product_item_filter_types (filter_name, created_at, updated_at)
SELECT 'All', NOW(), NOW() WHERE NOT EXISTS (SELECT 1 FROM product_item_filter_types WHERE filter_name='All');

INSERT INTO product_item_filter_types (filter_name, created_at, updated_at)
SELECT 'Offers', NOW(), NOW() WHERE NOT EXISTS (SELECT 1 FROM product_item_filter_types WHERE filter_name='Offers');

INSERT INTO product_item_filter_types (filter_name, created_at, updated_at)
SELECT 'New Arrivals', NOW(), NOW() WHERE NOT EXISTS (SELECT 1 FROM product_item_filter_types WHERE filter_name='New Arrivals');

INSERT INTO product_item_filter_types (filter_name, created_at, updated_at)
SELECT 'Popular', NOW(), NOW() WHERE NOT EXISTS (SELECT 1 FROM product_item_filter_types WHERE filter_name='Popular');

-- ============================================================
-- SUMMARY
-- ============================================================
SELECT '=== SEED COMPLETE ===' AS status;

SELECT entity, rows FROM (
  SELECT 'departments'    AS entity, COUNT(*)::int AS rows FROM departments         UNION ALL
  SELECT 'categories',              COUNT(*)        FROM categories                  UNION ALL
  SELECT 'sub_categories',         COUNT(*)        FROM sub_categories              UNION ALL
  SELECT 'admins (sellers)',        COUNT(*)        FROM admins                      UNION ALL
  SELECT 'shop_details',           COUNT(*)        FROM shop_details                UNION ALL
  SELECT 'products',               COUNT(*)        FROM products                    UNION ALL
  SELECT 'product_items',          COUNT(*)        FROM product_items               UNION ALL
  SELECT 'users (customers)',       COUNT(*)        FROM users                       UNION ALL
  SELECT 'coupons',                COUNT(*)        FROM coupons                     UNION ALL
  SELECT 'offers',                 COUNT(*)        FROM offers                      UNION ALL
  SELECT 'alert_templates',        COUNT(*)        FROM alert_templates             UNION ALL
  SELECT 'product_item_filter_types', COUNT(*)     FROM product_item_filter_types
) t ORDER BY entity;

SELECT '--- Seller Credentials ---' AS info;
SELECT full_name, mobile AS phone, status, verified_seller FROM admins ORDER BY mobile;

SELECT '--- Customer Credentials ---' AS info;
SELECT first_name || ' ' || last_name AS name, phone, email_verified, phone_verified, block_status FROM users ORDER BY phone;

SELECT '--- All passwords = Test@1234 ---' AS info;
