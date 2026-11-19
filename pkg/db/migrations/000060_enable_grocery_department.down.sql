-- 000060_enable_grocery_department.down.sql
-- Revert: disable the Grocery department.

UPDATE departments
SET is_active = FALSE
WHERE id = 'dept_00005';
