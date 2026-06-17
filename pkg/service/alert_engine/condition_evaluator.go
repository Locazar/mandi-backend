package alert_engine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// Condition represents a single condition to evaluate
type Condition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// Conditions represents a group of conditions (AND logic)
type Conditions struct {
	Conditions []Condition `json:"conditions"`
	Logic      string      `json:"logic"` // "AND" or "OR"
}

// ConditionEvaluator evaluates JSON-based conditions
type ConditionEvaluator struct{}

// NewConditionEvaluator creates a new condition evaluator
func NewConditionEvaluator() *ConditionEvaluator {
	return &ConditionEvaluator{}
}

// EvaluateConditions evaluates a set of conditions against aggregated data
func (e *ConditionEvaluator) EvaluateConditions(data domain.AggregatedData, conditionsJSON json.RawMessage) (bool, error) {
	if len(conditionsJSON) == 0 {
		return true, nil // No conditions = always true
	}

	// Parse conditions
	var conditions Conditions
	if err := json.Unmarshal(conditionsJSON, &conditions); err != nil {
		return false, fmt.Errorf("failed to parse conditions: %w", err)
	}

	if conditions.Logic == "" {
		conditions.Logic = "AND"
	}

	// Evaluate each condition
	results := make([]bool, 0, len(conditions.Conditions))
	for _, cond := range conditions.Conditions {
		result, err := e.evaluateCondition(data, cond)
		if err != nil {
			return false, err
		}
		results = append(results, result)
	}

	// Apply logic
	if conditions.Logic == "AND" {
		for _, result := range results {
			if !result {
				return false, nil
			}
		}
		return true, nil
	} else if conditions.Logic == "OR" {
		for _, result := range results {
			if result {
				return true, nil
			}
		}
		return false, nil
	}

	return false, fmt.Errorf("unknown logic: %s", conditions.Logic)
}

// evaluateCondition evaluates a single condition
func (e *ConditionEvaluator) evaluateCondition(data domain.AggregatedData, cond Condition) (bool, error) {
	var fieldValue interface{}

	// Extract field value from aggregated data
	switch cond.Field {
	case "product_count":
		fieldValue = data.ProductCount
	case "is_verified":
		fieldValue = data.IsVerified
	case "has_shop_photo":
		fieldValue = data.HasShopPhoto
	case "shop_name":
		fieldValue = data.Shop.ShopName
	case "shop_status":
		fieldValue = data.Shop.ShopStatus
	case "admin_id":
		fieldValue = data.AdminID
	case "shop_id":
		fieldValue = data.ShopID
	default:
		return false, fmt.Errorf("unknown field: %s", cond.Field)
	}

	// Apply operator
	return e.applyOperator(fieldValue, cond.Operator, cond.Value)
}

// applyOperator applies a comparison operator
func (e *ConditionEvaluator) applyOperator(fieldValue interface{}, operator string, expectedValue interface{}) (bool, error) {
	switch operator {
	case "=", "equals":
		return fieldValue == expectedValue, nil
	case "!=", "not_equals":
		return fieldValue != expectedValue, nil
	case ">", "greater_than":
		return e.compareNumeric(fieldValue, expectedValue, func(a, b float64) bool { return a > b })
	case "<", "less_than":
		return e.compareNumeric(fieldValue, expectedValue, func(a, b float64) bool { return a < b })
	case ">=", "greater_than_or_equal":
		return e.compareNumeric(fieldValue, expectedValue, func(a, b float64) bool { return a >= b })
	case "<=", "less_than_or_equal":
		return e.compareNumeric(fieldValue, expectedValue, func(a, b float64) bool { return a <= b })
	case "in":
		return e.inArray(fieldValue, expectedValue)
	case "not_in":
		result, _ := e.inArray(fieldValue, expectedValue)
		return !result, nil
	case "contains":
		return e.stringContains(fieldValue, expectedValue)
	default:
		return false, fmt.Errorf("unknown operator: %s", operator)
	}
}

// compareNumeric compares two numeric values
func (e *ConditionEvaluator) compareNumeric(a, b interface{}, compareFn func(float64, float64) bool) (bool, error) {
	aVal, aOk := toFloat64(a)
	bVal, bOk := toFloat64(b)
	if !aOk || !bOk {
		return false, fmt.Errorf("cannot compare non-numeric values")
	}
	return compareFn(aVal, bVal), nil
}

// inArray checks if value is in array
func (e *ConditionEvaluator) inArray(fieldValue interface{}, expectedValue interface{}) (bool, error) {
	arr, ok := expectedValue.([]interface{})
	if !ok {
		return false, fmt.Errorf("expected array for 'in' operator")
	}
	for _, v := range arr {
		if fieldValue == v {
			return true, nil
		}
	}
	return false, nil
}

// stringContains checks if string contains substring
func (e *ConditionEvaluator) stringContains(fieldValue interface{}, expectedValue interface{}) (bool, error) {
	field, ok := fieldValue.(string)
	if !ok {
		return false, nil
	}
	substring, ok := expectedValue.(string)
	if !ok {
		return false, nil
	}
	return contains(field, substring), nil
}

// toFloat64 converts interface to float64
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint64:
		return float64(val), true
	default:
		return 0, false
	}
}

// contains is a helper for string contains
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// DBDrivenAlertRule represents a dynamically generated rule from DB
type DBDrivenAlertRule struct {
	Template  domain.AlertTemplate
	Evaluator *ConditionEvaluator
}

// NewDBDrivenAlertRule creates a new database-driven rule
func NewDBDrivenAlertRule(template domain.AlertTemplate) *DBDrivenAlertRule {
	return &DBDrivenAlertRule{
		Template:  template,
		Evaluator: NewConditionEvaluator(),
	}
}

// Key returns the template key
func (r *DBDrivenAlertRule) Key() string {
	return "db_" + r.Template.Key
}

// Evaluate evaluates the database-driven rule
func (r *DBDrivenAlertRule) Evaluate(ctx context.Context, sellerID string, data domain.AggregatedData) (*domain.Alert, error) {
	// Check if template is enabled
	if !r.Template.Enabled || !r.Template.IsActive {
		return nil, nil
	}

	// Evaluate conditions
	conditionsMet, err := r.Evaluator.EvaluateConditions(data, r.Template.Conditions)
	if err != nil || !conditionsMet {
		return nil, err
	}

	// Build alert from template
	alert := &domain.Alert{
		SellerID:    data.AdminID,
		Key:         r.Template.Key,
		Title:       r.Template.Title,
		Description: r.Template.Description,
		Type:        domain.AlertType(r.Template.Type),
		Priority:    r.Template.Priority,
		IsActive:    r.Template.IsActive,
	}

	// Parse actions from template JSON
	if r.Template.Actions != nil {
		var rawActions []map[string]interface{}
		if err := json.Unmarshal(r.Template.Actions, &rawActions); err == nil {
			actions := make([]domain.AlertAction, 0, len(rawActions))
			for _, a := range rawActions {
				action := domain.AlertAction{}
				if v, ok := a["label"].(string); ok {
					action.Label = v
				}
				if v, ok := a["action_type"].(string); ok {
					action.ActionType = domain.AlertActionType(v)
				}
				if v, ok := a["action_url"].(string); ok {
					action.ActionURL = v
				}
				actions = append(actions, action)
			}
			alert.Actions = actions
		} else {
			alert.Actions = make([]domain.AlertAction, 0)
		}
	}

	// Parse frequency config
	var freqMap map[string]interface{}
	if err := json.Unmarshal(r.Template.FrequencyConfig, &freqMap); err == nil {
		if freqType, ok := freqMap["type"].(string); ok {
			alert.Frequency = freqType
		}
	}

	return alert, nil
}
