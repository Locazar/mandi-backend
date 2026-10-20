-- product_items has no real title column — the display name today is the
-- overloaded sub_category_name (so every listing in, say, "Chairs" is
-- literally named "Chairs"). This adds an optional, seller-editable product
-- name; NULL/empty falls back to sub_category_name at read time so nothing
-- existing breaks. AI listing suggestion (pkg/service/ai) prefills this from
-- the product photo, but the seller can always overwrite it.
ALTER TABLE product_items ADD COLUMN IF NOT EXISTS name VARCHAR(120);
