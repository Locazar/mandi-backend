// Package alert_engine implements a rule-based alert system for seller shops.
// It evaluates seller data against registered rules and returns actionable alerts.
//
// Architecture:
//   - AlertRule interface: implement Key() + Evaluate() to create a new rule.
//   - RuleRegistry: thread-safe map of registered rules; managed via Register/RegisterMultiple.
//   - AlertEvaluator: runs all registered rules against a seller's AggregatedData and returns triggered alerts sorted by priority.
//
// To add a new rule:
//  1. Create a struct in rules.go implementing AlertRule (Key() string + Evaluate(...) (*domain.Alert, error)).
//  2. Register it in pkg/di/wire.go inside the NewAlertRuleRegistry factory.
//  3. No other changes required — EvaluateAll picks it up automatically.
//
// Existing rules (rules.go): MissingShopPhotoRule, NoProductsRule, ShopNotVerifiedRule.
package alert_engine

import (
	"context"
	"sync"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// AlertRule defines the interface for all alert rules
type AlertRule interface {
	// Key returns a unique identifier for this rule
	Key() string
	// Evaluate checks if this rule applies and returns an alert if it does
	Evaluate(ctx context.Context, sellerID string, data domain.AggregatedData) (*domain.Alert, error)
}

// RuleRegistry manages all available alert rules
type RuleRegistry struct {
	rules map[string]AlertRule
	mu    sync.RWMutex
}

// NewRuleRegistry creates a new rule registry
func NewRuleRegistry() *RuleRegistry {
	return &RuleRegistry{
		rules: make(map[string]AlertRule),
	}
}

// Register adds a new rule to the registry
func (r *RuleRegistry) Register(rule AlertRule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules[rule.Key()] = rule
}

// RegisterMultiple adds multiple rules to the registry
func (r *RuleRegistry) RegisterMultiple(rules ...AlertRule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, rule := range rules {
		r.rules[rule.Key()] = rule
	}
}

// GetAll returns all registered rules
func (r *RuleRegistry) GetAll() []AlertRule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	rules := make([]AlertRule, 0, len(r.rules))
	for _, rule := range r.rules {
		rules = append(rules, rule)
	}
	return rules
}

// Get retrieves a specific rule by key
func (r *RuleRegistry) Get(key string) (AlertRule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rule, exists := r.rules[key]
	return rule, exists
}

// Remove removes a rule from the registry
func (r *RuleRegistry) Remove(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.rules, key)
}

// AlertEvaluator coordinates rule evaluation
type AlertEvaluator struct {
	registry *RuleRegistry
}

// NewAlertEvaluator creates a new alert evaluator
func NewAlertEvaluator(registry *RuleRegistry) *AlertEvaluator {
	return &AlertEvaluator{
		registry: registry,
	}
}

// EvaluateAll evaluates all registered rules and returns triggered alerts
func (e *AlertEvaluator) EvaluateAll(ctx context.Context, sellerID string, data domain.AggregatedData) ([]*domain.Alert, error) {
	alerts := make([]*domain.Alert, 0)
	rules := e.registry.GetAll()

	for _, rule := range rules {
		alert, err := rule.Evaluate(ctx, sellerID, data)
		if err != nil {
			// Log error but continue with other rules
			continue
		}
		if alert != nil {
			alerts = append(alerts, alert)
		}
	}

	// Sort by priority (higher priority first)
	sortAlertsByPriority(alerts)
	return alerts, nil
}

// sortAlertsByPriority sorts alerts by priority in descending order
func sortAlertsByPriority(alerts []*domain.Alert) {
	for i := 0; i < len(alerts)-1; i++ {
		for j := 0; j < len(alerts)-i-1; j++ {
			if alerts[j].Priority < alerts[j+1].Priority {
				alerts[j], alerts[j+1] = alerts[j+1], alerts[j]
			}
		}
	}
}

// IsAlertValid checks if alert is within its validity window
func IsAlertValid(alert *domain.Alert) bool {
	now := time.Now()
	
	if alert.ValidFrom != nil && now.Before(*alert.ValidFrom) {
		return false
	}
	
	if alert.ValidUntil != nil && now.After(*alert.ValidUntil) {
		return false
	}
	
	return alert.IsActive
}

// ShouldShowAlert determines if alert should be shown based on frequency and last shown time
func ShouldShowAlert(alert *domain.Alert, lastShownAt *time.Time) bool {
	if lastShownAt == nil {
		return true // First time showing
	}

	switch alert.Frequency {
	case "once":
		return false // Only show once ever
	case "daily":
		return time.Since(*lastShownAt) >= 24*time.Hour
	case "weekly":
		return time.Since(*lastShownAt) >= 7*24*time.Hour
	case "monthly":
		return time.Since(*lastShownAt) >= 30*24*time.Hour
	default:
		return true // No frequency limit
	}
}
