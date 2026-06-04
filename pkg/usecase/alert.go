package usecase

import (
	"context"
	"errors"
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
	fmt.Printf("Fetching alerts for seller ID %s\n", sellerID)

	// Fetch aggregated data with minimal queries
	aggregatedData, err := uc.alertRepo.GetAggregatedDataForSeller(ctx, sellerID)
	if err != nil {
		// Seller has no shop yet — return empty alerts rather than 500
		if errors.Is(err, ErrShopNotFound) {
			return []*domain.Alert{}, nil
		}
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

	// Collect all alert keys first so we can fetch show-times in one query
	keys := make([]string, 0, len(alerts))
	for _, alert := range alerts {
		keys = append(keys, alert.Key)
	}
	lastShownTimes, _ := uc.alertRepo.GetLastAlertActionTimes(ctx, sellerID, keys)

	// Filter alerts by validity and frequency
	validAlerts := make([]*domain.Alert, 0)
	for _, alert := range alerts {
		if !alert_engine.IsAlertValid(alert) {
			continue
		}
		if !alert_engine.ShouldShowAlert(alert, lastShownTimes[alert.Key]) {
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
	log := &domain.SellerAlertLog{
		SellerID: sellerID,
		AlertKey: alertKey,
		Action:   "dismissed",
	}

	return uc.alertRepo.LogAlertAction(ctx, log)
}

// LogAlertView logs when alert is shown to seller
func (uc *AlertUseCaseImpl) LogAlertView(ctx context.Context, sellerID string, alertKey string) error {
	log := &domain.SellerAlertLog{
		SellerID: sellerID,
		AlertKey: alertKey,
		Action:   "shown",
	}

	return uc.alertRepo.LogAlertAction(ctx, log)
}
