-- Optional per-shop override of the WhatsApp-status share link (falls back
-- to the global Config page's `webLink` base + shop id when empty).
ALTER TABLE shop_details ADD COLUMN IF NOT EXISTS whatsapp_share_link TEXT NOT NULL DEFAULT '';
