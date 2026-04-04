package usecase

import "testing"

func TestIsCloudFunctionEnquiryHandler(t *testing.T) {
	t.Setenv("ENQUIRY_NOTIFICATION_HANDLER", `"cf"`)

	if !isCloudFunctionEnquiryHandler() {
		t.Fatal("expected quoted cf mode to disable the server enquiry watcher")
	}
}

func TestIsCloudFunctionEnquiryHandler_ServerMode(t *testing.T) {
	t.Setenv("ENQUIRY_NOTIFICATION_HANDLER", "server")

	if isCloudFunctionEnquiryHandler() {
		t.Fatal("expected server mode to keep the server enquiry watcher enabled")
	}
}