package usecase

import (
	"context"
	"fmt"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type shopProductEnquiryUseCase struct {
	repo         repo.ShopProductEnquiryRepository
	notification interfaces.NotificationUseCase
}

func NewShopProductEnquiryUseCase(r repo.ShopProductEnquiryRepository, n interfaces.NotificationUseCase) interfaces.ShopProductEnquiryUseCase {
	return &shopProductEnquiryUseCase{repo: r, notification: n}
}

func (uc *shopProductEnquiryUseCase) Submit(ctx context.Context, e domain.ShopProductEnquiry) error {
	return uc.repo.Create(ctx, e)
}

func (uc *shopProductEnquiryUseCase) List(ctx context.Context, shopID string, actioned *bool, enquiryType string) ([]domain.ShopProductEnquiry, error) {
	return uc.repo.List(ctx, shopID, actioned, enquiryType)
}

func (uc *shopProductEnquiryUseCase) MarkActioned(ctx context.Context, id string, note string) error {
	return uc.repo.MarkActioned(ctx, id, note)
}

func (uc *shopProductEnquiryUseCase) NotifyShop(ctx context.Context, enquiryID string) (bool, error) {
	e, err := uc.repo.GetByID(ctx, enquiryID)
	if err != nil {
		return false, fmt.Errorf("get enquiry: %w", err)
	}

	title := "📦 New product request from a customer"
	body := e.Message
	if len(body) > 200 {
		body = body[:197] + "…"
	}

	pushReq := request.SendPushRequest{
		OwnerID:   e.ShopID,
		OwnerType: "seller",
		Title:     title,
		Body:      body,
		EventType: "shop_product_enquiry",
		Data: map[string]string{
			"enquiry_id": e.ID,
			"shop_id":    e.ShopID,
			"type":       e.EnquiryType,
		},
	}

	// Use Background context so the push completes even if the HTTP request is gone.
	return uc.notification.SendPushNotification(context.Background(), pushReq)
}
