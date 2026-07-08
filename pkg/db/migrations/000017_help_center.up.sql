CREATE TABLE IF NOT EXISTS help_settings (
    id VARCHAR(32) PRIMARY KEY,
    support_phone VARCHAR(20),
    support_email VARCHAR(120),
    whatsapp_number VARCHAR(20),
    support_hours VARCHAR(120),
    about_text TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS help_faqs (
    id VARCHAR(32) PRIMARY KEY,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_help_faqs_sort_order ON help_faqs (sort_order);
