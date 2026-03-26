# Dynamic Offer Graphics Generation

This module provides dynamic image generation for offers using Go graphics libraries.

## Features

- **Automatic Image Generation**: Creates beautiful offer images when offers are created
- **Dynamic Theming**: Different color schemes based on offer type and discount percentage
- **Thumbnail Generation**: Creates smaller thumbnail versions for listing pages
- **Customizable Templates**: Different layouts for different offer types (flash, seasonal, clearance, etc.)

## Components

### Graphics Service
- `pkg/service/graphics/offer_image.go` - Main image generation logic
- `pkg/service/graphics/service.go` - Service implementation
- `pkg/service/graphics/interfaces/graphics.go` - Service interface

### Image Features
- **Gradient Backgrounds**: Dynamic color gradients based on offer type
- **Decorative Elements**: Geometric patterns and circles for visual appeal
- **Discount Badge**: Prominent circular badge showing discount percentage
- **Typography**: Dynamic text sizing and positioning
- **Date Display**: Formatted offer validity dates

## Offer Types and Themes

### Flash Offers
- **Size**: 800x400px
- **Colors**: Deep Orange background with Amber accents
- **Use Case**: Limited time flash sales

### Seasonal Offers
- **Size**: 600x600px (square format)
- **Colors**: Green background with Yellow accents
- **Use Case**: Holiday and seasonal promotions

### Clearance Offers
- **Size**: 700x350px
- **Colors**: Purple background with Deep Orange accents
- **Use Case**: Inventory clearance sales

### Default Offers
- **Size**: 600x400px
- **Colors**: Blue background with Amber accents
- **Use Case**: General promotions

## Dynamic Color Schemes

The system automatically selects colors based on discount percentage:

- **50%+ Discount**: Red theme (high urgency)
- **25-49% Discount**: Orange theme (medium urgency)
- **Below 25%**: Blue theme (standard)

## Usage

### Automatic Generation
Images are automatically generated when creating offers through the standard offer creation endpoint:
```
POST /api/admin/offers
```

### Manual Preview Generation
For testing and preview purposes:
```
POST /api/admin/graphics/offer-preview
```

### Request Format
```json
{
    "offer_name": "Summer Sale",
    "description": "Amazing discounts on summer collection",
    "discount_rate": 30,
    "start_date": "2025-12-25T00:00:00.000",
    "end_date": "2025-12-31T23:59:59.000",
    "offer_type": "seasonal"
}
```

## File Structure

Generated images are stored in:
```
uploads/offers/
├── offer_[uuid].png          # Main offer images
└── thumbnail/
    └── offer_thumb_[uuid].png # Thumbnail versions
```

## Dependencies

- `github.com/fogleman/gg` - 2D graphics library
- `github.com/disintegration/imaging` - Image processing utilities

## Configuration

Images are generated in the following output directory (configurable):
```
./uploads/offers/
```

Thumbnails are stored in a subdirectory:
```
./uploads/offers/thumbnail/
```

## Integration

The graphics service is integrated into the offer usecase and automatically generates images during offer creation. The service uses dependency injection and can be easily mocked for testing.

## Error Handling

- Graceful fallback when fonts can't be loaded
- Comprehensive error reporting
- Directory creation for output paths
- Validation of input parameters