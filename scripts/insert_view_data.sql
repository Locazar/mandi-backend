-- Insert test view data for product_item_views table
-- Assuming admin_id is the viewer, product_item_id is the item viewed, view_count is 1 per view

-- For product_item_id 40, add 30 views (from different admins or same)
INSERT INTO product_item_views (product_item_id, admin_id, view_count) VALUES
(40, 'admin1', 1),
(40, 'admin2', 1),
-- ... repeat for 30 times, but to sum, can do one row with 30, but since it's per view, multiple rows
-- For simplicity, insert one row with view_count = 30
(40, 'admin1', 30);

-- For product_item_id 39, 10 views
INSERT INTO product_item_views (product_item_id, admin_id, view_count) VALUES
(39, 'admin1', 10);

-- For product_item_id 38, 70 views
INSERT INTO product_item_views (product_item_id, admin_id, view_count) VALUES
(38, 'admin1', 70);

-- For others, leave as 0 or add some
INSERT INTO product_item_views (product_item_id, admin_id, view_count) VALUES
(49, 'admin1', 5),
(48, 'admin1', 3);