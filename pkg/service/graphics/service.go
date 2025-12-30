package graphics

import (
	"os"
	"path/filepath"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/service/graphics/interfaces"
)

type GraphicsService struct {
	imageGenerator *OfferImageGenerator
}

func NewGraphicsService(outputDir string) interfaces.GraphicsService {
	// Ensure output directories exist
	os.MkdirAll(outputDir, 0755)
	os.MkdirAll(filepath.Join(outputDir, "thumbnail"), 0755)

	return &GraphicsService{
		imageGenerator: NewOfferImageGenerator(outputDir),
	}
}

func (g *GraphicsService) GenerateOfferImage(offer request.Offer, filename string) (string, error) {
	return g.imageGenerator.GenerateOfferImage(offer, filename)
}

func (g *GraphicsService) GenerateThumbnail(offer request.Offer, filename string) (string, error) {
	return g.imageGenerator.GenerateThumbnail(offer, filename)
}
