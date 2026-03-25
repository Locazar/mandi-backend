# Firebase FCM 404 Error - Troubleshooting Guide

## Problem
You're receiving `404 Not Found` error on Firebase's `/batch` endpoint when trying to send notifications via `SendMulticast`.

```
SendMulticast err-- unexpected http response with status: 404
Error: The requested URL `/batch` was not found on this server.
```

## Causes

This error typically indicates one of these issues:

### 1. **Credentials File Not Found or Invalid Path**
- The Firebase credentials JSON file path is incorrect
- The file doesn't exist at the specified location
- Relative path isn't resolving correctly

### 2. **Invalid or Expired Service Account Credentials**
- The service account key is invalid
- The key has been regenerated in Firebase Console
- The service account has been deleted

### 3. **Insufficient Permissions**
- Service account doesn't have FCM permissions
- Missing "Firebase Cloud Messaging Admin" role
- Missing "Service Accounts Token Creator" role

### 4. **Wrong Firebase Project**
- Credentials are for a different Firebase project
- Project ID mismatch between credentials and configuration

### 5. **Network/Firewall Issues**
- Firewall blocking access to `fcm.googleapis.com`
- Proxy or corporate network filtering

## Solutions

### Step 1: Verify Credentials File

```bash
# Check if file exists
ls -la locazar-f20b6-f024a4597849.json

# View the contents (safely)
cat locazar-f20b6-f024a4597849.json | jq '{type, project_id, client_email}'
```

**Expected output:**
```json
{
  "type": "service_account",
  "project_id": "mandi-backend-379522",
  "client_email": "firebase-adminsdk-xxxxx@mandi-backend-379522.iam.gserviceaccount.com"
}
```

### Step 2: Set Absolute Path (Recommended)

Instead of using a relative path, set an absolute path:

```bash
# In your .env file
export FIREBASE_CREDENTIALS_FILE=/path/to/mandi-backend/locazar-f20b6-f024a4597849.json

# Or when running locally
export FIREBASE_CREDENTIALS_FILE=$(pwd)/locazar-f20b6-f024a4597849.json
go run ./cmd/api/main.go
```

### Step 3: Verify Firebase Project Setup

1. Go to [Firebase Console](https://console.firebase.google.com/)
2. Select your project: `mandi-backend-379522`
3. Go to **Cloud Messaging** tab
4. Verify it shows configuration correctly
5. Check **Service Accounts** (Project Settings → Service Accounts)
6. The account should have these roles:
   - `Firebase Cloud Messaging Admin`
   - `Service Accounts Token Creator`

### Step 4: Regenerate Service Account Key (If Needed)

1. Go to **Project Settings** → **Service Accounts**
2. Click on your service account email
3. Go to **Keys** tab
4. If old keys exist, delete them
5. Click **Add Key** → **Create new key**
6. Choose **JSON** format
7. Replace your `locazar-f20b6-f024a4597849.json` file

### Step 5: Test Connectivity

```go
// Test basic Firebase connectivity
ctx := context.Background()
opt := option.WithCredentialsFile("/absolute/path/to/credentials.json")
app, err := firebase.NewApp(ctx, nil, opt)
if err != nil {
    log.Fatalf("Failed to initialize app: %v", err)
}

client, err := app.Messaging(ctx)
if err != nil {
    log.Fatalf("Failed to get client: %v", err)
}

// Test with a single valid FCM token
response, err := client.SendMulticast(ctx, &messaging.MulticastMessage{
    Tokens: []string{"valid_fcm_token_here"},
    Notification: &messaging.Notification{
        Title: "Test",
        Body:  "Test notification",
    },
})
log.Printf("Response: %v, Error: %v", response, err)
```

## Code Changes Made

The updated `pkg/usecase/fcm_token.go` now:

✓ **Initializes Firebase once** (singleton pattern) instead of on every request
✓ **Verifies credentials file exists** before attempting initialization  
✓ **Resolves absolute paths** properly
✓ **Adds proper error handling** with meaningful error messages
✓ **Uses context timeouts** to prevent hanging
✓ **Logs detailed send results** including which tokens failed

### Key Improvements

**Before (Problematic):**
```go
// Creates new Firebase app on EVERY request - inefficient
app, err := firebase.NewApp(ctx, nil, opt)
client, err := app.Messaging(ctx)
msg, err := client.SendMulticast(ctx, message)
```

**After (Optimized):**
```go
// Singleton client - reused for all requests
client, err := u.getMessagingClient(ctx)
response, err := client.SendMulticast(sendCtx, message)
log.Printf("Sent %d, Failed %d", response.SuccessCount, response.FailureCount)
```

## Debugging Steps

1. **Check logs** for the exact credentials file path being used:
   ```
   Initializing Firebase app with credentials: /path/to/file
   ```

2. **Enable verbose logging** in your application to see exact error from Firebase API

3. **Test with curl** (if you have the credentials ready):
   ```bash
   # Get OAuth token from credentials
   # Send test request to https://fcm.googleapis.com/v1/projects/{PROJECT_ID}/messages:send
   ```

4. **Check Firebase Security Quotas**:
   - Go to Firebase Console → Functions → Add-ons
   - Verify usage isn't exceeding free tier limits

## Prevention

- Always use **absolute paths** for credentials files
- Store credentials in environment variables, not hardcoded
- Implement **credential validation on startup** (done in updated code)
- Use **dependency injection** to pass the client once (recommended next step)
- Add **retry logic** for transient failures
- **Monitor FCM** metrics in Firebase Console

## Additional Resources

- [Firebase Admin SDK - Go](https://firebase.google.com/docs/admin/setup)
- [FCM Troubleshooting](https://firebase.google.com/docs/cloud-messaging/troubleshooting)
- [Firebase Cloud Messaging API](https://firebase.google.com/docs/reference/admin/go/firebase.google.com/go/v4/messaging)
