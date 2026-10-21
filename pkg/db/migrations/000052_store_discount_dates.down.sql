ALTER TABLE store_discounts
  DROP COLUMN IF EXISTS start_date,
  DROP COLUMN IF EXISTS end_date;
