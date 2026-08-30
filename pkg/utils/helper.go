package utils

import (
	"bytes"
	"encoding/json"
	"errors"
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

// imageModerationEnabled is the process-wide toggle for the Sightengine
// adult/offensive-content check. Set once at startup from config
// (IMAGE_MODERATION_ENABLED) and defaults to true so moderation stays on unless
// a deployment explicitly opts out.
//
// It is a package-level flag rather than an os.Getenv read because the app
// loads .env through viper, which does NOT export those values into the process
// environment — os.Getenv would silently miss a value set only in .env.
var imageModerationEnabled = true

// sightengineAPIUser / sightengineAPISecret are the Sightengine credentials,
// set at startup from config. They were previously hard-coded in this file.
var (
	sightengineAPIUser   string
	sightengineAPISecret string
)

// SetImageModerationEnabled configures the global image-moderation toggle.
// Called at startup from config; may be toggled by tests.
func SetImageModerationEnabled(enabled bool) { imageModerationEnabled = enabled }

// ImageModerationEnabled reports whether uploaded images are sent to Sightengine.
func ImageModerationEnabled() bool { return imageModerationEnabled }

// SetSightengineCredentials configures the Sightengine API credentials.
// Called at startup from config.
func SetSightengineCredentials(apiUser, apiSecret string) {
	sightengineAPIUser = apiUser
	sightengineAPISecret = apiSecret
}

// ErrModerationUnavailable reports that the moderation service could not give a
// verdict (unreachable, bad credentials, quota exceeded, malformed response).
// It is distinct from a genuine "this image is adult content" verdict so the
// caller can fail open on the former while still blocking on the latter.
var ErrModerationUnavailable = errors.New("image moderation unavailable")

// CheckNudity reports whether an image is adult/offensive per Sightengine.
//
// Returns (false, nil) without contacting Sightengine when moderation is
// disabled via IMAGE_MODERATION_ENABLED=false.
//
// A service-side problem is returned wrapped in ErrModerationUnavailable rather
// than as a plain error, so callers can distinguish "the service is broken"
// (allow, log) from "this image violates policy" (block).
func CheckNudity(path string) (bool, error) {
	// Bypass: no external call at all, image treated as safe.
	if !imageModerationEnabled {
		return false, nil
	}

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

	// Add API parameters.
	// IMPORTANT: Sightengine rejects the *entire* request (status: failure)
	// if even one model name here is invalid — verify a name against
	// https://sightengine.com/docs/models before adding it back.
	_ = writer.WriteField("models", "nudity-2.1,gore-2.0,violence,self-harm,weapon,offensive,scam,tobacco,recreational_drug,alcohol")
	_ = writer.WriteField("api_user", sightengineAPIUser)
	_ = writer.WriteField("api_secret", sightengineAPISecret)

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
		return false, fmt.Errorf("%w: send request: %v", ErrModerationUnavailable, err)
	}
	defer res.Body.Close()

	var result ModerationResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("%w: decode response: %v", ErrModerationUnavailable, err)
	}

	// Print the full response for debugging
	fmt.Printf("Sightengine API Response: %+v\n", result)
	fmt.Printf("Response Status: %s\n", result.Status)
	if result.Status == "failure" {
		fmt.Printf("Error Type: %s\n", result.Error.Type)
		fmt.Printf("Error Code: %d\n", result.Error.Code)
		fmt.Printf("Error Message: %s\n", result.Error.Message)
	}

	// No clean success means no verdict either way — report it as unavailable
	// so the caller decides the policy (see handleUpload, which fails open).
	if result.Status != "success" {
		return false, fmt.Errorf(
			"%w: %s (code %d): %s",
			ErrModerationUnavailable, result.Error.Type, result.Error.Code, result.Error.Message,
		)
	}

	const threshold = 0.5

	// sexual_activity/sexual_display are clear indicators of adult content;
	// erotica/very_suggestive are also strong indicators.
	if result.Nudity.SexualActivity > threshold ||
		result.Nudity.SexualDisplay > threshold ||
		result.Nudity.Erotica > threshold ||
		result.Nudity.VerySuggestive > threshold {
		return true, nil
	}

	if result.Gore.Prob > threshold ||
		result.Violence.Prob > threshold ||
		result.SelfHarm.Prob > threshold {
		return true, nil
	}

	if result.WeaponFirearm > threshold || result.WeaponKnife > threshold {
		return true, nil
	}

	if result.Offensive.Prob > threshold ||
		result.Offensive.Nazi > threshold ||
		result.Offensive.Confederate > threshold ||
		result.Offensive.Supremacist > threshold ||
		result.Offensive.Terrorist > threshold {
		return true, nil
	}

	if result.Scam.Prob > threshold {
		return true, nil
	}

	if result.Tobacco.Prob > threshold ||
		result.RecreationalDrug.Prob > threshold ||
		result.Alcohol.Prob > threshold {
		return true, nil
	}

	return false, nil // Safe
}
