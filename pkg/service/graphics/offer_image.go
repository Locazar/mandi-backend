package graphics

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"github.com/fogleman/gg"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
)

type OfferImageGenerator struct {
	outputDir string
}

type OfferImageConfig struct {
	Width           int
	Height          int
	FontSize        float64
	BackgroundColor color.Color
	TextColor       color.Color
	AccentColor     color.Color
}

func NewOfferImageGenerator(outputDir string) *OfferImageGenerator {
	return &OfferImageGenerator{
		outputDir: outputDir,
	}
}

// GenerateOfferImage creates a dynamic offer image based on the offer data
func (o *OfferImageGenerator) GenerateOfferImage(offer request.Offer, filename string) (string, error) {
	config := o.getImageConfig(offer.Type)

	// Create context with specified dimensions
	dc := gg.NewContext(config.Width, config.Height)

	// Create gradient background
	o.drawGradientBackground(dc, config)

	// Draw decorative elements
	o.drawDecorativeElements(dc, config)

	// Draw offer content
	o.drawOfferContent(dc, offer, config)

	// Draw discount badge
	o.drawDiscountBadge(dc, offer.DiscountRate, config)

	// Save the image
	imagePath := filepath.Join(o.outputDir, filename)
	err := dc.SavePNG(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to save offer image: %w", err)
	}

	// Return web-accessible URL instead of file system path
	webURL := fmt.Sprintf("uploads/offers/%s", filename)
	return webURL, nil
}

func (o *OfferImageGenerator) getImageConfig(offerType string) OfferImageConfig {
	switch strings.ToLower(offerType) {
	case "flash":
		return OfferImageConfig{
			Width:           800,
			Height:          400,
			FontSize:        24,
			BackgroundColor: color.RGBA{255, 87, 34, 255},   // Deep Orange
			TextColor:       color.RGBA{255, 255, 255, 255}, // White
			AccentColor:     color.RGBA{255, 193, 7, 255},   // Amber
		}
	case "seasonal":
		return OfferImageConfig{
			Width:           600,
			Height:          600,
			FontSize:        20,
			BackgroundColor: color.RGBA{76, 175, 80, 255},   // Green
			TextColor:       color.RGBA{255, 255, 255, 255}, // White
			AccentColor:     color.RGBA{255, 235, 59, 255},  // Yellow
		}
	case "clearance":
		return OfferImageConfig{
			Width:           700,
			Height:          350,
			FontSize:        22,
			BackgroundColor: color.RGBA{156, 39, 176, 255},  // Purple
			TextColor:       color.RGBA{255, 255, 255, 255}, // White
			AccentColor:     color.RGBA{255, 87, 34, 255},   // Deep Orange
		}
	default:
		return OfferImageConfig{
			Width:           600,
			Height:          400,
			FontSize:        20,
			BackgroundColor: color.RGBA{33, 150, 243, 255},  // Blue
			TextColor:       color.RGBA{255, 255, 255, 255}, // White
			AccentColor:     color.RGBA{255, 193, 7, 255},   // Amber
		}
	}
}

func (o *OfferImageGenerator) drawGradientBackground(dc *gg.Context, config OfferImageConfig) {
	// Create gradient from main color to slightly darker shade
	gradient := gg.NewLinearGradient(0, 0, float64(config.Width), float64(config.Height))

	r, g, b, a := config.BackgroundColor.RGBA()
	gradient.AddColorStop(0, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)})

	// Darker shade for gradient end
	darkR := uint8(float64(r>>8) * 0.7)
	darkG := uint8(float64(g>>8) * 0.7)
	darkB := uint8(float64(b>>8) * 0.7)
	gradient.AddColorStop(1, color.RGBA{darkR, darkG, darkB, uint8(a >> 8)})

	dc.SetFillStyle(gradient)
	dc.DrawRectangle(0, 0, float64(config.Width), float64(config.Height))
	dc.Fill()
}

func (o *OfferImageGenerator) drawDecorativeElements(dc *gg.Context, config OfferImageConfig) {
	// Draw decorative circles
	dc.SetColor(config.AccentColor)

	// Top right circle
	dc.DrawCircle(float64(config.Width)*0.9, float64(config.Height)*0.1, 30)
	dc.SetRGBA255(255, 255, 255, 50)
	dc.Fill()

	// Bottom left circle
	dc.SetColor(config.AccentColor)
	dc.DrawCircle(float64(config.Width)*0.1, float64(config.Height)*0.9, 40)
	dc.SetRGBA255(255, 255, 255, 30)
	dc.Fill()

	// Draw geometric patterns
	dc.SetColor(color.RGBA{255, 255, 255, 80})
	dc.SetLineWidth(2)
	for i := 0; i < 5; i++ {
		y := float64(config.Height) * (0.2 + float64(i)*0.15)
		dc.DrawLine(float64(config.Width)*0.8, y, float64(config.Width)*0.95, y)
		dc.Stroke()
	}
}

func (o *OfferImageGenerator) drawOfferContent(dc *gg.Context, offer request.Offer, config OfferImageConfig) {
	dc.SetColor(config.TextColor)

	// Draw offer name
	titleY := float64(config.Height) * 0.25
	dc.DrawStringAnchored(strings.ToUpper(offer.Name), float64(config.Width)*0.5, titleY, 0.5, 0.5)

	// Draw description with word wrapping
	descY := float64(config.Height) * 0.45
	maxWidth := float64(config.Width) * 0.7
	o.drawWrappedText(dc, offer.Description, float64(config.Width)*0.5, descY, maxWidth, 1.2)

	// Draw dates
	dateY := float64(config.Height) * 0.75
	dateText := fmt.Sprintf("Valid from %s to %s",
		offer.StartDate.Format("Jan 02"),
		offer.EndDate.Format("Jan 02, 2006"))
	dc.DrawStringAnchored(dateText, float64(config.Width)*0.5, dateY, 0.5, 0.5)
}

func (o *OfferImageGenerator) drawDiscountBadge(dc *gg.Context, discountRate uint, config OfferImageConfig) {
	// Position for discount badge (top left)
	badgeX := float64(config.Width) * 0.15
	badgeY := float64(config.Height) * 0.15
	badgeRadius := 60.0

	// Draw badge background
	dc.SetColor(config.AccentColor)
	dc.DrawCircle(badgeX, badgeY, badgeRadius)
	dc.Fill()

	// Draw badge border
	dc.SetColor(config.TextColor)
	dc.SetLineWidth(3)
	dc.DrawCircle(badgeX, badgeY, badgeRadius)
	dc.Stroke()

	// Draw discount percentage
	dc.SetColor(color.RGBA{0, 0, 0, 255}) // Black for contrast
	dc.DrawStringAnchored(fmt.Sprintf("%d%%", discountRate), badgeX, badgeY-10, 0.5, 0.5)

	// Draw "OFF" text
	dc.DrawStringAnchored("OFF", badgeX, badgeY+15, 0.5, 0.5)
}

func (o *OfferImageGenerator) drawWrappedText(dc *gg.Context, text string, x, y, maxWidth, lineSpacing float64) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return
	}

	lines := []string{}
	currentLine := ""

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		w, _ := dc.MeasureString(testLine)
		if w <= maxWidth || currentLine == "" {
			currentLine = testLine
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	totalHeight := float64(len(lines)-1) * lineSpacing * 20 // Approximate line height
	startY := y - totalHeight/2

	for i, line := range lines {
		lineY := startY + float64(i)*lineSpacing*20
		dc.DrawStringAnchored(line, x, lineY, 0.5, 0.5)
	}
}

// GenerateThumbnail creates a smaller thumbnail version of the offer image
func (o *OfferImageGenerator) GenerateThumbnail(offer request.Offer, filename string) (string, error) {
	// Create smaller context for thumbnail
	dc := gg.NewContext(200, 200)

	// Simple thumbnail design
	config := o.getImageConfig(offer.Type)

	// Background
	dc.SetColor(config.BackgroundColor)
	dc.DrawRectangle(0, 0, 200, 200)
	dc.Fill()

	// Discount text
	dc.SetColor(config.TextColor)
	dc.DrawStringAnchored(fmt.Sprintf("%d%%", offer.DiscountRate), 100, 80, 0.5, 0.5)

	// OFF text
	dc.DrawStringAnchored("OFF", 100, 120, 0.5, 0.5)

	// Offer name (truncated)
	name := offer.Name
	if len(name) > 15 {
		name = name[:12] + "..."
	}
	dc.DrawStringAnchored(name, 100, 160, 0.5, 0.5)

	// Save thumbnail
	thumbDir := filepath.Join(o.outputDir, "thumbnail")
	// Create thumbnail directory if it doesn't exist
	os.MkdirAll(thumbDir, 0755)

	thumbPath := filepath.Join(thumbDir, filename)
	err := dc.SavePNG(thumbPath)
	if err != nil {
		return "", fmt.Errorf("failed to save thumbnail: %w", err)
	}

	// Return web-accessible URL instead of file system path
	webURL := fmt.Sprintf("uploads/offers/thumbnail/%s", filename)
	return webURL, nil
}

// GetOfferImageColors returns colors based on discount percentage for dynamic theming
func (o *OfferImageGenerator) GetOfferImageColors(discountRate uint) OfferImageConfig {
	switch {
	case discountRate >= 50:
		// High discount - Red theme
		return OfferImageConfig{
			BackgroundColor: color.RGBA{244, 67, 54, 255},   // Red
			TextColor:       color.RGBA{255, 255, 255, 255}, // White
			AccentColor:     color.RGBA{255, 193, 7, 255},   // Amber
		}
	case discountRate >= 25:
		// Medium discount - Orange theme
		return OfferImageConfig{
			BackgroundColor: color.RGBA{255, 152, 0, 255},   // Orange
			TextColor:       color.RGBA{255, 255, 255, 255}, // White
			AccentColor:     color.RGBA{76, 175, 80, 255},   // Green
		}
	default:
		// Low discount - Blue theme
		return OfferImageConfig{
			BackgroundColor: color.RGBA{33, 150, 243, 255},  // Blue
			TextColor:       color.RGBA{255, 255, 255, 255}, // White
			AccentColor:     color.RGBA{255, 193, 7, 255},   // Amber
		}
	}
}
