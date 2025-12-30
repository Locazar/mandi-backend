package interfaces

import "github.com/rohit221990/mandi-backend/pkg/api/handler/request"

type GraphicsService interface {
	GenerateOfferImage(offer request.Offer, filename string) (string, error)
	GenerateThumbnail(offer request.Offer, filename string) (string, error)
}
