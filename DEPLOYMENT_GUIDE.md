# Cloud Function Deployment & Testing Guide

## Overview
This guide will help you deploy and test the **Enquiry Notification Cloud Function** that sends push notifications to mobile devices when enquiry status changes.

**Project**: locazar-f20b6  
**Region**: asia-south1  
**Function Name**: enquiry-notification-handler

---

## 🔧 Phase 1: Local Setup & Testing

### Step 1: Install Prerequisites

#### On Windows:
```powershell
# 1. Install Go 1.21+ (if not already installed)
# Download from: https://golang.org/dl/
go version  # Verify installation

# 2. Install Google Cloud SDK
# Download: https://cloud.google.com/sdk/docs/install

# 3. Install Firebase CLI
npm install -g firebase-tools

# 4. Install gcloud components
gcloud components install cloud-functions-emulator
```

#### Verify Installation:
```powershell
go version                    # Should show Go 1.21+
gcloud version                # Should show up-to-date
firebase --version            # Should show latest version
```

### Step 2: Configure GCP Authentication

```powershell
# Login to Google Cloud
gcloud auth login

# Set default project
gcloud config set project locazar-f20b6

# Create/obtain service account key (needed for Firebase)
gcloud iam service-accounts keys create key.json `
  --iam-account=firebase-adminsdk@locazar-f20b6.iam.gserviceaccount.com

# Set credentials environment variable
$env:GOOGLE_APPLICATION_CREDENTIALS = "C:\Users\HP\Documents\Locazar\mandi-backend\locazar-f20b6-c125ff67e902.json"

# For persistent setup, add to PowerShell profile or use:
[Environment]::SetEnvironmentVariable('GOOGLE_APPLICATION_CREDENTIALS', 'C:\Users\HP\Documents\Locazar\mandi-backend\locazar-f20b6-c125ff67e902.json', [EnvironmentVariableTarget]::User)
```

### Step 3: Download Dependencies

```powershell
cd cmd/cloudfunctions/enquiry-notification-handler
go mod download
go mod tidy
```

### Step 4: Start Local Development Server

There are two ways to test locally:

#### Option A: Using Functions Framework (Recommended)

```powershell
# In cmd/cloudfunctions/enquiry-notification-handler directory
go run main.go

# Server starts on http://localhost:8080
# You should see: [INFO] Functions-framework listening on :8080
```

#### Option B: Using Local Emulator

```powershell
# Start the functions emulator
functions-framework --target=ProcessEnquiryUpdate --debug --port=8080
```

### Step 5: Test with Sample Event

#### In a NEW PowerShell terminal:

```powershell
# Navigate to the function directory
cd cmd/cloudfunctions/enquiry-notification-handler

# Send test event
$event = Get-Content test_event.json
Invoke-WebRequest -Uri "http://localhost:8080/" `
  -Method POST `
  -Headers @{"Content-Type"="application/json"} `
  -Body $event
```

**Expected Response**: 
- `200 OK` with success message (if all goes well)
- Check the server terminal for detailed logs

#### Test Different Scenarios:

```powershell
# Test 1: Status Change Only
$test1 = @{
  id = "test-1"
  data = @{
    value = @{
      name = "projects/locazar-f20b6/databases/(default)/documents/enquiries/doc1"
      fields = @{
        status = @{stringValue = "resolved"}
        userId = @{stringValue = "USER-123"}
      }
      updateTime = "2024-01-02T12:00:00Z"
    }
    oldValue = @{
      name = "projects/locazar-f20b6/databases/(default)/documents/enquiries/doc1"
      fields = @{
        status = @{stringValue = "in_progress"}
      }
      updateTime = "2024-01-02T11:00:00Z"
    }
    updateMask = @{fieldPaths = @("status")}
  }
} | ConvertTo-Json -Depth 10

Invoke-WebRequest -Uri "http://localhost:8080/" `
  -Method POST `
  -Headers @{"Content-Type"="application/json"} `
  -Body $test1

# Test 2: Multiple Field Changes
$test2 = @{
  id = "test-2"
  data = @{
    value = @{
      name = "projects/locazar-f20b6/databases/(default)/documents/enquiries/doc2"
      fields = @{
        status = @{stringValue = "assigned"}
        assignedTo = @{stringValue = "ADMIN-TEAM"}
        priority = @{stringValue = "high"}
        userId = @{stringValue = "USER-456"}
      }
      updateTime = "2024-01-02T12:00:00Z"
    }
    oldValue = @{
      name = "projects/locazar-f20b6/databases/(default)/documents/enquiries/doc2"
      fields = @{
        status = @{stringValue = "new"}
        assignedTo = @{nullValue = "NULL_VALUE"}
        priority = @{stringValue = "low"}
      }
      updateTime = "2024-01-02T11:00:00Z"
    }
    updateMask = @{fieldPaths = @("status", "assignedTo", "priority")}
  }
} | ConvertTo-Json -Depth 10

Invoke-WebRequest -Uri "http://localhost:8080/" `
  -Method POST `
  -Headers @{"Content-Type"="application/json"} `
  -Body $test2
```

---

## 🚀 Phase 2: Deploy to GCP Cloud Functions

### Step 1: Prepare for Deployment

```powershell
cd cmd/cloudfunctions/enquiry-notification-handler

# Verify files
ls -Path Dockerfile, main.go, go.mod, cloudbuild.yaml

# Make sure .env is configured (already done)
cat .env
```

### Step 2: Deploy via gcloud CLI

```powershell
# Option A: Using bash script (Windows with Git Bash or WSL)
bash deploy.sh locazar-f20b6 asia-south1

# Option B: Using PowerShell (direct gcloud command)
gcloud functions deploy enquiry-notification-handler `
  --gen2 `
  --runtime=go121 `
  --region=asia-south1 `
  --source=. `
  --entry-point=ProcessEnquiryUpdate `
  --trigger-event-filters="type=google.cloud.firestore.document.v1.updated" `
  --trigger-event-filters="database=(default)" `
  --trigger-event-filters-path-pattern="document=enquiries/{enquiryId}" `
  --memory=512MB `
  --timeout=60s `
  --service-account=firebase-adminsdk@locazar-f20b6.iam.gserviceaccount.com `
  --set-env-vars="GCP_PROJECT_ID=locazar-f20b6,LOG_LEVEL=INFO,MONITORED_FIELDS=status,assignedTo,priority"

# Wait for deployment (typically 2-5 minutes)
# You'll see: "Deploying function..."
# Then: "deploymentsPatch deployment [enquiry-notification-handler] completed successfully"
```

### Step 3: Verify Deployment

```powershell
# View function details
gcloud functions describe enquiry-notification-handler --gen2 --region=asia-south1

# View function logs
gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1

# Monitor real-time logs
gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1 --follow
```

### Step 4: Test in Production

Once deployed to GCP, test by:

#### Method 1: Trigger via Firestore (Real Test)
```powershell
# Update an enquiry document in Firestore
# This will trigger the cloud function automatically via Eventarc
# Check Cloud Function logs to see the execution

gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1 --follow
```

#### Method 2: Manual HTTP Trigger (If enabled)
```powershell
# Get the function URL
$FUNCTION_URL = (gcloud functions describe enquiry-notification-handler `
  --gen2 --region=asia-south1 --format="value(serviceConfig.uri)")

# Send test event (requires authentication)
Invoke-WebRequest -Uri $FUNCTION_URL `
  -Method POST `
  -Headers @{
    "Content-Type"="application/json"
    "Authorization"="Bearer $(gcloud auth print-identity-token)"
  } `
  -Body (Get-Content test_event.json)
```

---

## 📱 Testing Mobile Notifications

### Prerequisites for End-to-End Mobile Testing:

1. **FCM Tokens**: Users must have registered FCM tokens in Firestore
2. **Firebase Config**: Firebase must be properly configured in your mobile app
3. **Device**: Android or iOS device with your app installed

### Step 1: Ensure FCM Tokens are in Firestore

```
Firestore Structure:
firestore/
├── users/
│   └── {userId}/
│       ├── name: "John Doe"
│       ├── email: "john@example.com"
│       └── fcmTokens/
│           ├── {tokenId1}
│           │   ├── token: "exxJlRl_O5k..."
│           │   ├── isActive: true
│           │   └── platform: "android"
│           └── {tokenId2}
│               ├── token: "d8xZkM9_R4l..."
│               ├── isActive: true
│               └── platform: "ios"
```

**To register FCM tokens from mobile app**:
```javascript
// Android/iOS mobile app code
firebase.messaging().getToken().then(token => {
  firebase.firestore().collection('users').doc(userId)
    .collection('fcmTokens').doc(token).set({
      token: token,
      isActive: true,
      platform: 'android', // or 'ios'
      registeredAt: new Date()
    });
});
```

### Step 2: Update Enquiry and Watch Notifications

1. Open your mobile app
2. Go to Firebase Console
3. Navigate to Firestore → enquiries collection
4. Update any enquiry's `status` field or `assignedTo` field
5. **Result**: Mobile user should receive a push notification (if app is in foreground or background)

### Step 3: Check Notification History

```powershell
# Query notification sending logs
gcloud functions logs read enquiry-notification-handler `
  --gen2 --region=asia-south1 --limit=50
```

---

## 🐛 Troubleshooting

### Issue: "Failed to authenticate"
```powershell
# Re-authenticate with Google Cloud
gcloud auth login
gcloud auth application-default login
```

### Issue: "Function deployment failed"
```powershell
# Check errors
gcloud functions deploy enquiry-notification-handler --gen2 --region=asia-south1

# View build logs
gcloud builds log $(gcloud builds list --filter "FAILURE_MESSAGE!='' AND substitutions._FUNCTION_NAME=enquiry-notification-handler" --format='value(ID)' --limit=1)
```

### Issue: "No notifications received"

1. **Check FCM tokens exist**:
   ```bash
   firebase firestore --project=locazar-f20b6
   > db.collection('users').doc('USER-001').collection('fcmTokens').get()
   ```

2. **Check Cloud Function logs**:
   ```powershell
   gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1
   ```

3. **Verify database schema**:
   ```bash
   # Make sure enquiries collection exists and has documents
   firebase firestore --project=locazar-f20b6
   > db.collection('enquiries').get()
   ```

### Issue: "Function is slow / timing out"

Increase timeout:
```powershell
gcloud functions deploy enquiry-notification-handler `
  --gen2 --region=asia-south1 `
  --timeout=120s
```

---

## 📊 Monitoring & Logs

### View Real-Time Logs
```powershell
gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1 --follow
```

### Filter Logs by Level
```powershell
# View errors only
gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1 --limit=50 | Select-String "ERROR"

# View all debug logs
gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1 --limit=100 | Select-String "DEBUG"
```

### Check Metrics
```powershell
# View function execution metrics
gcloud monitoring metrics-descriptors list --filter="metric.type:cloudfunctions*"
```

---

## ✅ Deployment Checklist

- [x] Prerequisites installed (Go, gcloud, Firebase CLI)
- [x] GCP authentication configured
- [x] .env file created with project ID
- [x] Local testing completed successfully
- [x] Function deployed to asia-south1
- [ ] Eventarc trigger configured
- [ ] FCM tokens registered in Firestore
- [ ] Mobile app receives notifications
- [ ] Logs and monitoring configured

---

## 🔗 Useful Commands Reference

```powershell
# List all deployed functions
gcloud functions list --gen2 --region=asia-south1

# Update function configuration
gcloud functions update enquiry-notification-handler --gen2 --region=asia-south1 --memory=1GB

# Delete function (if needed)
gcloud functions delete enquiry-notification-handler --gen2 --region=asia-south1

# View function URL (for HTTP trigger)
gcloud functions describe enquiry-notification-handler --gen2 --region=asia-south1 --format="value(serviceConfig.uri)"

# Set secret environment variables
gcloud functions deploy enquiry-notification-handler --gen2 --region=asia-south1 --update-secrets="NOTIFICATION_SECRET=my-secret:latest"
```

---

## 📖 Additional Resources

- **Firebase Console**: https://console.firebase.google.com/u/0/project/locazar-f20b6
- **GCP Console**: https://console.cloud.google.com/functions?project=locazar-f20b6
- **Cloud Functions Docs**: https://cloud.google.com/functions/docs
- **Firebase FCM Docs**: https://firebase.google.com/docs/cloud-messaging
- **Go Functions Framework**: https://github.com/GoogleCloudPlatform/functions-framework-go

---

**Next Steps**: Start with Phase 1 (Local Testing). Once verified, proceed to Phase 2 (GCP Deployment).
