-- Restore the original FKs. Rollback will fail if any rows now carry
-- non-user IDs (e.g. admin IDs); clean such rows before rolling back.
ALTER TABLE subscription_orders
    ADD CONSTRAINT subscription_orders_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;

ALTER TABLE user_subscriptions
    ADD CONSTRAINT user_subscriptions_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
