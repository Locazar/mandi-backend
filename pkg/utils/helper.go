package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type ModerationResponse struct {
	Status string `json:"status"`
	Nudity struct {
		Raw float64 `json:"raw"`
	} `json:"nudity"`
	Error struct {
		Type    string `json:"type"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func CheckNudity(filename string) (bool, error) {
	// Extract just the filename if full path is provided
	// e.g., "uploads/products/abc123.jpg" -> "abc123.jpg"
	justFilename := filepath.Base(filename)

	// Get the current working directory (server root)
	wd, err := os.Getwd()
	if err != nil {
		return false, fmt.Errorf("failed to get working directory: %w", err)
	}

	fullPath := filepath.Join(wd, "uploads", "products", filepath.Base(justFilename))
	fmt.Printf("Checking nudity for file: %s\n", fullPath)

	// Open the file
	file, err := os.Open(fullPath)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create multipart form data
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	// Add file to multipart form with key 'media'
	part, err := writer.CreateFormFile("media", justFilename)
	if err != nil {
		return false, fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return false, fmt.Errorf("failed to copy file: %w", err)
	}

	// Add API parameters
	_ = writer.WriteField("models", "nudity-2.1")
	_ = writer.WriteField("api_user", "1350960651")
	_ = writer.WriteField("api_secret", "xD7trXQ3EDEzJsd4Msy5bZzVZCXADoJf")

	err = writer.Close()
	if err != nil {
		return false, fmt.Errorf("failed to close writer: %w", err)
	}

	// Build API request
	apiURL := "https://api.sightengine.com/1.0/check.json"
	req, err := http.NewRequest("POST", apiURL, payload)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	var result ModerationResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	// Print the full response for debugging
	fmt.Printf("Sightengine API Response: %+v\n", result)
	fmt.Printf("Response Status: %s\n", result.Status)
	if result.Status == "success" {
		fmt.Printf("Nudity Score (raw): %f\n", result.Nudity.Raw)
	} else if result.Status == "failure" {
		fmt.Printf("Error Type: %s\n", result.Error.Type)
		fmt.Printf("Error Code: %d\n", result.Error.Code)
		fmt.Printf("Error Message: %s\n", result.Error.Message)
	}

	// If 'raw' nudity score > 0.5, it's adult content
	if result.Status == "success" && result.Nudity.Raw > 0.5 {
		return true, nil // It is adult content
	}

	return false, nil // Safe
}
