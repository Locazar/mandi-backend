ALTER TABLE shop_details
    ADD COLUMN IF NOT EXISTS pan_front_image_url TEXT,
    ADD COLUMN IF NOT EXISTS pan_back_image_url  TEXT;
