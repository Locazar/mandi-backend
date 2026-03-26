# AI Service Integration Guide

This document explains how the AI Service microservice has been integrated with the mandi-backend.

## Overview

The AI Service is a separate microservice that handles Claude Vision and embedding APIs. Instead of calling external APIs directly from the main backend, all AI-related operations are now delegated to this dedicated service.

### Architecture

```
mandi-backend (port 3000)
    └──> HTTP calls
         └──> ai-service (port 3001)
              └──> Claude Vision API
              └──> Embedding Service
```

## Setup & Configuration

### 1. Environment Variables

Add the following to your `.env` file in the mandi-backend root:

```env
# AI Service Configuration
AI_SERVICE_URL=http://localhost:3001
```

For production, update this to your deployed AI service URL:
```env
AI_SERVICE_URL=http://ai-service.production.example.com
```

### 2. Docker Setup

Both services can be run together using docker-compose:

```bash
docker-compose up -d
```

This will start:
- **PostgreSQL** on port 5432
- **mandi-backend** on port 3000
- **ai-service** on port 3001

The services are connected via the `ecommerce-network` bridge network.

### 3. Manual Setup (Without Docker)

To run both services locally:

**Terminal 1 - AI Service:**
```bash
cd ../ai-service
cp .env.example .env
# Add your CLAUDE_API_KEY to .env
make run
# Service runs on http://localhost:3001
```

**Terminal 2 - Main Backend:**
```bash
cd mandi-backend
# Set AI_SERVICE_URL in .env
export AI_SERVICE_URL=http://localhost:3001
make run
# Backend runs on http://localhost:3000
```

## API Integration Points

### Product Image Validation

When a product is uploaded with a category, the system automatically validates the image:

**Flow:**
1. User uploads product image with category name
2. mandi-backend/ProductHandler receives the request
3. Calls `aiClient.ValidateProduct(imagePath, category)`
4. ai-service processes the image using Claude Vision API
5. Returns validation result (valid: true/false, confidence, reason)
6. Product is accepted or rejected based on validation confidence

**Code Location:** `pkg/api/handler/product.go:677-695`

### Code Example

```go
// In ProductHandler
validationResult, err := p.aiClient.ValidateProduct(imagePath, categoryName)
if err != nil {
    response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to validate product image", err, nil)
    return
}

// Check confidence threshold
if !validationResult.Valid && validationResult.Confidence > 0.6 {
    response.ErrorResponse(ctx, http.StatusBadRequest,
        fmt.Sprintf("Product image does not match '%s' category", categoryName),
        nil, nil)
    return
}
```

## AI Service Client

The AI service client is located at `pkg/service/ai/client.go` and provides the following methods:

### ValidateProduct

```go
result, err := aiClient.ValidateProduct(imagePath, category)
// Returns: *ProductValidationResponse {
//   Valid: bool,
//   Confidence: float64,
//   Reason: string
// }
```

### CompareImages

```go
result, err := aiClient.CompareImages(imagePath1, imagePath2)
// Returns: *ImageComparisonResponse
```

### GenerateEmbedding

```go
result, err := aiClient.GenerateEmbedding(text)
// Returns: *EmbeddingResponse {
//   Embedding: []float64,
//   Model: string
// }
```

### GenerateEmbeddings (Batch)

```go
result, err := aiClient.GenerateEmbeddings(texts)
// Returns: *BatchEmbeddingResponse {
//   Embeddings: [][]float64,
//   Model: string,
//   Count: int
// }
```

## Dependency Injection

The AI service client is automatically injected into the ProductHandler via dependency injection.

**File:** `pkg/di/wire_gen.go`

```go
// Created in InitializeApi()
aiClient := aiservice.NewClient(cfg.AIServiceURL)
productHandler := handler.NewProductHandler(productUseCase, tokenService, aiClient)
```

## Testing

### Manual Testing

Test the AI service health check:
```bash
curl http://localhost:3001/api/health
```

Test product validation:
```bash
curl -X POST http://localhost:3001/api/ai/validate-product \
  -H "Content-Type: application/json" \
  -d '{
    "image_path": "./uploads/products/sample.jpg",
    "category": "Electronics"
  }'
```

### Unit Tests

Run tests for the AI service client:
```bash
cd mandi-backend
go test ./pkg/service/ai/... -v
```

## Troubleshooting

### AI Service Not Responding

Check if the service is running:
```bash
curl http://localhost:3001/api/health
```

If not running, start it:
```bash
cd ../ai-service
make run
```

###  Connection Refused

Ensure `AI_SERVICE_URL` is correctly configured in your `.env`:
```env
AI_SERVICE_URL=http://localhost:3001
```

For Docker, use the service name:
```env
AI_SERVICE_URL=http://ai-service:3001
```

### Image Validation Failing

1. Ensure Claude API key is set in ai-service `.env`
2. Check image paths are absolute and valid
3. Review Claude API response for error details

## Performance Considerations

- **Timeout:** 60 seconds for image validation (Claude Vision API)
- **Timeout:** 30 seconds for embeddings
- **Batch Processing:** Send multiple texts at once for embeddings to reduce latency

## Security Notes

- The AI service runs on a private network when using Docker
- For production, run behind a reverse proxy (nginx, k8s ingress)
- Consider adding authentication tokens between services
- Secure Claude API key in environment variables only

## Future Enhancements

1. Add request caching for repeated validations
2. Implement rate limiting on AI service
3. Add monitoring and logging
4. Support additional AI models
5. Implement async processing for heavy operations

## References

- AI Service Repository: `../ai-service/`
- Copilot Instructions: `./.github/copilot-instructions.md`
- AI Service README: `../ai-service/README.md`
