package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"gorm.io/gorm"
)

type alertTemplateUseCase struct {
	repo repo.AlertRepository
}

func NewAlertTemplateUseCase(repo repo.AlertRepository) interfaces.AlertTemplateUseCase {
	return &alertTemplateUseCase{repo: repo}
}

// ListTemplates returns all alert templates as maps
func (uc *alertTemplateUseCase) ListTemplates(ctx context.Context) ([]map[string]interface{}, error) {
	templates, err := uc.repo.GetAllAlertTemplates(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(templates))
	for _, t := range templates {
		b, err := json.Marshal(t)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal template %q: %w", t.Key, err)
		}
		var m map[string]interface{}
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, fmt.Errorf("failed to unmarshal template %q: %w", t.Key, err)
		}
		result = append(result, m)
	}
	return result, nil
}

// CreateTemplate creates a new alert template from a request map
func (uc *alertTemplateUseCase) CreateTemplate(ctx context.Context, req map[string]interface{}) error {
	t := domain.AlertTemplate{}

	if v, ok := req["key"].(string); ok {
		t.Key = v
	}
	if v, ok := req["title"].(string); ok {
		t.Title = v
	}
	if v, ok := req["description"].(string); ok {
		t.Description = v
	}
	if v, ok := req["type"].(string); ok {
		t.Type = v
	}
	if v, ok := req["template_type"].(string); ok {
		t.TemplateType = v
	}
	if v, ok := req["display_type"].(string); ok {
		t.DisplayType = domain.AlertDisplayType(v)
	}
	if v, ok := req["flow_key"].(string); ok {
		t.FlowKey = v
	}
	if v, ok := req["priority"].(float64); ok {
		t.Priority = int(v)
	}
	if v, ok := req["enabled"].(bool); ok {
		t.Enabled = v
	}
	if v, ok := req["is_active"].(bool); ok {
		t.IsActive = v
	}

	if v, ok := req["content_schema"]; ok && v != nil {
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("invalid content_schema: %w", err)
		}
		t.ContentSchema = json.RawMessage(b)
	}
	if v, ok := req["conditions"]; ok && v != nil {
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("invalid conditions: %w", err)
		}
		t.Conditions = json.RawMessage(b)
	}
	if v, ok := req["actions"]; ok && v != nil {
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("invalid actions: %w", err)
		}
		t.Actions = json.RawMessage(b)
	}
	if v, ok := req["frequency_config"]; ok && v != nil {
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("invalid frequency_config: %w", err)
		}
		t.FrequencyConfig = json.RawMessage(b)
	}

	return uc.repo.SaveAlertTemplate(ctx, &t)
}

// UpdateTemplate applies partial updates from req to the existing template identified by key
func (uc *alertTemplateUseCase) UpdateTemplate(ctx context.Context, key string, req map[string]interface{}) error {
	t, err := uc.repo.GetAlertTemplateByKey(ctx, key)
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	if v, ok := req["title"].(string); ok && v != "" {
		t.Title = v
	}
	if v, ok := req["description"].(string); ok && v != "" {
		t.Description = v
	}
	if v, ok := req["type"].(string); ok && v != "" {
		t.Type = v
	}
	if v, ok := req["template_type"].(string); ok && v != "" {
		t.TemplateType = v
	}
	if v, ok := req["display_type"].(string); ok && v != "" {
		t.DisplayType = domain.AlertDisplayType(v)
	}
	if v, ok := req["flow_key"].(string); ok && v != "" {
		t.FlowKey = v
	}
	if v, ok := req["priority"].(float64); ok {
		t.Priority = int(v)
	}
	if v, ok := req["enabled"].(bool); ok {
		t.Enabled = v
	}
	if v, ok := req["is_active"].(bool); ok {
		t.IsActive = v
	}
	if v, ok := req["content_schema"]; ok && v != nil {
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("invalid content_schema: %w", err)
		}
		t.ContentSchema = json.RawMessage(b)
	}
	if v, ok := req["conditions"]; ok && v != nil {
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("invalid conditions: %w", err)
		}
		t.Conditions = json.RawMessage(b)
	}
	if v, ok := req["actions"]; ok && v != nil {
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("invalid actions: %w", err)
		}
		t.Actions = json.RawMessage(b)
	}
	if v, ok := req["frequency_config"]; ok && v != nil {
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("invalid frequency_config: %w", err)
		}
		t.FrequencyConfig = json.RawMessage(b)
	}

	return uc.repo.UpdateAlertTemplate(ctx, t)
}

// DeleteTemplate removes the alert template by key
func (uc *alertTemplateUseCase) DeleteTemplate(ctx context.Context, key string) error {
	return uc.repo.DeleteAlertTemplate(ctx, key)
}

// ToggleTemplate enables or disables a template by key
func (uc *alertTemplateUseCase) ToggleTemplate(ctx context.Context, key string, enabled bool) error {
	t, err := uc.repo.GetAlertTemplateByKey(ctx, key)
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}
	t.Enabled = enabled
	return uc.repo.UpdateAlertTemplate(ctx, t)
}

// SeedDefaults creates the four built-in templates if they don't already exist
func (uc *alertTemplateUseCase) SeedDefaults(ctx context.Context) error {
	seeds := []struct {
		key           string
		title         string
		description   string
		templateType  string
		displayType   string
		alertType     string
		priority      int
		contentSchema string
	}{
		{
			key:          "promo_welcome",
			title:        "Welcome Offer",
			description:  "Special promo for new sellers",
			templateType: "promo_card",
			displayType:  "bottom_sheet",
			alertType:    "promotion",
			priority:     5,
			contentSchema: `{"title":"Welcome to Locazar!","description":"Get 10% off your first order","badge_text":"New Seller","image_url":"","background_color":"#F0F9FF","text_color":"#1E40AF","actions":[{"label":"Claim Offer","action_type":"navigate","link":"/promotions"}]}`,
		},
		{
			key:          "onboarding_setup",
			title:        "Complete Your Setup",
			description:  "Guide seller through initial setup",
			templateType: "onboarding_step",
			displayType:  "bottom_sheet",
			alertType:    "action_required",
			priority:     10,
			contentSchema: `{"steps":[{"step_number":1,"title":"Add Your Shop Photo","content":{"title":"Add a Shop Photo","description":"Shops with photos get 3x more customers","icon":"photo","bullets":["Upload a clear storefront photo","Accepted formats: JPG, PNG"]},"actions":[{"label":"Upload Photo","action_type":"navigate","link":"/shop/photo"}]},{"step_number":2,"title":"Add Your First Product","content":{"title":"List Your First Product","description":"Start selling by adding a product","icon":"product","bullets":["Add name, price, and photos","Set stock quantity"]},"actions":[{"label":"Add Product","action_type":"navigate","link":"/products/add"}]}]}`,
		},
		{
			key:          "video_seller_guide",
			title:        "Seller Guide",
			description:  "Video guide for new sellers",
			templateType: "video",
			displayType:  "bottom_sheet",
			alertType:    "info",
			priority:     3,
			contentSchema: `{"title":"Getting Started","description":"Watch this short guide","video_url":"","video_aspect_ratio":1.78,"show_video_description":true}`,
		},
		{
			key:          "webview_help",
			title:        "Help Center",
			description:  "In-app help center",
			templateType: "webview",
			displayType:  "bottom_sheet",
			alertType:    "info",
			priority:     1,
			contentSchema: `{"title":"Help Center","url":"https://locazar.com/help","show_close_icon":true}`,
		},
	}

	for _, s := range seeds {
		_, err := uc.repo.GetAlertTemplateByKey(ctx, s.key)
		if err == nil {
			// Already exists — skip
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("error checking template %q: %w", s.key, err)
		}

		t := &domain.AlertTemplate{
			Key:           s.key,
			Title:         s.title,
			Description:   s.description,
			TemplateType:  s.templateType,
			DisplayType:   domain.AlertDisplayType(s.displayType),
			Type:          s.alertType,
			Priority:      s.priority,
			Enabled:       true,
			IsActive:      true,
			ContentSchema: json.RawMessage(s.contentSchema),
		}
		if err := uc.repo.SaveAlertTemplate(ctx, t); err != nil {
			return fmt.Errorf("failed to seed template %q: %w", s.key, err)
		}
	}
	return nil
}

// GetTemplateForSeller returns a seller-facing view of a template
func (uc *alertTemplateUseCase) GetTemplateForSeller(ctx context.Context, sellerID string, key string) (map[string]interface{}, error) {
	t, err := uc.repo.GetAlertTemplateByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}
	if !t.Enabled || !t.IsActive {
		return nil, fmt.Errorf("template %q is not active", key)
	}

	var content map[string]interface{}
	if len(t.ContentSchema) > 0 {
		if err := json.Unmarshal(t.ContentSchema, &content); err != nil {
			return nil, fmt.Errorf("invalid content_schema: %w", err)
		}
	}

	var actions interface{}
	if len(t.Actions) > 0 {
		if err := json.Unmarshal(t.Actions, &actions); err != nil {
			return nil, fmt.Errorf("invalid actions: %w", err)
		}
	}

	return map[string]interface{}{
		"id":            t.ID,
		"key":           t.Key,
		"title":         t.Title,
		"description":   t.Description,
		"type":          t.Type,
		"template_type": t.TemplateType,
		"display_type":  t.DisplayType,
		"flow_key":      t.FlowKey,
		"priority":      t.Priority,
		"content":       content,
		"actions":       actions,
	}, nil
}

// GetFlow returns the onboarding flow state for a seller
func (uc *alertTemplateUseCase) GetFlow(ctx context.Context, sellerID string, flowKey string) (map[string]interface{}, error) {
	t, err := uc.repo.GetAlertTemplateByKey(ctx, flowKey)
	if err != nil {
		return nil, fmt.Errorf("flow template not found: %w", err)
	}
	if t.TemplateType != "onboarding_step" {
		return nil, fmt.Errorf("template %q is not an onboarding_step (got %q)", flowKey, t.TemplateType)
	}

	if len(t.ContentSchema) == 0 {
		return nil, fmt.Errorf("template %q has no content_schema", flowKey)
	}
	var contentSchema map[string]interface{}
	if err := json.Unmarshal(t.ContentSchema, &contentSchema); err != nil {
		return nil, fmt.Errorf("invalid content_schema: %w", err)
	}

	stepsRaw, _ := contentSchema["steps"].([]interface{})
	steps := make([]map[string]interface{}, 0, len(stepsRaw))
	for _, s := range stepsRaw {
		if m, ok := s.(map[string]interface{}); ok {
			steps = append(steps, m)
		}
	}
	totalSteps := len(steps)

	completedSteps, err := uc.repo.GetStepCompletions(ctx, sellerID, flowKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get step completions: %w", err)
	}

	completedSet := make(map[int]bool, len(completedSteps))
	for _, n := range completedSteps {
		completedSet[n] = true
	}

	currentStep := totalSteps + 1
	for i := 1; i <= totalSteps; i++ {
		if !completedSet[i] {
			currentStep = i
			break
		}
	}

	return map[string]interface{}{
		"key":          t.Key,
		"title":        t.Title,
		"description":  t.Description,
		"total_steps":  totalSteps,
		"current_step": currentStep,
		"steps":        steps,
	}, nil
}

// CompleteStep logs a step completion for the given seller and flow
func (uc *alertTemplateUseCase) CompleteStep(ctx context.Context, sellerID string, flowKey string, stepNumber int) error {
	if stepNumber <= 0 {
		return fmt.Errorf("stepNumber must be >= 1, got %d", stepNumber)
	}
	log := &domain.SellerAlertLog{
		SellerID: sellerID,
		AlertKey: flowKey,
		Action:   "step_completed",
		Metadata: json.RawMessage(fmt.Sprintf(`{"step":%d}`, stepNumber)),
	}
	return uc.repo.LogAlertAction(ctx, log)
}
