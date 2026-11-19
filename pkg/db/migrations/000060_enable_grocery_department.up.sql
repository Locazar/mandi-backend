-- 000060_enable_grocery_department.up.sql
-- Enable the Grocery department (dept_00005) which ships inactive in the seed.

UPDATE departments
SET is_active = TRUE
WHERE id = 'dept_00005';
