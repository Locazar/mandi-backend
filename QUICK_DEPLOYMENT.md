# Quick Deployment & Testing Guide

## 📱 What This Does

This cloud function automatically sends push notifications to mobile users when enquiry status changes. The flow is:

```
1. User updates enquiry in Firestore (via admin/web)
   ↓
2. Firestore triggers Eventarc event
   ↓
3. Cloud Function receives event
   ↓
4. Function checks if monitored fields changed (status, assignedTo, priority, etc.)
   ↓
5. Gets FCM tokens for relevant users from Firestore
   ↓
6. Sends push notification to mobile app
   ↓
7. User receives notification (if app is running or backgrounded)
```

---

## 🚀 One-Command Deployment (Windows)

### Run This in PowerShell:

```powershell
# Navigate to project root
cd C:\Users\HP\Documents\Locazar\mandi-backend

# Run deployment script
.\Deploy-CloudFunction.ps1
```

**That's it!** The script will:
- ✓ Verify gcloud and Go are installed
- ✓ Set up GCP project
- ✓ Create service account (if needed)
- ✓ Enable required APIs
- ✓ Deploy the function
- ✓ Show you the deployment URL and logs

---

## 📊 For Linux/Mac Users:

```bash
cd /path/to/mandi-backend
chmod +x deploy-cloud-function.sh
bash deploy-cloud-function.sh
```

---

## 🧪 Testing Locally (Before Deployment)

If you want to test locally first:

### Step 1: Install Prerequisites
```powershell
# Install Google Cloud SDK: https://cloud.google.com/sdk/docs/install
# Install Go 1.21+: https://golang.org/dl/
# Install Firebase CLI: npm install -g firebase-tools
```

### Step 2: Set Credentials
```powershell
# Set your service account credentials
$env:GOOGLE_APPLICATION_CREDENTIALS = "C:\Users\HP\Documents\Locazar\mandi-backend\locazar-f20b6-c125ff67e902.json"

# Verify
echo $env:GOOGLE_APPLICATION_CREDENTIALS
```

### Step 3: Start Local Server
```powershell
cd cmd/cloudfunctions/enquiry-notification-handler
go run main.go &
```

### Step 4: Test with Sample Event
```powershell
$event = Get-Content test_event.json -Raw
Invoke-WebRequest -Uri "http://localhost:8080/" `
  -Method POST `
  -Headers @{"Content-Type"="application/json"} `
  -Body $event
```

---

## 🔄 Testing in Production (After Deployment)

### Method 1: Via Firestore Console (Recommended)

1. Go to: https://console.firebase.google.com/u/0/project/locazar-f20b6
2. Select **Firestore Database** → **enquiries** collection
3. Find any enquiry and update the `status` field
4. **Result**: Check Cloud Function logs to see if notification was sent

### Method 2: Manual Update Using Firebase CLI
```powershell
# Install Firebase CLI
npm install -g firebase-tools

# Authenticate
firebase login

# Update an enquiry (example)
firebase firestore --project=locazar-f20b6 \
  update documents/enquiries/ENQ-123 \
  "status=resolved" "updatedAt=now"
```

### Method 3: Via gcloud CLI
```powershell
# Query enquiries
gcloud firestore documents list --collection-id=enquiries --project=locazar-f20b6

# Update an enquiry
gcloud firestore documents update enquiries/ENQ-123 \
  --update-data status=resolved,updatedAt=now \
  --project=locazar-f20b6
```

### Method 4: Check Real-Time Logs
```powershell
# Watch function execution logs in real-time
gcloud functions logs read enquiry-notification-handler `
  --gen2 --region=asia-south1 --follow --project=locazar-f20b6
```

---

## 📱 Mobile Push Notifications Setup

For users to receive push notifications, you need:

### 1. FCM Tokens in Firestore
Your mobile app must register FCM tokens in this structure:

```
firestore/
users/
├── {USER_ID}/
│   ├── name: "John Doe"
│   ├── email: "john@example.com"
│   └── fcmTokens/
│       ├── token_1/
│       │   ├── token: "exxJlRl_O5k..." ← FCM token from device
│       │   ├── isActive: true
│       │   ├── platform: "android"
│       │   └── registeredAt: 2024-01-01...
│       └── token_2/
│           ├── token: "d8xZkM9_R4l..."
│           ├── isActive: true
│           ├── platform: "ios"
│           └── registeredAt: 2024-01-01...
```

### 2. Mobile App Code (Firebase Messaging)
```javascript
// Register FCM token from mobile app
firebase.messaging().getToken().then(token => {
  // Save token to Firestore
  firebase.firestore()
    .collection('users')
    .doc(currentUserId)
    .collection('fcmTokens')
    .doc(token)
    .set({
      token: token,
      isActive: true,
      platform: Platform.OS, // 'android', 'ios', or 'web'
      registeredAt: new Date()
    });
});

// Handle foreground notifications
firebase.messaging().onMessage(message => {
  console.log('Notification received:', message);
  // Display notification UI
});

// Messages in background are handled automatically
```

### 3. Test Flow
```
1. User registers FCM token in app
2. Admin updates enquiry status in Firestore
3. Cloud Function detects change
4. Function reads FCM tokens for relevant users
5. Function sends push notification via Firebase Messaging API
6. Device receives and displays notification
```

---

## 🛠️ **Common Use Cases**

### When Status Changes to "Assigned"
```
Firestore Update:
  enquiries/ENQ-123 {
    status: "assigned",
    assignedTo: "ADMIN-001"
  }

Function Action:
  → Detects status & assignedTo changed
  → Finds admin with ID "ADMIN-001"
  → Reads their FCM tokens
  → Sends: "Enquiry #ENQ-123 has been assigned to you"
```

### When Priority Changes
```
Firestore Update:
  enquiries/ENQ-123 {
    priority: "high"
  }

Function Action:
  → Detects priority changed
  → Finds customer user ID from enquiry
  → Reads customer's FCM tokens
  → Sends: "Your enquiry priority has been updated to High"
```

### When Enquiry is Resolved
```
Firestore Update:
  enquiries/ENQ-123 {
    status: "resolved"
  }

Function Action:
  → Detects status changed
  → Finds customer user ID
  → Sends: "Your enquiry has been resolved! 🎉"
```

---

## 📊 Monitoring & Debugging

### View All Logs
```powershell
gcloud functions logs read enquiry-notification-handler `
  --gen2 --region=asia-south1 `
  --limit=100 `
  --project=locazar-f20b6
```

### Filter Logs
```powershell
# Show only errors
gcloud functions logs read enquiry-notification-handler `
  --gen2 --region=asia-south1 `
  --project=locazar-f20b6 | `
  Select-String "ERROR"

# Show only successful sends
gcloud functions logs read enquiry-notification-handler `
  --gen2 --region=asia-south1 `
  --project=locazar-f20b6 | `
  Select-String "sent successfully"
```

### Check Function Metrics
```powershell
# View invocations and performance
gcloud functions describe enquiry-notification-handler `
  --gen2 --region=asia-south1 `
  --project=locazar-f20b6
```

### View Recent Errors
```powershell
# Detailed error messages
gcloud functions logs read enquiry-notification-handler `
  --gen2 --region=asia-south1 `
  --project=locazar-f20b6 | `
  Select-String "FAILED|ERROR|Exception"
```

---

## 🔧 Troubleshooting

### Issue: Function didn't send notification
```powershell
1. Check FCM tokens exist:
   gcloud firestore documents list --collection-id=fcmTokens --project=locazar-f20b6

2. Check function logs:
   gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1 --follow

3. Verify enquiry structure matches expected format
```

### Issue: "No recipients found" error
```
This means:
- Firestore doesn't have a collection "users/{userId}/fcmTokens"
- Or the user ID isn't properly linked to enquiry

Solution:
- Make sure mobile app registers FCM tokens
- Check userId field in enquiry matches user document ID
```

### Issue: Function times out
```powershell
# Increase timeout (currently 60s)
gcloud functions deploy enquiry-notification-handler `
  --gen2 `
  --region=asia-south1 `
  --timeout=120s `
  --project=locazar-f20b6
```

### Issue: Authentication errors
```powershell
# Re-authenticate
gcloud auth login
gcloud auth application-default login

# Verify credentials
gcloud config list
```

---

## 🎯 Deployment Checklist

- [ ] Run deployment script (Deploy-CloudFunction.ps1 or deploy-cloud-function.sh)
- [ ] Wait for "Deployment Complete!" message
- [ ] Verify function appears in GCP Console
- [ ] Test by updating an enquiry in Firestore
- [ ] Check logs to confirm notification was sent
- [ ] Register test FCM tokens in Firestore
- [ ] Verify mobile device receives notification when status changes

---

## 📚 Useful Links

- **GCP Console**: https://console.cloud.google.com/functions?project=locazar-f20b6
- **Firebase Console**: https://console.firebase.google.com/u/0/project/locazar-f20b6
- **Cloud Function Logs**: https://console.cloud.google.com/logs/query?project=locazar-f20b6
- **Firestore Data**: https://console.firebase.google.com/u/0/project/locazar-f20b6/firestore/data
- **gcloud Reference**: https://cloud.google.com/sdk/gcloud/reference/functions

---

## 🔐 Credentials Note

The service account key is already created at:
```
locazar-f20b6-c125ff67e902.json
```

This key is used by the cloud function to authenticate with Firebase services.

---

## ⏭️ What's Next?

1. **Deploy** - Run the deployment script
2. **Test** - Update an enquiry and check logs
3. **Monitor** - Watch for successful notifications
4. **Mobile** - Integrate into your app so users register FCM tokens
5. **Iterate** - Customize notification messages and conditions

---

**Questions?** Check DEPLOYMENT_GUIDE.md for detailed information
