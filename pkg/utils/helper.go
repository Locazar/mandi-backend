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
		SexualActivity   float64 `json:"sexual_activity"`
		SexualDisplay    float64 `json:"sexual_display"`
		Erotica          float64 `json:"erotica"`
		VerySuggestive   float64 `json:"very_suggestive"`
		Suggestive       float64 `json:"suggestive"`
		MildlySuggestive float64 `json:"mildly_suggestive"`
		None             float64 `json:"none"`
	} `json:"nudity"`
	Gore struct {
		Prob float64 `json:"prob"`
	} `json:"gore"`
	Violence struct {
		Prob float64 `json:"prob"`
	} `json:"violence"`
	SelfHarm struct {
		Prob float64 `json:"prob"`
	} `json:"self-harm"`
	// WeaponFirearm/WeaponKnife are Sightengine's top-level convenience
	// aggregates for the "weapon" model (rather than the nested per-class
	// breakdown), same shape as the drug/medical top-level scores below.
	WeaponFirearm float64 `json:"weapon_firearm"`
	WeaponKnife   float64 `json:"weapon_knife"`
	Offensive     struct {
		Prob         float64 `json:"prob"`
		Nazi         float64 `json:"nazi"`
		Confederate  float64 `json:"confederate"`
		Supremacist  float64 `json:"supremacist"`
		Terrorist    float64 `json:"terrorist"`
		MiddleFinger float64 `json:"middle_finger"`
	} `json:"offensive"`
	Scam struct {
		Prob float64 `json:"prob"`
	} `json:"scam"`
	Tobacco struct {
		Prob float64 `json:"prob"`
	} `json:"tobacco"`
	RecreationalDrug struct {
		Prob float64 `json:"prob"`
	} `json:"recreational_drug"`
	Alcohol struct {
		Prob float64 `json:"prob"`
	} `json:"alcohol"`
	Error struct {
		Type    string `json:"type"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// CheckNudity runs the uploaded image through Sightengine's moderation models.
// It reports whether the image must be rejected and, when it must, a short
// human-readable reason naming the category that tripped, so callers never tell
// a seller "adult content" for a weapon or scam hit.
func CheckNudity(path string) (bool, string, error) {
	justFilename := filepath.Base(path)

	file, err := os.Open(path)
	if err != nil {
		return false, "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create multipart form data
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	// Add file to multipart form with key 'media'
	part, err := writer.CreateFormFile("media", justFilename)
	if err != nil {
		return false, "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return false, "", fmt.Errorf("failed to copy file: %w", err)
	}

	// Add API parameters.
	// IMPORTANT: Sightengine rejects the *entire* request (status: failure)
	// if even one model name here is invalid — verify a name against
	// https://sightengine.com/docs/models before adding it back.
	_ = writer.WriteField("models", "nudity-2.1,gore-2.0,violence,self-harm,weapon,offensive,scam,tobacco,recreational_drug,alcohol")
	_ = writer.WriteField("api_user", "1350960651")
	_ = writer.WriteField("api_secret", "xD7trXQ3EDEzJsd4Msy5bZzVZCXADoJf")

	err = writer.Close()
	if err != nil {
		return false, "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Build API request
	apiURL := "https://api.sightengine.com/1.0/check.json"
	req, err := http.NewRequest("POST", apiURL, payload)
	if err != nil {
		return false, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	var result ModerationResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return false, "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Fail closed: if Sightengine didn't return a clean success, we have no
	// signal either way, so don't let the image through unmoderated.
	if result.Status != "success" {
		return false, "", fmt.Errorf(
			"moderation check failed: %s (code %d): %s",
			result.Error.Type, result.Error.Code, result.Error.Message,
		)
	}

	const threshold = 0.5

	// The scam model fires on text and graphics overlaid on the photo: stamped
	// prices, phone numbers, "WhatsApp us", watermarks, collage borders. That
	// is how most sellers shoot their catalogue, so at the 0.5 used for
	// genuinely unsafe categories it rejects ordinary listings (a clean
	// clothing photo scored 0.74). Only a near-certain score is worth blocking
	// an upload over.
	const scamThreshold = 0.9

	reason := ""
	switch {
	// sexual_activity/sexual_display are clear indicators of adult content;
	// erotica/very_suggestive are also strong indicators.
	case result.Nudity.SexualActivity > threshold,
		result.Nudity.SexualDisplay > threshold,
		result.Nudity.Erotica > threshold,
		result.Nudity.VerySuggestive > threshold:
		reason = "adult content"

	case result.Gore.Prob > threshold,
		result.Violence.Prob > threshold,
		result.SelfHarm.Prob > threshold:
		reason = "violent or graphic content"

	case result.WeaponFirearm > threshold, result.WeaponKnife > threshold:
		reason = "weapons"

	case result.Offensive.Prob > threshold,
		result.Offensive.Nazi > threshold,
		result.Offensive.Confederate > threshold,
		result.Offensive.Supremacist > threshold,
		result.Offensive.Terrorist > threshold:
		reason = "offensive content"

	case result.Scam.Prob > scamThreshold:
		reason = "content that looks like a scam"

	case result.Tobacco.Prob > threshold,
		result.RecreationalDrug.Prob > threshold,
		result.Alcohol.Prob > threshold:
		reason = "tobacco, drugs or alcohol"
	}

	if reason != "" {
		// Log the scores behind a rejection so a false positive can be
		// diagnosed without dumping the full response for every clean upload.
		log.Printf("image moderation rejected %s: reason=%q nudity=%.2f/%.2f/%.2f/%.2f gore=%.2f violence=%.2f self-harm=%.2f weapon=%.2f/%.2f offensive=%.2f scam=%.2f tobacco=%.2f drug=%.2f alcohol=%.2f",
			justFilename, reason,
			result.Nudity.SexualActivity, result.Nudity.SexualDisplay, result.Nudity.Erotica, result.Nudity.VerySuggestive,
			result.Gore.Prob, result.Violence.Prob, result.SelfHarm.Prob,
			result.WeaponFirearm, result.WeaponKnife,
			result.Offensive.Prob, result.Scam.Prob,
			result.Tobacco.Prob, result.RecreationalDrug.Prob, result.Alcohol.Prob)
		return true, reason, nil
	}

	return false, "", nil // Safe
}
