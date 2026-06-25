ALTER TABLE advertisements
    DROP COLUMN IF EXISTS audience,
    DROP COLUMN IF EXISTS department_id,
    DROP COLUMN IF EXISTS category_id;
