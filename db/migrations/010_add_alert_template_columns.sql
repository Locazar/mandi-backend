-- Add missing columns to alert_templates table
ALTER TABLE alert_templates
  ADD COLUMN IF NOT EXISTS template_type VARCHAR(50) DEFAULT 'announcement',
  ADD COLUMN IF NOT EXISTS display_type VARCHAR(50) DEFAULT 'bottom_sheet',
  ADD COLUMN IF NOT EXISTS content_schema JSONB,
  ADD COLUMN IF NOT EXISTS flow_key VARCHAR(100);

-- Create index on flow_key for better query performance
CREATE INDEX IF NOT EXISTS idx_alert_templates_flow_key ON alert_templates(flow_key);
