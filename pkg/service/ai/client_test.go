package ai

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// writeServiceResponse encodes a standard ai-service envelope as the HTTP response.
func writeServiceResponse(t *testing.T, w http.ResponseWriter, resp ServiceResponse) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		t.Fatalf("failed to encode response: %v", err)
	}
}

func TestDetectObjects(t *testing.T) {
	var gotPath string
	var gotReq ObjectDetectionRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		writeServiceResponse(t, w, ServiceResponse{
			Success: true,
			Message: "Object detection completed",
			Data: ObjectDetectionResponse{
				Objects: []DetectedObject{
					{Label: "bottle", Category: "Beverages", Color: "green", Material: "glass", Brand: "Acme", Confidence: 0.91},
				},
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.DetectObjects("ZmFrZQ==", true)
	if err != nil {
		t.Fatalf("DetectObjects returned error: %v", err)
	}

	if gotPath != "/api/ai/detect-objects" {
		t.Errorf("request path = %q, want /api/ai/detect-objects", gotPath)
	}
	if !gotReq.DetectAll || gotReq.ImageBase64 != "ZmFrZQ==" {
		t.Errorf("request body = %+v, want DetectAll=true ImageBase64=ZmFrZQ==", gotReq)
	}
	if len(resp.Objects) != 1 {
		t.Fatalf("got %d objects, want 1", len(resp.Objects))
	}
	if obj := resp.Objects[0]; obj.Label != "bottle" || obj.Confidence != 0.91 {
		t.Errorf("object = %+v, want label=bottle confidence=0.91", obj)
	}
}

func TestVerifyImage(t *testing.T) {
	var gotPath string
	var gotReq ImageVerificationRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		writeServiceResponse(t, w, ServiceResponse{
			Success: true,
			Message: "Image verification completed",
			Data:    ImageVerificationResponse{Matches: true, Confidence: 0.8, Reason: "looks right"},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.VerifyImage("ZmFrZQ==", "is it a red shirt?")
	if err != nil {
		t.Fatalf("VerifyImage returned error: %v", err)
	}

	if gotPath != "/api/ai/verify-image" {
		t.Errorf("request path = %q, want /api/ai/verify-image", gotPath)
	}
	if gotReq.Prompt != "is it a red shirt?" || gotReq.ImageBase64 != "ZmFrZQ==" {
		t.Errorf("request body = %+v", gotReq)
	}
	if !resp.Matches || resp.Confidence != 0.8 || resp.Reason != "looks right" {
		t.Errorf("response = %+v, want Matches=true Confidence=0.8", resp)
	}
}

func TestDetectObjects_ServiceFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 200 OK but the envelope reports failure — the client must surface this as an error.
		writeServiceResponse(t, w, ServiceResponse{Success: false, Error: "model unavailable"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	if _, err := c.DetectObjects("ZmFrZQ==", false); err == nil {
		t.Fatal("expected an error when success=false, got nil")
	}
}

func TestDetectObjectsFromPath(t *testing.T) {
	content := []byte{0xFF, 0xD8, 0xFF, 0x00, 0x01, 0x02, 0x03}
	path := filepath.Join(t.TempDir(), "img.jpg")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp image: %v", err)
	}
	want := base64.StdEncoding.EncodeToString(content)

	var gotReq ObjectDetectionRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		writeServiceResponse(t, w, ServiceResponse{Success: true, Data: ObjectDetectionResponse{}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	if _, err := c.DetectObjectsFromPath(path, false); err != nil {
		t.Fatalf("DetectObjectsFromPath returned error: %v", err)
	}
	if gotReq.ImageBase64 != want {
		t.Errorf("ImageBase64 = %q, want %q", gotReq.ImageBase64, want)
	}
}

func TestVerifyImageFromPath_MissingFile(t *testing.T) {
	c := NewClient("http://127.0.0.1:0")
	if _, err := c.VerifyImageFromPath(filepath.Join(t.TempDir(), "does-not-exist.jpg"), "x"); err == nil {
		t.Fatal("expected an error for a missing image file, got nil")
	}
}
