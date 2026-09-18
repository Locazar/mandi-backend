package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoInterfaces "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	usecaseInterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

// stubOnboardingNudgeRepo is an in-memory stand-in — only the methods
// RunSweep actually calls are exercised.
type stubOnboardingNudgeRepo struct {
	repoInterfaces.OnboardingNudgeRepository
	settings   domain.OnboardingNudgeSettings
	templates  []domain.OnboardingNudgeTemplate
	candidates []domain.OnboardingNudgeCandidate
	sentSlots  map[string]map[[2]int]bool
	marked     []([3]interface{}) // shopID, day, slot
}

func (s *stubOnboardingNudgeRepo) GetSettings(_ context.Context) (domain.OnboardingNudgeSettings, error) {
	return s.settings, nil
}

func (s *stubOnboardingNudgeRepo) GetTemplates(_ context.Context) ([]domain.OnboardingNudgeTemplate, error) {
	return s.templates, nil
}

func (s *stubOnboardingNudgeRepo) ActiveCandidates(_ context.Context, _ time.Time, _ int) ([]domain.OnboardingNudgeCandidate, error) {
	return s.candidates, nil
}

func (s *stubOnboardingNudgeRepo) SentSlotsByShop(_ context.Context, _ []string) (map[string]map[[2]int]bool, error) {
	if s.sentSlots == nil {
		return map[string]map[[2]int]bool{}, nil
	}
	return s.sentSlots, nil
}

func (s *stubOnboardingNudgeRepo) MarkSent(_ context.Context, shopID string, day, slot int) error {
	s.marked = append(s.marked, [3]interface{}{shopID, day, slot})
	return nil
}

// stubNotificationUC captures every SendPushNotification call.
type stubNotificationUC struct {
	usecaseInterfaces.NotificationUseCase
	sent []request.SendPushRequest
	err  error
}

func (s *stubNotificationUC) SendPushNotification(_ context.Context, req request.SendPushRequest) (bool, error) {
	s.sent = append(s.sent, req)
	return s.err == nil, s.err
}

func allFourTemplates() []domain.OnboardingNudgeTemplate {
	return []domain.OnboardingNudgeTemplate{
		{Key: domain.NudgeAddProducts, Title: "Add products", Body: "add some products"},
		{Key: domain.NudgeUpdatePhoto, Title: "Add photo", Body: "add a photo"},
		{Key: domain.NudgeUpdateAddress, Title: "Confirm address", Body: "confirm your address"},
		{Key: domain.NudgeViewShop, Title: "View shop", Body: "here: {{shop_link}}"},
	}
}

func TestRunSweep_Disabled_SendsNothing(t *testing.T) {
	repo := &stubOnboardingNudgeRepo{settings: domain.OnboardingNudgeSettings{Enabled: false}}
	notif := &stubNotificationUC{}
	uc := NewOnboardingNudgeUseCase(repo, notif)

	result, err := uc.RunSweep(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Disabled {
		t.Error("expected Disabled=true")
	}
	if len(notif.sent) != 0 {
		t.Errorf("expected no sends while disabled, got %d", len(notif.sent))
	}
}

func TestRunSweep_SendsFirstSlotImmediatelyAtGoLive(t *testing.T) {
	now := time.Now()
	repo := &stubOnboardingNudgeRepo{
		settings:  domain.OnboardingNudgeSettings{Enabled: true, GapHours: 2, DurationDays: 7},
		templates: allFourTemplates(),
		candidates: []domain.OnboardingNudgeCandidate{
			{ShopID: "shp_1", GoLiveAt: now.Add(-1 * time.Minute), ShopName: "Test Shop", City: "Jaipur"},
		},
	}
	notif := &stubNotificationUC{}
	uc := NewOnboardingNudgeUseCase(repo, notif)

	result, err := uc.RunSweep(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only day-1 slot-0 (add_products) is due — go-live was 1 minute ago,
	// slot 1 isn't due for another ~2 hours.
	if result.Sent != 1 {
		t.Fatalf("expected exactly 1 send, got %d (marked=%v)", result.Sent, repo.marked)
	}
	if notif.sent[0].Title != "Add products" {
		t.Errorf("expected the add_products template to fire first, got %q", notif.sent[0].Title)
	}
	if notif.sent[0].OwnerType != "seller" || notif.sent[0].OwnerID != "shp_1" {
		t.Errorf("unexpected target: %+v", notif.sent[0])
	}
}

func TestRunSweep_SkipsAlreadySentSlots(t *testing.T) {
	now := time.Now()
	repo := &stubOnboardingNudgeRepo{
		settings:  domain.OnboardingNudgeSettings{Enabled: true, GapHours: 2, DurationDays: 7},
		templates: allFourTemplates(),
		candidates: []domain.OnboardingNudgeCandidate{
			{ShopID: "shp_1", GoLiveAt: now.Add(-1 * time.Minute)},
		},
		sentSlots: map[string]map[[2]int]bool{
			"shp_1": {{1, 0}: true}, // day 1, slot 0 already sent
		},
	}
	notif := &stubNotificationUC{}
	uc := NewOnboardingNudgeUseCase(repo, notif)

	result, err := uc.RunSweep(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Sent != 0 {
		t.Errorf("expected 0 sends (only due slot already sent), got %d", result.Sent)
	}
}

func TestRunSweep_CatchesUpMultipleDueSlotsInOneRun(t *testing.T) {
	now := time.Now()
	repo := &stubOnboardingNudgeRepo{
		settings:  domain.OnboardingNudgeSettings{Enabled: true, GapHours: 2, DurationDays: 7},
		templates: allFourTemplates(),
		candidates: []domain.OnboardingNudgeCandidate{
			// Went live 5 hours ago: slots at +0h, +2h, +4h are all due (3 of 4 today).
			{ShopID: "shp_1", GoLiveAt: now.Add(-5 * time.Hour), ShopName: "Test Shop"},
		},
	}
	notif := &stubNotificationUC{}
	uc := NewOnboardingNudgeUseCase(repo, notif)

	result, err := uc.RunSweep(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Sent != 3 {
		t.Fatalf("expected 3 sends (slots 0,1,2 due), got %d", result.Sent)
	}
}

func TestRunSweep_InterpolatesShopLink(t *testing.T) {
	now := time.Now()
	// NudgeTemplateOrder puts view_shop at slot 3, due at go_live + 3*GapHours
	// — go live 4 hours ago with a 1-hour gap so that slot is due.
	repo := &stubOnboardingNudgeRepo{
		settings:  domain.OnboardingNudgeSettings{Enabled: true, GapHours: 1, DurationDays: 1},
		templates: []domain.OnboardingNudgeTemplate{{Key: domain.NudgeViewShop, Title: "View", Body: "Link: {{shop_link}}"}},
		candidates: []domain.OnboardingNudgeCandidate{
			{ShopID: "shp_1", GoLiveAt: now.Add(-4 * time.Hour), ShopName: "Fashion ForU", City: "Banglore"},
		},
	}
	notif := &stubNotificationUC{}
	uc := NewOnboardingNudgeUseCase(repo, notif)

	result, err := uc.RunSweep(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Sent != 1 {
		t.Fatalf("expected 1 send (view_shop due), got %d", result.Sent)
	}
	want := "Link: https://locazar.in/shop/fashion-foru-banglore"
	if notif.sent[0].Body != want {
		t.Errorf("body = %q, want %q", notif.sent[0].Body, want)
	}
}

func TestRunSweep_ForwardsTemplateRouteAsDataRoute(t *testing.T) {
	now := time.Now()
	repo := &stubOnboardingNudgeRepo{
		settings: domain.OnboardingNudgeSettings{Enabled: true, GapHours: 2, DurationDays: 7},
		templates: []domain.OnboardingNudgeTemplate{
			{Key: domain.NudgeAddProducts, Title: "Add products", Body: "add some products", Route: "/home?tab=2"},
		},
		candidates: []domain.OnboardingNudgeCandidate{
			{ShopID: "shp_1", GoLiveAt: now.Add(-1 * time.Minute)},
		},
	}
	notif := &stubNotificationUC{}
	uc := NewOnboardingNudgeUseCase(repo, notif)

	result, err := uc.RunSweep(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Sent != 1 {
		t.Fatalf("expected 1 send, got %d", result.Sent)
	}
	if got := notif.sent[0].Data["route"]; got != "/home?tab=2" {
		t.Errorf("data.route = %q, want %q", got, "/home?tab=2")
	}
}

func TestRunSweep_OutsideWindow_NoCandidates(t *testing.T) {
	repo := &stubOnboardingNudgeRepo{
		settings:  domain.OnboardingNudgeSettings{Enabled: true, GapHours: 2, DurationDays: 7},
		templates: allFourTemplates(),
		// ActiveCandidates is the repo's job to filter by window — simulate
		// "nothing in window" by returning an empty slice.
		candidates: nil,
	}
	notif := &stubNotificationUC{}
	uc := NewOnboardingNudgeUseCase(repo, notif)

	result, err := uc.RunSweep(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Sent != 0 {
		t.Errorf("expected 0 sends with no candidates, got %d", result.Sent)
	}
}
