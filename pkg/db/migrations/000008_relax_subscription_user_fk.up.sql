-- Sellers (admins) subscribe to billing plans through the same flow as users.
-- The FK to users(id) prevents admin IDs from being stored in user_id.
ALTER TABLE subscription_orders DROP CONSTRAINT IF EXISTS subscription_orders_user_id_fkey;
ALTER TABLE user_subscriptions  DROP CONSTRAINT IF EXISTS user_subscriptions_user_id_fkey;
