package service

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

// ProductValidationResponse for single image category validation
type ProductValidationResponse struct {
	Valid      bool    `json:"valid"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

// ImageComparisonResponse for comparing two product images
type ImageComparisonResponse struct {
	SameCategory         bool    `json:"same_category"`
	Confidence           float64 `json:"confidence"`
	CategoryDetectedImg1 string  `json:"category_detected_image1"`
	CategoryDetectedImg2 string  `json:"category_detected_image2"`
	Reason               string  `json:"reason"`
}

// ProductValidator handles product image validation using Claude Vision
type ProductValidator struct {
	APIKey     string
	HTTPClient *http.Client
}

// NewProductValidator creates a new product validator
func NewProductValidator(apiKey string) *ProductValidator {
	return &ProductValidator{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ValidateProductCategory validates if an uploaded image matches the product category
func (pv *ProductValidator) ValidateProductCategory(imagePath string, categoryName string) (*ProductValidationResponse, error) {
	// Read the image file
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file: %w", err)
	}

	// Encode to base64
	base64Data := base64.StdEncoding.EncodeToString(imageData)

	// Create the validation prompt
	prompt := fmt.Sprintf(`You are a strict e-commerce product validator.

Your task:
Determine whether the uploaded product image belongs to the given category.

Rules:
- Be strict.
- Focus on the main object in the image.
- Ignore background.
- Ignore branding unless it changes product type.
- If the main object does not clearly belong to the category, mark as INVALID.

Category: "%s"

Analyze the image and respond ONLY in valid JSON with this structure:
{
  "valid": true or false,
  "confidence": 0.0 to 1.0,
  "reason": "short explanation"
}`, categoryName)

	// Call Claude API
	responseText, err := pv.callClaudeAPI(base64Data, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call Claude API: %w", err)
	}

	// Parse the response
	var result ProductValidationResponse
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse validation response: %w", err)
	}

	return &result, nil
}

// CompareProductImages checks if two product images belong to the same category
func (pv *ProductValidator) CompareProductImages(imagePath1 string, imagePath2 string) (*ImageComparisonResponse, error) {
	// Read both image files
	imageData1, err := os.ReadFile(imagePath1)
	if err != nil {
		return nil, fmt.Errorf("failed to read first image file: %w", err)
	}

	imageData2, err := os.ReadFile(imagePath2)
	if err != nil {
		return nil, fmt.Errorf("failed to read second image file: %w", err)
	}

	base64Data1 := base64.StdEncoding.EncodeToString(imageData1)
	base64Data2 := base64.StdEncoding.EncodeToString(imageData2)

	// Create the comparison prompt
	prompt := `You are an AI product classifier.

Two product images are provided.

Your task:
Determine whether both images belong to the same product category.

Be strict and focus on:
- Type of product
- Shape
- Usage
- Structure

Ignore:
- Color differences
- Branding
- Background

Analyze both images and respond ONLY in valid JSON with this structure:
{
  "same_category": true or false,
  "confidence": 0.0 to 1.0,
  "category_detected_image1": "detected category",
  "category_detected_image2": "detected category",
  "reason": "brief explanation"
}`

	// Call Claude API with both images
	responseText, err := pv.callClaudeAPIWithMultipleImages(base64Data1, base64Data2, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call Claude API: %w", err)
	}

	// Parse the response
	var result ImageComparisonResponse
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse comparison response: %w", err)
	}

	return &result, nil
}

// callClaudeAPI sends a request to Claude API with a single image
func (pv *ProductValidator) callClaudeAPI(base64Image string, prompt string) (string, error) {
	// Create the request body
	requestBody := map[string]interface{}{
		"model":      "claude-3-5-sonnet-20241022",
		"max_tokens": 1024,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{
						"type": "image",
						"source": map[string]interface{}{
							"type":       "base64",
							"media_type": "image/jpeg",
							"data":       base64Image,
						},
					},
					map[string]interface{}{
						"type": "text",
						"text": prompt,
					},
				},
			},
		},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	return pv.sendAPIRequest(body)
}

// callClaudeAPIWithMultipleImages sends a request to Claude API with multiple images
func (pv *ProductValidator) callClaudeAPIWithMultipleImages(base64Image1 string, base64Image2 string, prompt string) (string, error) {
	// Create the request body with two images
	requestBody := map[string]interface{}{
		"model":      "claude-3-5-sonnet-20241022",
		"max_tokens": 1024,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "First product image:",
					},
					map[string]interface{}{
						"type": "image",
						"source": map[string]interface{}{
							"type":       "base64",
							"media_type": "image/jpeg",
							"data":       base64Image1,
						},
					},
					map[string]interface{}{
						"type": "text",
						"text": "Second product image:",
					},
					map[string]interface{}{
						"type": "image",
						"source": map[string]interface{}{
							"type":       "base64",
							"media_type": "image/jpeg",
							"data":       base64Image2,
						},
					},
					map[string]interface{}{
						"type": "text",
						"text": prompt,
					},
				},
			},
		},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	return pv.sendAPIRequest(body)
}

// sendAPIRequest makes the actual API call to Claude
func (pv *ProductValidator) sendAPIRequest(body []byte) (string, error) {
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", pv.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := pv.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Claude API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse Claude API response
	var apiResponse map[string]interface{}
	if err := json.Unmarshal(respBody, &apiResponse); err != nil {
		return "", fmt.Errorf("failed to parse Claude response: %w", err)
	}

	// Extract the text content from Claude's response
	content, ok := apiResponse["content"].([]interface{})
	if !ok || len(content) == 0 {
		return "", fmt.Errorf("invalid Claude response structure")
	}

	firstContent, ok := content[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid content structure in Claude response")
	}

	responseText, ok := firstContent["text"].(string)
	if !ok {
		return "", fmt.Errorf("no text in Claude response")
	}

	return responseText, nil
}
