CREATE TABLE IF NOT EXISTS lead_requests (
    id SERIAL PRIMARY KEY,
    phone VARCHAR(20) NOT NULL,
    requirement TEXT NOT NULL,
    city VARCHAR(100) NOT NULL DEFAULT '',
    source VARCHAR(500),
    status VARCHAR(20) NOT NULL DEFAULT 'new',
    admin_note TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lead_requests_phone ON lead_requests(phone);
CREATE INDEX IF NOT EXISTS idx_lead_requests_status ON lead_requests(status);
CREATE INDEX IF NOT EXISTS idx_lead_requests_city ON lead_requests(city);
