package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoInterfaces "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
)

type stubCustomerNudgeRepo struct {
	repoInterfaces.CustomerNudgeRepository
	settings   domain.CustomerNudgeSettings
	templates  []domain.CustomerNudgeTemplate
	candidates []domain.CustomerNudgeCandidate
	sentSlots  map[string]map[[2]int]bool
	since      time.Time
	marked     []bool
}

func (s *stubCustomerNudgeRepo) GetSettings(context.Context) (domain.CustomerNudgeSettings, error) {
	return s.settings, nil
}
func (s *stubCustomerNudgeRepo) GetTemplates(context.Context) ([]domain.CustomerNudgeTemplate, error) {
	return s.templates, nil
}
func (s *stubCustomerNudgeRepo) Candidates(_ context.Context, since, _ time.Time) ([]domain.CustomerNudgeCandidate, error) {
	s.since = since
	return s.candidates, nil
}
func (s *stubCustomerNudgeRepo) SentSlotsByUser(context.Context, []string) (map[string]map[[2]int]bool, error) {
	if s.sentSlots == nil {
		return map[string]map[[2]int]bool{}, nil
	}
	return s.sentSlots, nil
}
func (s *stubCustomerNudgeRepo) MarkSent(_ context.Context, _ string, _, _ int, delivered bool) error {
	s.marked = append(s.marked, delivered)
	return nil
}

func customerTemplates() []domain.CustomerNudgeTemplate {
	return []domain.CustomerNudgeTemplate{
		{Key: domain.CustomerNudgeExploreShops, Title: "Explore", Body: "Hi {{name}}"},
		{Key: domain.CustomerNudgeSetLocation, Title: "Location", Body: "b"},
		{Key: domain.CustomerNudgeSendEnquiry, Title: "Enquiry", Body: "b"},
		{Key: domain.CustomerNudgeShareApp, Title: "Share", Body: "b"},
	}
}

func TestCustomerSweep_DisabledSendsNothing(t *testing.T) {
	repo := &stubCustomerNudgeRepo{settings: domain.CustomerNudgeSettings{Enabled: false}}
	notif := &stubNotificationUC{}
	res, err := NewCustomerNudgeUseCase(repo, notif).RunSweep(context.Background())
	if err != nil || !res.Disabled || len(notif.sent) != 0 {
		t.Fatalf("expected disabled no-op, got res=%+v err=%v sent=%d", res, err, len(notif.sent))
	}
}

func TestCustomerSweep_SendsFirstSlotWithNameToUserOwner(t *testing.T) {
	repo := &stubCustomerNudgeRepo{
		settings:   domain.CustomerNudgeSettings{Enabled: true, GapHours: 2, DurationDays: 7},
		templates:  customerTemplates(),
		candidates: []domain.CustomerNudgeCandidate{{UserID: "u1", FirstName: "Asha", StartAt: time.Now().Add(-time.Minute)}},
	}
	notif := &stubNotificationUC{}
	res, err := NewCustomerNudgeUseCase(repo, notif).RunSweep(context.Background())
	if err != nil || res.Sent != 1 {
		t.Fatalf("expected 1 send, got res=%+v err=%v", res, err)
	}
	got := notif.sent[0]
	if got.OwnerType != "user" || got.OwnerID != "u1" || got.Body != "Hi Asha" {
		t.Errorf("unexpected push: %+v", got)
	}
}

func TestCustomerSweep_SkipsAlreadySent(t *testing.T) {
	repo := &stubCustomerNudgeRepo{
		settings:   domain.CustomerNudgeSettings{Enabled: true, GapHours: 2, DurationDays: 7},
		templates:  customerTemplates(),
		candidates: []domain.CustomerNudgeCandidate{{UserID: "u1", StartAt: time.Now().Add(-time.Minute)}},
		sentSlots:  map[string]map[[2]int]bool{"u1": {{1, 0}: true}},
	}
	notif := &stubNotificationUC{}
	res, _ := NewCustomerNudgeUseCase(repo, notif).RunSweep(context.Background())
	if res.Sent != 0 || len(notif.sent) != 0 {
		t.Errorf("expected no sends, got %+v", res)
	}
}

func TestCustomerSweep_WindowNeverPredatesStartsAt(t *testing.T) {
	startsAt := time.Now().Add(-1 * time.Hour)
	repo := &stubCustomerNudgeRepo{
		settings:   domain.CustomerNudgeSettings{Enabled: true, GapHours: 2, DurationDays: 7, StartsAt: startsAt},
		templates:  customerTemplates(),
		candidates: []domain.CustomerNudgeCandidate{{UserID: "u1", StartAt: time.Now()}},
	}
	NewCustomerNudgeUseCase(repo, &stubNotificationUC{}).RunSweep(context.Background())
	if !repo.since.Equal(startsAt) {
		t.Errorf("candidate window starts %v, want StartsAt %v — existing customers must be excluded", repo.since, startsAt)
	}
}
