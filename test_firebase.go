package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// Simple test to verify Firebase Admin SDK can communicate with Firebase
func main() {
	credFile := "locazar-f20b6-c125ff67e902.json"

	// Read credentials
	credData, err := os.ReadFile(credFile)
	if err != nil {
		log.Fatalf("Failed to read credentials: %v", err)
	}

	var cred struct {
		ProjectID   string `json:"project_id"`
		ClientEmail string `json:"client_email"`
	}
	json.Unmarshal(credData, &cred)
	fmt.Printf("Project: %s\n", cred.ProjectID)
	fmt.Printf("Service Account: %s\n", cred.ClientEmail)

	// Initialize Firebase
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opt := option.WithCredentialsFile(credFile)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}
	fmt.Println("✓ Firebase app initialized")

	// Get messaging client
	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("Failed to get messaging client: %v", err)
	}
	fmt.Println("✓ Messaging client created")

	// Test with a known valid token from your database
	// Replace with an actual token from your fcm_tokens table
	testToken := "c_t2aVnlTr2fmDsmNtdUUQ:APA91bHDymCxxx"

	fmt.Printf("\nTesting Send with token: %s...\n", testToken[:30]+"...")

	msg := &messaging.Message{
		Notification: &messaging.Notification{
			Title: "Firebase Test",
			Body:  "Testing Firebase connectivity",
		},
		Token: testToken,
	}

	resp, err := client.Send(ctx, msg)
	if err != nil {
		fmt.Printf("✗ Send failed:\n%v\n", err)
		fmt.Println("\nDiagnostics:")
		fmt.Println("1. Check FCM API is enabled in Google Cloud Console")
		fmt.Println("2. Verify service account has Firebase Cloud Messaging Admin role")
		fmt.Println("3. Ensure token is valid (from your own Firebase app)")
		fmt.Println("4. Check GCP project quota and billing")
		return
	}

	fmt.Printf("✓ Message sent successfully!\n")
	fmt.Printf("  Message ID: %s\n", resp)
}
