package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type ModerationResponse struct {
	Status string `json:"status"`
	Nudity struct {
		SexualActivity   float64 `json:"sexual_activity"`
		SexualDisplay    float64 `json:"sexual_display"`
		Erotica          float64 `json:"erotica"`
		VerySuggestive   float64 `json:"very_suggestive"`
		Suggestive       float64 `json:"suggestive"`
		MildlySuggestive float64 `json:"mildly_suggestive"`
		None             float64 `json:"none"`
	} `json:"nudity"`
	Error struct {
		Type    string `json:"type"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func CheckNudity(path string) (bool, error) {
	justFilename := filepath.Base(path)

	file, err := os.Open(path)
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
	_ = writer.WriteField("models", "nudity-2.1,weapons-1.0,offensive-1.0,drugs-1.0,gore-2.0,alcohol-1.0,political-1.0,terrorism-1.0,self-harm-1.0,discrimination-1.0,adult-1.0,sexual-1.0,explicit-1.0,sexual_activity-1.0,sexual_display-1.0,erotica-1.0,very_suggestive-1.0,suggestive-1.0,mildly_suggestive-1.0,none-1.0,military-1.0")
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
		fmt.Printf("Sexual Activity: %f\n", result.Nudity.SexualActivity)
		fmt.Printf("Sexual Display: %f\n", result.Nudity.SexualDisplay)
		fmt.Printf("Erotica: %f\n", result.Nudity.Erotica)
		fmt.Printf("Very Suggestive: %f\n", result.Nudity.VerySuggestive)
		fmt.Printf("Suggestive: %f\n", result.Nudity.Suggestive)
	} else if result.Status == "failure" {
		fmt.Printf("Error Type: %s\n", result.Error.Type)
		fmt.Printf("Error Code: %d\n", result.Error.Code)
		fmt.Printf("Error Message: %s\n", result.Error.Message)
	}

	// If any of the explicit adult content scores > 0.5, it's adult content
	// sexual_activity and sexual_display are clear indicators of adult content
	// erotica and very_suggestive are also strong indicators
	if result.Status == "success" {
		if result.Nudity.SexualActivity > 0.5 ||
			result.Nudity.SexualDisplay > 0.5 ||
			result.Nudity.Erotica > 0.5 ||
			result.Nudity.VerySuggestive > 0.5 {
			return true, nil // It is adult content
		}
	}

	return false, nil // Safe
}