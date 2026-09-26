package usecase

import (
	"context"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	service "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

// RewardUseCase holds every rule of the in-store rewards program. Anything
// that changes a balance runs inside repo.InTx on the transaction-bound repo.
type RewardUseCase struct {
	repo  repo.RewardRepository
	notif service.NotificationUseCase
	now   func() time.Time
	// dispatch runs fire-and-forget side effects (pushes). Tests replace it
	// with a synchronous runner.
	dispatch func(func())
}

func NewRewardUseCase(r repo.RewardRepository, notif service.NotificationUseCase) *RewardUseCase {
	return &RewardUseCase{
		repo:     r,
		notif:    notif,
		now:      time.Now,
		dispatch: func(f func()) { go f() },
	}
}

// push sends a notification without blocking or failing the caller.
func (u *RewardUseCase) push(req request.SendPushRequest) {
	if u.notif == nil {
		return
	}
	u.dispatch(func() { _, _ = u.notif.SendPushNotification(context.Background(), req) })
}
