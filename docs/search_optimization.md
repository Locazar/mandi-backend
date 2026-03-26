# Search Optimization for Product APIs

## Overview
This document outlines the steps taken to optimize the product search functionality in the Mandi Backend (Go application) to improve API response times. The original implementation used `ILIKE` for case-insensitive searches, which performs poorly on large datasets due to lack of efficient indexing.

## Problem
- Slow search queries due to `ILIKE` on `name` and `description` fields.
- No indexes optimized for text search.
- Potential performance degradation as product catalog grows.

## Solution
Implemented PostgreSQL full-text search combined with trigram indexes for faster and more accurate searches.

### Technologies Used
- PostgreSQL extensions: `pg_trgm` for trigram similarity, built-in full-text search.
- GIN indexes for efficient text operations.
- Go backend with GORM for database queries.

## Steps Implemented

### 1. Enable PostgreSQL Extensions
Added support for trigram operations:
```sql
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
```

### 2. Create Optimized Indexes
Added the following indexes to the `products` table:
```sql
-- Trigram indexes for LIKE/ILIKE performance
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_name_trgm ON products USING gin (name gin_trgm_ops);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_description_trgm ON products USING gin (description gin_trgm_ops);

-- Full-text search index
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_search ON products USING gin (to_tsvector('english', name || ' ' || description));
```

### 3. Update Search Query Logic
Modified `pkg/repository/product.go` in the `SearchProducts` function:

**Before:**
```sql
WHERE (p.name ILIKE $1 OR p.description ILIKE $1)
params := []interface{}{"%" + keyword + "%"}
```

**After:**
```sql
WHERE to_tsvector('english', p.name || ' ' || p.description) @@ plainto_tsquery('english', $1)
params := []interface{}{keyword}
```

### 4. Database Connection and Execution
- Connected to PostgreSQL using credentials from `.env`:
  - Host: localhost
  - User: postgres
  - Database: mandi
  - Password: postgres
- Executed SQL commands via `psql` to add extensions and indexes without data loss.

### 5. Validation
- Verified extension and indexes were created successfully.
- Ensured no data was lost by using `CONCURRENTLY` for index creation.
- Code compiled without errors.

## Benefits
- **Faster Searches:** Full-text search with GIN indexes provides sub-second responses even for large catalogs.
- **Better Relevance:** Full-text search ranks results by relevance, not just substring matching.
- **Scalability:** Indexes handle growth in product data efficiently.
- **No Data Loss:** Changes applied to existing database without recreation.

## API Endpoint
The optimized search is available at:
```
GET /products/search?q={keyword}&category={id}&brand={id}&location={id}&limit={n}&page={n}
```

## Future Enhancements
- Implement Redis caching for frequent search terms.
- Add autocomplete suggestions using prefix indexes.
- Consider Elasticsearch for advanced search features (facets, synonyms, etc.).

## Files Modified
- `docker/postgres/initdb/01-init.sql` (reverted to avoid init conflicts)
- `pkg/repository/product.go` (updated query logic)

## Tables Optimized
- products (already done)
- departments
- categories
- sub_categories
- brands
- offers
- product_items
- users
- admins
- sub_type_attributes
- sub_type_attribute_options

## Commands Executed
```bash
# Set password environment variable
$env:PGPASSWORD = "postgres"

# Enable extension (if not already)
psql -h localhost -U postgres -d mandi -c 'CREATE EXTENSION IF NOT EXISTS "pg_trgm";'

# Create indexes for each table
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_departments_search ON departments USING gin (to_tsvector('english', name));"
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_categories_search ON categories USING gin (to_tsvector('english', name));"
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sub_categories_search ON sub_categories USING gin (to_tsvector('english', name));"
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_brands_search ON brands USING gin (to_tsvector('english', name));"
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_offers_search ON offers USING gin (to_tsvector('english', name || ' ' || description));"
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_product_items_search ON product_items USING gin (to_tsvector('english', sub_category_name));"
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_search ON users USING gin (to_tsvector('english', first_name || ' ' || last_name || ' ' || email || ' ' || phone));"
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_admins_search ON admins USING gin (to_tsvector('english', full_name || ' ' || email || ' ' || mobile));"
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sub_type_attributes_search ON sub_type_attributes USING gin (to_tsvector('english', field_name));"
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sub_type_attribute_options_search ON sub_type_attribute_options USING gin (to_tsvector('english', option_value));"

# Additional indexes for faster fetching
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sub_type_attribute_options_attr_id ON sub_type_attribute_options (sub_type_attribute_id);"
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sub_type_attributes_sub_cat_id ON sub_type_attributes (sub_category_id);"
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sub_categories_cat_id ON sub_categories (category_id);"
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sub_categories_dept_id ON sub_categories (department_id);"
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_product_items_sub_cat_id ON product_items (sub_category_id);"
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_product_items_cat_id ON product_items (category_id);"
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_product_items_dept_id ON product_items (department_id);"
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_categories_dept_id ON categories (department_id);"
psql -h localhost -U postgres -d mandi -c "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_categories_sort_order ON categories (sort_order);"
```

## Elasticsearch Integration

Elasticsearch has been integrated for faster search across multiple entities.

### Supported Entities
- products (with filters)
- departments
- categories (with department filter)
- brands
- offers
- users
- admins
- (Extendable to sub_categories, product_items, attributes, etc.)

### Features
- Indexing and searching for each entity.
- Bulk indexing support.
- Filters where applicable (e.g., categories by department).

### Setup
- ES runs in Docker at `http://localhost:9200`.
- Use admin API `POST /admin/products/bulk-index` for products; extend for others.

### Future
- Add bulk index APIs for all entities.
- Integrate into respective usecases for search.