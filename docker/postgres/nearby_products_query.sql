-- Corrected Query: Get nearby products by pincode (without locations table)
-- This query assumes pincode is stored in the dynamic_fields JSONB column of product_items

-- Version 1: Simple filter by pincode (no geospatial calculations)
SELECT
    pi.id AS product_item_id,
    p.name,
    p.description,
    p.category_id,
    c.name AS category_name,
    sc.name AS sub_category_name,
    pi.sub_category_id,
    p.image,
    pi.created_at,
    pi.updated_at
FROM product_items pi
INNER JOIN products p ON pi.id = p.id
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN sub_categories sc ON pi.sub_category_id = sc.id
WHERE pi.dynamic_fields->>'pincode' = $1  -- Extract pincode from JSONB
ORDER BY pi.created_at DESC
LIMIT $2 OFFSET $3;

-- Version 2: Filter by pincode with additional fields from dynamic_fields
SELECT
    pi.id AS product_item_id,
    p.name,
    p.description,
    p.category_id,
    c.name AS category_name,
    sc.name AS sub_category_name,
    pi.sub_category_id,
    p.image,
    pi.dynamic_fields->>'pincode' AS pincode,
    pi.dynamic_fields->>'latitude' AS latitude,
    pi.dynamic_fields->>'longitude' AS longitude,
    pi.dynamic_fields->>'city' AS city,
    pi.created_at,
    pi.updated_at
FROM product_items pi
INNER JOIN products p ON pi.id = p.id
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN sub_categories sc ON pi.sub_category_id = sc.id
WHERE pi.dynamic_fields->>'pincode' = $1
ORDER BY pi.created_at DESC
LIMIT $2 OFFSET $3;

-- Version 3: Filter by city (if you want multiple pincodes from same city)
SELECT
    pi.id AS product_item_id,
    p.name,
    p.description,
    p.category_id,
    c.name AS category_name,
    sc.name AS sub_category_name,
    pi.sub_category_id,
    p.image,
    pi.dynamic_fields->>'pincode' AS pincode,
    pi.dynamic_fields->>'city' AS city,
    pi.created_at,
    pi.updated_at
FROM product_items pi
INNER JOIN products p ON pi.id = p.id
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN sub_categories sc ON pi.sub_category_id = sc.id
WHERE pi.dynamic_fields->>'city' = $1  -- Filter by city instead
ORDER BY pi.created_at DESC
LIMIT $2 OFFSET $3;

-- Version 4: If you have distance-based filtering (with latitude/longitude in dynamic_fields)
-- This requires PostGIS but no locations table
SELECT
    pi.id AS product_item_id,
    p.name,
    p.description,
    p.category_id,
    c.name AS category_name,
    sc.name AS sub_category_name,
    pi.sub_category_id,
    p.image,
    pi.dynamic_fields->>'pincode' AS pincode,
    pi.dynamic_fields->>'latitude'::NUMERIC AS latitude,
    pi.dynamic_fields->>'longitude'::NUMERIC AS longitude,
    -- Calculate distance (simple Pythagorean, not accurate for long distances)
    SQRT(
        POW((pi.dynamic_fields->>'latitude'::NUMERIC - $2::NUMERIC) * 111, 2) +
        POW((pi.dynamic_fields->>'longitude'::NUMERIC - $3::NUMERIC) * 111, 2)
    ) AS distance_km,
    pi.created_at,
    pi.updated_at
FROM product_items pi
INNER JOIN products p ON pi.id = p.id
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN sub_categories sc ON pi.sub_category_id = sc.id
WHERE pi.dynamic_fields IS NOT NULL
ORDER BY distance_km ASC
LIMIT $4 OFFSET $5;
