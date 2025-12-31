-- Sample SQL to insert an active offer and link it to a product item
-- Replace the IDs and dates as needed for your test

-- Insert an active offer
INSERT INTO offers (id, name, discount_rate, description, start_date, end_date, image, thumbnail)
VALUES (1001, 'Test Offer', 15, 'Test discount offer', NOW() - INTERVAL '1 day', NOW() + INTERVAL '7 days', 'offer_image.jpg', 'offer_thumb.jpg');

-- Link the offer to your product item (e.g., product_item_id = 36)
INSERT INTO offer_products (id, product_item_id, product_id, offer_id)
VALUES (2001, 36, 0, 1001); -- product_id can be 0 or the correct product if needed
