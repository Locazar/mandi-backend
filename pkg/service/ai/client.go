package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
