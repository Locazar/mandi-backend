package usecase

import (
	"context"
	"fmt"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/service/alert_engine"
	"github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

// AlertUseCaseImpl implements AlertUseCase
type AlertUseCaseImpl struct {
	alertRepo repo.AlertRepository
	registry  *alert_engine.RuleRegistry
	evaluator *alert_engine.AlertEvaluator
	condEval  *alert_engine.ConditionEvaluator
}

// NewAlertUseCase creates a new alert usecase
func NewAlertUseCase(
	alertRepo repo.AlertRepository,
	registry *alert_engine.RuleRegistry,
) interfaces.AlertUseCase {
	return &AlertUseCaseImpl{
		alertRepo: alertRepo,
		registry:  registry,
		evaluator: alert_engine.NewAlertEvaluator(registry),
		condEval:  alert_engine.NewConditionEvaluator(),
	}
}

// GetSellerAlerts fetches alerts for a seller by evaluating all rules
func (uc *AlertUseCaseImpl) GetSellerAlerts(ctx context.Context, sellerID string) ([]*domain.Alert, error) {
	// Parse seller ID to uint (assuming it comes as string from auth)
	var adminID uint
	if _, err := fmt.Sscanf(sellerID, "%d", &adminID); err != nil {
		return nil, fmt.Errorf("invalid seller_id format: %w", err)
	}
	fmt.Printf("Fetching alerts for seller ID %d\n", adminID)

	// Fetch aggregated data with minimal queries
	aggregatedData, err := uc.alertRepo.GetAggregatedDataForSeller(ctx, adminID)
	if err != nil {
		return nil, err
	}

	// Evaluate all code-based rules
	alerts, err := uc.evaluator.EvaluateAll(ctx, sellerID, *aggregatedData)
	if err != nil {
		return nil, err
	}

	// Fetch and evaluate database-driven rules
	dbAlerts, err := uc.evaluateDBDrivenRules(ctx, sellerID, *aggregatedData)
	if err != nil {
		// Log error but continue - DB rules are optional
	}
	alerts = append(alerts, dbAlerts...)

	// Filter alerts by validity and frequency
	validAlerts := make([]*domain.Alert, 0)
	for _, alert := range alerts {
		// Check validity window
		if !alert_engine.IsAlertValid(alert) {
			continue
		}

		// Check frequency - get last shown time
		lastShownAt, _ := uc.alertRepo.GetLastAlertActionTime(ctx, adminID, alert.Key)

		// Check if we should show this alert
		if !alert_engine.ShouldShowAlert(alert, lastShownAt) {
			continue
		}

		validAlerts = append(validAlerts, alert)
	}

	return validAlerts, nil
}

// evaluateDBDrivenRules evaluates rules from alert_templates table
func (uc *AlertUseCaseImpl) evaluateDBDrivenRules(ctx context.Context, sellerID string, data domain.AggregatedData) ([]*domain.Alert, error) {
	templates, err := uc.alertRepo.GetActiveAlertTemplates(ctx)
	if err != nil {
		return nil, err
	}

	alerts := make([]*domain.Alert, 0)
	for _, template := range templates {
		rule := alert_engine.NewDBDrivenAlertRule(*template)
		alert, err := rule.Evaluate(ctx, sellerID, data)
		if err != nil || alert == nil {
			continue
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// DismissAlert logs alert dismissal (prevents repeated display based on frequency)
func (uc *AlertUseCaseImpl) DismissAlert(ctx context.Context, sellerID string, alertKey string) error {
	var adminID uint
	if _, err := fmt.Sscanf(sellerID, "%d", &adminID); err != nil {
		return fmt.Errorf("invalid seller_id format: %w", err)
	}

	log := &domain.SellerAlertLog{
		SellerID: adminID,
		AlertKey: alertKey,
		Action:   "dismissed",
	}

	return uc.alertRepo.LogAlertAction(ctx, log)
}

// LogAlertView logs when alert is shown to seller
func (uc *AlertUseCaseImpl) LogAlertView(ctx context.Context, sellerID string, alertKey string) error {
	var adminID uint
	if _, err := fmt.Sscanf(sellerID, "%d", &adminID); err != nil {
		return fmt.Errorf("invalid seller_id format: %w", err)
	}

	log := &domain.SellerAlertLog{
		SellerID: adminID,
		AlertKey: alertKey,
		Action:   "shown",
	}

	return uc.alertRepo.LogAlertAction(ctx, log)
}
