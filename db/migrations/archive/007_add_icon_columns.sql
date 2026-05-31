-- Add icon columns to departments, categories and sub_categories
ALTER TABLE departments ADD COLUMN IF NOT EXISTS icon VARCHAR(255);
ALTER TABLE categories ADD COLUMN IF NOT EXISTS icon VARCHAR(255);
ALTER TABLE sub_categories ADD COLUMN IF NOT EXISTS icon VARCHAR(255);
