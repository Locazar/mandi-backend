-- Best-effort relevance ranking (SEARCH_RANKING_ENABLED) relies on leading-wildcard
-- ILIKE '%q%' across the searched name columns, which a b-tree index cannot serve.
-- Enable pg_trgm and add GIN trigram indexes so those ILIKE scans stay index-assisted.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_product_items_sub_category_name_trgm
    ON product_items USING gin (sub_category_name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_categories_name_trgm
    ON categories USING gin (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_sub_categories_name_trgm
    ON sub_categories USING gin (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_departments_name_trgm
    ON departments USING gin (name gin_trgm_ops);
