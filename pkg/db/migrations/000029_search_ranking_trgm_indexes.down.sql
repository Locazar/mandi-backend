-- Drop the trigram indexes added for search relevance ranking.
-- Leave the pg_trgm extension in place (other objects may depend on it).
DROP INDEX IF EXISTS idx_departments_name_trgm;
DROP INDEX IF EXISTS idx_sub_categories_name_trgm;
DROP INDEX IF EXISTS idx_categories_name_trgm;
DROP INDEX IF EXISTS idx_product_items_sub_category_name_trgm;
