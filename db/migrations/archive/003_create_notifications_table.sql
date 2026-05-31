-- Migration for notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    sender_type VARCHAR(50) NOT NULL,
    receiver_type VARCHAR(50) NOT NULL,
    type VARCHAR(100) NOT NULL,
    sender_id INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    body TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    receiver_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    variation_id INTEGER NOT NULL,
    shop_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    admin_id INTEGER NOT NULL,
    order_id INTEGER NOT NULL,
    offer_id INTEGER NOT NULL,
    notification_meta_data TEXT,
    status VARCHAR(50) NOT NULL,
    created_at VARCHAR(50) NOT NULL,
    updated_at VARCHAR(50) NOT NULL
);

-- Migration for notification_device_tokens table
CREATE TABLE IF NOT EXISTS notification_device_tokens (
    id SERIAL PRIMARY KEY,
    owner_id VARCHAR(100) NOT NULL,
    owner_type VARCHAR(10) NOT NULL CHECK (owner_type IN ('user','seller')),
    token VARCHAR(255) UNIQUE NOT NULL,
    platform VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);