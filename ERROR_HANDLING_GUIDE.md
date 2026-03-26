## Mandi-Backend Error Handling Framework

This document explains the standardized error handling framework used across the mandi-backend application.

### Overview

The error handling framework is built on three main components:

1. **Error Domain Models** (`pkg/domain/error.go`) - Defines error codes and the `AppError` type
2. **Response Handler** (`pkg/api/response/response.go`) - Formats API responses
3. **Error Converters** (`pkg/utils/error_converter.go`) - Converts errors between layers

### Architecture

```
Repository Layer
    ↓
Converts DB errors to AppError using:
  - utils.ConvertDBError()
  - repository.ValidateRequired()
    ↓
UseCase Layer
    ↓
Returns AppError or business logic errors
    ↓
Handler Layer
    ↓
Formats response using:
  - response.SuccessResponse()
  - response.ErrorResponse(ctx, appErr)
    ↓
JSON Response with standardized structure
```

### Error Response Structure

All error responses follow this JSON structure:

```json
{
  "success": false,
  "message": "Human-readable error message",
  "error": {
    "code": "ERROR_CODE",
    "message": "Error message",
    "details": "Additional context (optional)"
  },
  "timestamp": 1234567890
}
```

### Success Response Structure

```json
{
  "success": true,
  "message": "Operation successful",
  "data": {
    "field": "value"
  },
  "timestamp": 1234567890
}
```

### Available Error Codes

#### Validation Errors (400 Bad Request)
- `VALIDATION_ERROR` - General validation failed
- `INVALID_INPUT` - Input validation failed
- `MISSING_FIELD` - Required field is missing
- `INVALID_FORMAT` - Invalid data format

#### Authentication Errors (401 Unauthorized)
- `UNAUTHORIZED` - User not authenticated
- `TOKEN_EXPIRED` - JWT token has expired
- `INVALID_TOKEN` - Invalid JWT token
- `FORBIDDEN` (403) - User lacks permission

#### Resource Errors (404/409)
- `NOT_FOUND` - Resource not found (404)
- `ALREADY_EXISTS` - Resource already exists (409)
- `CONFLICT` - Operation conflict (409)

#### Business Logic Errors (422 Unprocessable Entity)
- `IMAGE_MISMATCH` - Image doesn't match category
- `INSUFFICIENT_STOCK` - Not enough inventory
- `PAYMENT_FAILED` - Payment processing failed
- `BUSINESS_LOGIC_ERROR` - General business rule violation

#### Server Errors (500 Internal Server Error)
- `INTERNAL_SERVER_ERROR` - Generic server error
- `DATABASE_ERROR` - Database operation failed
- `EXTERNAL_SERVICE_ERROR` - Third-party service failed
- `TIMEOUT` - Operation timed out
- `FILE_OPERATION_ERROR` - File operation failed

### Usage Patterns

#### In Handler Layer

```go
func (h *MyHandler) GetProduct(ctx *gin.Context) {
    id := ctx.Param("id")
    
    // Validate input
    if id == "" {
        appErr := domain.ValidationError("id", "product ID is required")
        response.ErrorResponse(ctx, appErr)
        return
    }
    
    // Call usecase
    product, err := h.useCase.GetProduct(ctx, id)
    if err != nil {
        // If err is already *AppError, pass it directly
        if appErr, ok := err.(*domain.AppError); ok {
            response.ErrorResponse(ctx, appErr)
            return
        }
        // Otherwise wrap it
        appErr := domain.InternalError("failed to retrieve product", err)
        response.ErrorResponse(ctx, appErr)
        return
    }
    
    // Success response
    response.SuccessResponse(ctx, http.StatusOK, "Product retrieved successfully", product)
}
```

#### In Repository Layer

```go
func (r *ProductRepository) GetByID(ctx context.Context, id uint) (*domain.Product, *domain.AppError) {
    // Validate input
    if err := repository.ValidateID("product_id", id); err != nil {
        return nil, err
    }
    
    var product domain.Product
    err := r.db.WithContext(ctx).First(&product, id).Error
    
    // Convert database errors
    if err != nil {
        return nil, repository.HandleDBErrorContext(err, "GetByID", "Product")
    }
    
    return &product, nil
}
```

#### In UseCase Layer

```go
func (u *ProductUseCase) SaveProduct(ctx context.Context, req request.Product) (uint, *domain.AppError) {
    // Validate business rules
    if req.Name == "" {
        return 0, domain.ValidationError("name", "product name is required")
    }
    
    if req.Price <= 0 {
        return 0, domain.ValidationError("price", "price must be greater than 0")
    }
    
    // Check if product already exists
    existing, err := u.repo.GetByName(ctx, req.Name)
    if existing != nil {
        return 0, domain.AlreadyExistsError("product")
    }
    if err != nil && !errors.Is(err.Err, sql.ErrNoRows) {
        return 0, err // Return AppError from repository
    }
    
    // Create product
    product := domain.Product{
        Name:        req.Name,
        Price:       req.Price,
        Description: req.Description,
    }
    
    // Save to repository
    id, err := u.repo.Save(ctx, product)
    if err != nil {
        return 0, err
    }
    
    return id, nil
}
```

### Creating Custom AppErrors

#### Simple Error
```go
appErr := domain.NotFoundError("User")
// Result: "User not found"
```

#### With Details
```go
appErr := domain.VideoErrorError("Insufficient stock")
details := fmt.Sprintf("Product: %s, Requested: %d, Available: %d", name, requested, available)
appErr.Details = details
```

#### External Service Error
```go
appErr := domain.ExternalServiceError("payment-gateway", "timeout connecting to Stripe", err)
```

#### Custom Error
```go
appErr := domain.NewAppError(
    domain.ErrCodeCustom,
    "Operation failed",
    "Additional context here",
    underlyingError,
)
```

### Error Handling Best Practices

1. **Convert errors at layer boundaries**
   - Repository → UseCase: Convert DB errors to AppError
   - UseCase → Handler: Return AppError
   - Handler → Response: Format AppError for JSON

2. **Preserve error context**
   - Always wrap errors with meaningful messages
   - Use `Details` field for additional debugging info
   - Keep original error in `Err` field for logging

3. **Use specific error codes**
   - Don't use generic `INTERNAL_SERVER_ERROR` for known failures
   - Use `NOT_FOUND` for missing resources
   - Use `ALREADY_EXISTS` for duplicate resources
   - Use `IMAGE_MISMATCH` for business rule violations

4. **Handle errors early**
   ```go
   // BAD: Passing nil checks deep into code
   if product != nil && product.Category != nil {
       // deep nesting
   }
   
   // GOOD: Early return with error
   if appErr := validateProduct(product); appErr != nil {
       return nil, appErr
   }
   ```

5. **Log errors appropriately**
   - SystemErrors (5xx) should be logged for debugging
   - ClientErrors (4xx) might not need logging
   - Always log the original error wrapped in AppError

### Pagination with New Framework

```go
func (h *MyHandler) ListProducts(ctx *gin.Context) {
    pagination := request.GetPagination(ctx)
    
    products, total, err := h.useCase.FindAll(ctx, pagination)
    if err != nil {
        response.ErrorResponse(ctx, err)
        return
    }
    
    response.SuccessPaginatedResponse(ctx, http.StatusOK, "Products retrieved", products, total, pagination.Page, pagination.Limit)
}
```

### Testing with Error Framework

```go
func TestGetProduct_NotFound(t *testing.T) {
    mockRepo := new(MockRepository)
    mockRepo.On("GetByID", mock.Anything, uint(1)).Return(
        nil,
        domain.NotFoundError("Product"),
    )
    
    usecase := NewProductUseCase(mockRepo)
    result, err := usecase.GetByID(context.Background(), 1)
    
    assert.Nil(t, result)
    assert.NotNil(t, err)
    assert.Equal(t, domain.ErrCodeNotFound, err.Code)
    assert.Equal(t, http.StatusNotFound, err.StatusCode)
}
```

### Migration Guide

For existing code using the old error handling:

**Before:**
```go
response.ErrorResponse(ctx, http.StatusBadRequest, "message", err, nil)
```

**After:**
```go
appErr := domain.ValidationError("field", "reason")
response.ErrorResponse(ctx, appErr)
```

**Before:**
```go
response.SuccessResponse(ctx, http.StatusOK, "message", data)
```

**After:**
```go
response.SuccessResponse(ctx, http.StatusOK, "message", data)
// No change needed, signature is compatible
```

### Common Scenarios

#### Product Not Found
```go
return nil, domain.NotFoundError("Product")
// Returns 404
```

#### Duplicate Product
```go
return 0, domain.AlreadyExistsError("Product")
// Returns 409
```

#### Validation Failed
```go
return nil, domain.ValidationError("price", "price must be positive")
// Returns 400
```

#### External Service Failure
```go
return nil, domain.ExternalServiceError("sms-gateway", "failed to send OTP", err)
// Returns 500
```

### Summary

This error handling framework ensures:
- ✅ Consistent error responses across all endpoints
- ✅ Clear error codes for client-side handling
- ✅ Detailed error information for debugging
- ✅ Proper HTTP status codes
- ✅ Clean separation of concerns between layers
- ✅ Easy error conversion at layer boundaries
