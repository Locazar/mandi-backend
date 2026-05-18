# Seller Alerts System - Database Setup Guide

## Overview
The Seller Alerts System uses 4 main tables to manage dynamic seller alerts:

1. **alert_templates** - Database-driven rule configurations
2. **alerts** - Alert instances for sellers
3. **alert_actions** - Actions/buttons associated with alerts
4. **seller_alert_logs** - Audit trail of alert interactions

## Setup Instructions

### Option 1: Using Docker (Recommended)
The setup file is automatically executed when your PostgreSQL Docker container starts:
```bash
docker-compose up -d
```

The file `docker/postgres/initdb/06_alert_system.sql` will be automatically executed on container startup.

### Option 2: Manual SQL Execution
Run the setup script directly in your PostgreSQL database:

```bash
# Using psql
psql -U your_user -d your_database -f ALERT_SYSTEM_SETUP.sql

# Or using pgAdmin/DBeaver - copy and paste the SQL from ALERT_SYSTEM_SETUP.sql
```

### Option 3: Using Migration Files
Use the migration file located in:
```
db/migrations/006_alert_system.sql
```

---

## Table Structures

### alert_templates
```sql
CREATE TABLE alert_templates (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE,              -- Unique identifier (e.g., "db_high_product_price")
    title VARCHAR(255),                   -- Display title
    description TEXT,                     -- Detailed description
    type VARCHAR(50),                     -- info, warning, critical
    priority INTEGER,                     -- Sort order (higher = more important)
    is_active BOOLEAN,                    -- Display in UI
    enabled BOOLEAN,                      -- Include in evaluation
    conditions JSONB,                     -- JSON condition structure
    actions JSONB,                        -- JSON actions array
    frequency_config JSONB,               -- Frequency configuration
    validity JSONB,                       -- Time window validity
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

**Sample condition JSON:**
```json
{
  "conditions": [
    {"field": "product_count", "operator": ">", "value": 5},
    {"field": "is_verified", "operator": "=", "value": false}
  ],
  "logic": "AND"
}
```

**Supported fields:**
- `product_count` - Number of products seller has
- `is_verified` - Shop verification status
- `has_shop_photo` - Has shop photo
- `shop_name` - Shop name (string)
- `shop_status` - Shop status (string)
- `admin_id`, `shop_id` - IDs

**Supported operators:**
- `=`, `!=`, `>`, `<`, `>=`, `<=`
- `in` - Check if value is in array
- `contains` - String contains

### alerts
```sql
CREATE TABLE alerts (
    id SERIAL PRIMARY KEY,
    seller_id INTEGER,                    -- Seller/Admin ID
    key VARCHAR(100),                     -- Alert key (from template or rule)
    title VARCHAR(255),
    content TEXT,
    description TEXT,
    type VARCHAR(50),                     -- info, warning, critical
    priority INTEGER,
    is_active BOOLEAN,
    valid_from TIMESTAMP,                 -- Start showing alert
    valid_until TIMESTAMP,                -- Stop showing alert
    metadata JSONB,                       -- Custom metadata
    frequency VARCHAR(50),                -- once, daily, weekly, monthly
    last_shown_at TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### alert_actions
```sql
CREATE TABLE alert_actions (
    id SERIAL PRIMARY KEY,
    alert_id INTEGER,
    label VARCHAR(100),                   -- Button/Link text
    action_url VARCHAR(500),              -- Navigation URL
    action_type VARCHAR(50),              -- navigate, api_call, dismiss
    payload JSONB,                        -- Additional data
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### seller_alert_logs
```sql
CREATE TABLE seller_alert_logs (
    id SERIAL PRIMARY KEY,
    seller_id INTEGER,
    alert_key VARCHAR(100),
    action VARCHAR(50),                   -- shown, dismissed, clicked
    action_url VARCHAR(500),
    metadata JSONB,
    created_at TIMESTAMP
);
```

---

## Sample Alert Templates Included

### 1. db_high_product_price
- **Triggers when:** Product count > 5
- **Action:** Review Pricing
- **Frequency:** Weekly (max 1 per week)

### 2. db_low_sales_velocity
- **Triggers when:** Product count > 0
- **Action:** View Analytics
- **Frequency:** Daily (max 1 per day)

### 3. db_pending_verification
- **Triggers when:** Shop not verified
- **Actions:** Check Status, Resubmit
- **Frequency:** Weekly

### 4. db_complete_profile
- **Triggers when:** Shop name empty
- **Action:** Edit Profile
- **Frequency:** Weekly

### 5. db_update_shop_photo
- **Triggers when:** No shop photo
- **Action:** Upload Photo
- **Frequency:** Daily

### 6. db_inventory_status
- **Triggers when:** Has products
- **Action:** Manage Inventory
- **Frequency:** Weekly

### 7. db_negative_feedback
- **Triggers when:** Has products
- **Action:** View Feedback
- **Frequency:** Daily

### 8. db_promotion_available
- **Triggers when:** Shop verified
- **Action:** View Promotions
- **Frequency:** Weekly

---

## Adding New Alert Templates

You can add new templates directly via SQL:

```sql
INSERT INTO alert_templates (key, title, description, type, priority, is_active, enabled, conditions, actions, frequency_config)
VALUES (
    'db_my_custom_alert',                           -- Unique key
    'My Custom Alert',                              -- Title
    'This is a custom alert template',              -- Description
    'warning',                                      -- Type: info, warning, critical
    1,                                              -- Priority (0-10, higher = more important)
    true,                                           -- Is active in UI
    true,                                           -- Is enabled for evaluation
    '{"conditions": [{"field": "product_count", "operator": ">", "value": 0}], "logic": "AND"}',
    '[{"label": "Action Button", "action_type": "navigate", "action_url": "/seller/action"}]',
    '{"type": "daily", "limit": 1}'                 -- Frequency config
);
```

---

## API Endpoints

### Get Seller Alerts
```bash
GET /api/admin/seller/alerts
Authorization: Bearer {jwt_token}

Response:
{
  "success": true,
  "message": "Alerts fetched successfully",
  "data": [
    {
      "id": 1,
      "key": "no_products_added",
      "title": "Start Selling",
      "content": "You haven't added any products yet...",
      "type": "critical",
      "priority": 3,
      "actions": [
        {
          "label": "Add Product",
          "action_type": "navigate",
          "action_url": "/seller/products/add"
        }
      ]
    }
  ]
}
```

### Dismiss Alert
```bash
POST /api/admin/seller/alerts/{alert_key}/dismiss
Authorization: Bearer {jwt_token}

Response:
{
  "success": true,
  "message": "Alert dismissed successfully"
}
```

---

## Verification

After setup, verify the tables were created:

```sql
-- Check if tables exist
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name LIKE 'alert%' OR table_name = 'seller_alert_logs';

-- Check sample data
SELECT key, title, type, priority FROM alert_templates;
```

---

## Troubleshooting

### Error: "relation alert_templates does not exist"
- Run the setup script: `ALERT_SYSTEM_SETUP.sql`
- Make sure your database user has CREATE TABLE privileges

### Templates not showing up
- Check if `is_active = true` and `enabled = true` in the database
- Verify conditions are valid JSON

### Alert not triggering
- Check condition syntax in `alert_templates.conditions`
- Verify seller has required data (shop, products, etc.)
- Check `valid_from` and `valid_until` timestamps

---

## Files Created

1. **ALERT_SYSTEM_SETUP.sql** - Complete setup script with sample data
2. **db/migrations/006_alert_system.sql** - Migration file format
3. **docker/postgres/initdb/06_alert_system.sql** - Docker initialization script

All three files contain the same schema. Choose whichever is most convenient for your deployment.
