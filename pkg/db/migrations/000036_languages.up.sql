-- Selectable languages for the seller-app onboarding language picker
-- (feature-flag gated). Read-only from the client via GET /api/languages.
CREATE TABLE IF NOT EXISTS languages (
    id          VARCHAR(32) PRIMARY KEY,
    code        VARCHAR(10) NOT NULL UNIQUE,
    name        VARCHAR(50) NOT NULL,
    native_name VARCHAR(50) NOT NULL,
    sort_order  INT         NOT NULL DEFAULT 0,
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_languages_active_order ON languages (is_active, sort_order);

INSERT INTO languages (id, code, name, native_name, sort_order) VALUES
    ('lang_seed_en', 'en', 'English',   'English',   0),
    ('lang_seed_hi', 'hi', 'Hindi',     'हिन्दी',     1),
    ('lang_seed_mr', 'mr', 'Marathi',   'मराठी',      2),
    ('lang_seed_bn', 'bn', 'Bengali',   'বাংলা',      3),
    ('lang_seed_ta', 'ta', 'Tamil',     'தமிழ்',      4),
    ('lang_seed_te', 'te', 'Telugu',    'తెలుగు',     5),
    ('lang_seed_kn', 'kn', 'Kannada',   'ಕನ್ನಡ',      6),
    ('lang_seed_gu', 'gu', 'Gujarati',  'ગુજરાતી',    7),
    ('lang_seed_ml', 'ml', 'Malayalam', 'മലയാളം',    8),
    ('lang_seed_pa', 'pa', 'Punjabi',   'ਪੰਜਾਬੀ',     9),
    ('lang_seed_or', 'or', 'Odia',      'ଓଡ଼ିଆ',      10)
ON CONFLICT (code) DO NOTHING;
