-- Caps how many product items a seller (admin with role='seller') may
-- upload in total across their shop(s). Enforced in ProductUseCase.SaveProductItem.
ALTER TABLE admins ADD COLUMN IF NOT EXISTS product_limit INTEGER NOT NULL DEFAULT 100;
