package domain

import (
	"encoding/json"
	"time"
)

// Alert represents a seller alert
type Alert struct {
	ID          uint                   `json:"id" gorm:"primaryKey"`
	SellerID    uint                   `json:"seller_id" gorm:"not null;index"`
	Key         string                 `json:"key" gorm:"size:100;not null;index"`
	Title       string                 `json:"title" gorm:"size:255;not null"`
	Content     string                 `json:"content" gorm:"type:text"`
	Description string                 `json:"description" gorm:"type:text"`
	Type        string                 `json:"type" gorm:"size:50;default:'info'"` // info, warning, critical
	Priority    int                    `json:"priority" gorm:"default:0"`
	Actions     []AlertAction          `json:"actions" gorm:"foreignKey:AlertID;constraint:OnDelete:CASCADE"`
	IsActive    bool                   `json:"is_active" gorm:"default:true"`
	ValidFrom   *time.Time             `json:"valid_from" gorm:"index"`
	ValidUntil  *time.Time             `json:"valid_until" gorm:"index"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"` // Store custom data
	Frequency   string                 `json:"frequency" gorm:"size:50"`                // once, daily, weekly, etc.
	LastShownAt *time.Time             `json:"last_shown_at"`
	CreatedAt   time.Time              `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

// AlertAction represents an action that can be taken on an alert
type AlertAction struct {
	ID         uint            `json:"id" gorm:"primaryKey"`
	AlertID    uint            `json:"alert_id" gorm:"not null"`
	Label      string          `json:"label" gorm:"size:100;not null"`
	ActionURL  string          `json:"action_url" gorm:"size:500"`
	ActionType string          `json:"action_type" gorm:"size:50;not null"` // navigate, api_call, dismiss
	Payload    json.RawMessage `json:"payload" gorm:"type:jsonb;default:'{}'"`
	CreatedAt  time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

// AlertTemplate defines reusable alert configurations from DB
type AlertTemplate struct {
	ID              uint            `json:"id" gorm:"primaryKey"`
	Key             string          `json:"key" gorm:"size:100;not null;uniqueIndex"`
	Title           string          `json:"title" gorm:"size:255;not null"`
	Description     string          `json:"description" gorm:"type:text"`
	Type            string          `json:"type" gorm:"size:50;default:'info'"`
	Priority        int             `json:"priority" gorm:"default:0"`
	IsActive        bool            `json:"is_active" gorm:"default:true;index"`
	Conditions      json.RawMessage `json:"conditions" gorm:"type:jsonb"` // JSON condition parser
	Actions         json.RawMessage `json:"actions" gorm:"type:jsonb"`
	FrequencyConfig json.RawMessage `json:"frequency_config" gorm:"type:jsonb"` // {type: daily, limit: 1}
	Validity        json.RawMessage `json:"validity" gorm:"type:jsonb"`         // {valid_from, valid_until}
	Enabled         bool            `json:"enabled" gorm:"default:true"`
	TemplateType    string          `json:"template_type" gorm:"size:50;default:'announcement'"`
	DisplayType     string          `json:"display_type" gorm:"size:50;default:'bottom_sheet'"`
	ContentSchema   json.RawMessage `json:"content_schema" gorm:"type:jsonb"`
	FlowKey         string          `json:"flow_key" gorm:"size:100;index"`
	CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

// SellerAlertLog tracks alert dismissals and interactions
type SellerAlertLog struct {
	ID        uint            `json:"id" gorm:"primaryKey"`
	SellerID  uint            `json:"seller_id" gorm:"not null;index"`
	AlertKey  string          `json:"alert_key" gorm:"size:100;not null"`
	Action    string          `json:"action" gorm:"size:50;not null"` // shown, dismissed, clicked
	ActionURL string          `json:"action_url" gorm:"size:500"`
	Metadata  json.RawMessage `json:"metadata" gorm:"type:jsonb"`
	CreatedAt time.Time       `json:"created_at" gorm:"autoCreateTime;index"`
}

// AggregatedData contains pre-fetched data for rule evaluation
type AggregatedData struct {
	Shop         ShopDetails
	ProductCount int
	IsVerified   bool
	HasShopPhoto bool
	AdminID      uint
	ShopID       uint
	FetchedAt    time.Time
}

// AlertResponse is the API response structure for a single alert
type AlertResponse struct {
	ID          uint                   `json:"id"`
	Key         string                 `json:"key"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Priority    int                    `json:"priority"`
	Actions     []AlertActionResponse  `json:"actions"`
	IsActive    bool                   `json:"is_active"`
	ValidFrom   *time.Time             `json:"valid_from,omitempty"`
	ValidUntil  *time.Time             `json:"valid_until,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Frequency   string                 `json:"frequency,omitempty"`
}

// AlertActionResponse is the API response for alert actions
type AlertActionResponse struct {
	Label      string                 `json:"label"`
	ActionURL  string                 `json:"action_url,omitempty"`
	ActionType string                 `json:"action_type"`
	Payload    map[string]interface{} `json:"payload,omitempty"`
}

// TableName specifies the table name for AlertTemplate
func (AlertTemplate) TableName() string {
	return "alert_templates"
}

// TableName specifies the table name for SellerAlertLog
func (SellerAlertLog) TableName() string {
	return "seller_alert_logs"
}
