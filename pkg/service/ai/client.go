// Package ai provides an HTTP client for the ai-service microservice (default port 3001).
// It wraps Claude Vision and Embedding endpoints exposed by the ai-service.
//
// The base URL is configured via the AI_SERVICE_URL environment variable (see pkg/config).
// The client has a 60-second timeout — do not shorten it; Claude Vision calls are slow.
//
// Available operations:
//   - ValidateProduct: checks if a product image matches a given category (JPEG only).
//   - CompareImages: determines if two images belong to the same product category (JPEG only).
//   - DetectObjects: detects objects (label/category/color/material/brand) in a base64 JPEG.
//   - VerifyImage: checks whether a base64 JPEG satisfies a free-text condition prompt.
//   - GenerateEmbedding / GenerateEmbeddings: returns float64 embeddings for text search.
//
// ValidateProduct and CompareImages send a server-readable image_path; DetectObjects and
// VerifyImage send the image inline as base64 (use the *FromPath helpers to encode a local file).
//
// Warning: all image inputs must be JPEG. The ai-service hardcodes media_type: image/jpeg;
// PNG or WebP inputs will cause Claude API errors on the ai-service side.
package ai

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ServiceResponse is the standard response from AI service
type ServiceResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error"`
}

// ProductValidationRequest is the request body for product validation
type ProductValidationRequest struct {
	ImagePath    string `json:"image_path"`
	CategoryName string `json:"category"`
}

// ProductValidationResponse is the response for product validation
type ProductValidationResponse struct {
	Valid      bool    `json:"valid"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

// ImageComparisonRequest is the request body for image comparison
type ImageComparisonRequest struct {
	ImagePath1 string `json:"image_path1"`
	ImagePath2 string `json:"image_path2"`
}

// ImageComparisonResponse is the response for image comparison
type ImageComparisonResponse struct {
	SameCategory         bool    `json:"same_category"`
	Confidence           float64 `json:"confidence"`
	CategoryDetectedImg1 string  `json:"category_detected_image1"`
	CategoryDetectedImg2 string  `json:"category_detected_image2"`
	Reason               string  `json:"reason"`
}

// EmbeddingRequest is the request body for embedding generation
type EmbeddingRequest struct {
	Text string `json:"text"`
}

// EmbeddingResponse is the response for embedding
type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
	Model     string    `json:"model"`
}

// BatchEmbeddingRequest is the request for batch embeddings
type BatchEmbeddingRequest struct {
	Texts []string `json:"texts"`
}

// BatchEmbeddingResponse is the response for batch embeddings
type BatchEmbeddingResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
	Model      string      `json:"model"`
	Count      int         `json:"count"`
}

// ObjectDetectionRequest is the request body for object detection.
// The image is sent inline as a base64-encoded JPEG.
type ObjectDetectionRequest struct {
	ImageBase64 string `json:"image_base64"`
	DetectAll   bool   `json:"detect_all"`
}

// DetectedObject is a single object detected in an image.
type DetectedObject struct {
	Label      string  `json:"label"`
	Category   string  `json:"category"`
	Color      string  `json:"color"`
	Material   string  `json:"material"`
	Brand      string  `json:"brand"`
	Confidence float64 `json:"confidence"`
}

// ObjectDetectionResponse is the response for object detection.
type ObjectDetectionResponse struct {
	Objects []DetectedObject `json:"objects"`
}

// ImageVerificationRequest is the request body for prompt-based image verification.
// The image is sent inline as a base64-encoded JPEG.
type ImageVerificationRequest struct {
	ImageBase64 string `json:"image_base64"`
	Prompt      string `json:"prompt"`
}

// ImageVerificationResponse is the verdict for prompt-based image verification.
type ImageVerificationResponse struct {
	Matches    bool    `json:"matches"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

// Client is the HTTP client for AI service
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new AI service client
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:3001"
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ValidateProduct calls the product validation endpoint
func (c *Client) ValidateProduct(imagePath string, category string) (*ProductValidationResponse, error) {
	req := ProductValidationRequest{
		ImagePath:    imagePath,
		CategoryName: category,
	}

	var result ServiceResponse
	if err := c.post("/api/ai/validate-product", req, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("validation failed: %s", result.Error)
	}

	// Extract validation response from data
	dataBytes, err := json.Marshal(result.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var validationResp ProductValidationResponse
	if err := json.Unmarshal(dataBytes, &validationResp); err != nil {
		return nil, fmt.Errorf("failed to parse validation response: %w", err)
	}

	return &validationResp, nil
}

// CompareImages calls the image comparison endpoint
func (c *Client) CompareImages(imagePath1 string, imagePath2 string) (*ImageComparisonResponse, error) {
	req := ImageComparisonRequest{
		ImagePath1: imagePath1,
		ImagePath2: imagePath2,
	}

	var result ServiceResponse
	if err := c.post("/api/ai/compare-images", req, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("comparison failed: %s", result.Error)
	}

	// Extract comparison response from data
	dataBytes, err := json.Marshal(result.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var comparisonResp ImageComparisonResponse
	if err := json.Unmarshal(dataBytes, &comparisonResp); err != nil {
		return nil, fmt.Errorf("failed to parse comparison response: %w", err)
	}

	return &comparisonResp, nil
}

// GenerateEmbedding calls the embedding generation endpoint
func (c *Client) GenerateEmbedding(text string) (*EmbeddingResponse, error) {
	req := EmbeddingRequest{Text: text}

	var result ServiceResponse
	if err := c.post("/api/ai/embed", req, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("embedding generation failed: %s", result.Error)
	}

	// Extract embedding response from data
	dataBytes, err := json.Marshal(result.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var embeddingResp EmbeddingResponse
	if err := json.Unmarshal(dataBytes, &embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	return &embeddingResp, nil
}

// GenerateEmbeddings calls the batch embedding generation endpoint
func (c *Client) GenerateEmbeddings(texts []string) (*BatchEmbeddingResponse, error) {
	req := BatchEmbeddingRequest{Texts: texts}

	var result ServiceResponse
	if err := c.post("/api/ai/embed-batch", req, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("batch embedding generation failed: %s", result.Error)
	}

	// Extract batch embedding response from data
	dataBytes, err := json.Marshal(result.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var batchResp BatchEmbeddingResponse
	if err := json.Unmarshal(dataBytes, &batchResp); err != nil {
		return nil, fmt.Errorf("failed to parse batch embedding response: %w", err)
	}

	return &batchResp, nil
}

// DetectObjects calls the object detection endpoint with a base64-encoded JPEG.
// When detectAll is false the service returns only the single most prominent object.
func (c *Client) DetectObjects(imageBase64 string, detectAll bool) (*ObjectDetectionResponse, error) {
	req := ObjectDetectionRequest{
		ImageBase64: imageBase64,
		DetectAll:   detectAll,
	}

	var result ServiceResponse
	if err := c.post("/api/ai/detect-objects", req, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("object detection failed: %s", result.Error)
	}

	// Extract detection response from data
	dataBytes, err := json.Marshal(result.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var detectionResp ObjectDetectionResponse
	if err := json.Unmarshal(dataBytes, &detectionResp); err != nil {
		return nil, fmt.Errorf("failed to parse detection response: %w", err)
	}

	return &detectionResp, nil
}

// DetectObjectsFromPath reads a local JPEG file, base64-encodes it, and calls DetectObjects.
func (c *Client) DetectObjectsFromPath(imagePath string, detectAll bool) (*ObjectDetectionResponse, error) {
	encoded, err := encodeImageFile(imagePath)
	if err != nil {
		return nil, err
	}
	return c.DetectObjects(encoded, detectAll)
}

// VerifyImage calls the prompt-based image verification endpoint with a base64-encoded JPEG.
// It returns whether the image satisfies the supplied condition prompt.
func (c *Client) VerifyImage(imageBase64 string, prompt string) (*ImageVerificationResponse, error) {
	req := ImageVerificationRequest{
		ImageBase64: imageBase64,
		Prompt:      prompt,
	}

	var result ServiceResponse
	if err := c.post("/api/ai/verify-image", req, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("image verification failed: %s", result.Error)
	}

	// Extract verification response from data
	dataBytes, err := json.Marshal(result.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var verificationResp ImageVerificationResponse
	if err := json.Unmarshal(dataBytes, &verificationResp); err != nil {
		return nil, fmt.Errorf("failed to parse verification response: %w", err)
	}

	return &verificationResp, nil
}

// VerifyImageFromPath reads a local JPEG file, base64-encodes it, and calls VerifyImage.
func (c *Client) VerifyImageFromPath(imagePath string, prompt string) (*ImageVerificationResponse, error) {
	encoded, err := encodeImageFile(imagePath)
	if err != nil {
		return nil, err
	}
	return c.VerifyImage(encoded, prompt)
}

// encodeImageFile reads an image file from disk and returns its base64 (std) encoding.
func encodeImageFile(imagePath string) (string, error) {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image file %q: %w", imagePath, err)
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// HealthCheck checks if the AI service is healthy
func (c *Client) HealthCheck() error {
	resp, err := c.httpClient.Get(c.baseURL + "/api/health")
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// post is a helper method to make POST requests
func (c *Client) post(endpoint string, reqBody interface{}, respBody interface{}) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+endpoint, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respData))
	}

	if err := json.Unmarshal(respData, respBody); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}
