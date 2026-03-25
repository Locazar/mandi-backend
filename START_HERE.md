# 🚀 QUICK START - Copy & Paste Commands

**Do this in PowerShell (Run as Administrator)**

---

## ✅ Step 1: Check if gcloud is installed

```powershell
gcloud --version
```

**Result**:
- ✓ If it shows version → Skip to Step 3
- ✗ If "gcloud not found" → Do Step 2

---

## 📥 Step 2: Install Google Cloud SDK (If needed)

```powershell
# Download and install
# Go to: https://cloud.google.com/sdk/docs/install-cloud-sdk
# Download "Windows 64-bit" installer
# Run installer with "Add gcloud CLI to PATH" CHECKED ✓
# Restart PowerShell

# Verify installation
gcloud --version
```

---

## 🔑 Step 3: Authenticate & Set Project

```powershell
# Initialize gcloud
gcloud init

# When prompted:
# 1. Press 'Y' to log in (opens browser)
# 2. Log in with your Google account
# 3. Select project: locazar-f20b6
# 4. Press 'Y' to configure for Cloud Functions

# Verify
gcloud config list
```

---

## ⚙️ Step 4: Enable Required APIs

Copy-paste this entire block:

```powershell
gcloud services enable cloudfunctions.googleapis.com
gcloud services enable cloudbuild.googleapis.com
gcloud services enable firestore.googleapis.com
gcloud services enable eventarc.googleapis.com
gcloud services enable logging.googleapis.com
gcloud services enable pubsub.googleapis.com

# Verify
gcloud services list --enabled | Select-String "functions"
```

---

## 🚀 Step 5: Deploy Cloud Function

Navigate to project directory and deploy:

```powershell
# Change to project directory
cd C:\Users\HP\Documents\Locazar\mandi-backend

# Deploy (this takes 2-5 minutes)
gcloud functions deploy enquiry-notification-handler `
  --gen2 `
  --runtime=go121 `
  --region=asia-south1 `
  --source=cmd/cloudfunctions/enquiry-notification-handler `
  --entry-point=ProcessEnquiryUpdate `
  --memory=512MB `
  --timeout=60 `
  --trigger-event-filters="type=google.cloud.firestore.document.v1.updated" `
  --trigger-event-filters="database=(default)" `
  --trigger-event-filters-path-pattern="document=enquiries/{enquiryId}" `
  --set-env-vars="GCP_PROJECT_ID=locazar-f20b6,LOG_LEVEL=INFO,MONITORED_FIELDS=status,assignedTo,priority"

# Watch for: "Deploying function..." then "deploymentsPatch deployment [...] completed successfully"
```

---

## ✅ Step 6: Verify Deployment

```powershell
# Check function is deployed
gcloud functions describe enquiry-notification-handler `
  --gen2 `
  --region=asia-south2

# Should show: STATUS = ACTIVE
```

---

## 🧪 Step 7: Test the Function

### Option A: Via Firestore Console (Easiest)

1. Open: https://console.firebase.google.com/u/0/project/locazar-f20b6
2. Click **Firestore Database** → **enquiries** collection
3. Click any enquiry document
4. Edit the `status` field (change it to "resolved")
5. Click **Save**
6. Run this to see if notification was sent:

```powershell
# Watch logs (Ctrl+C to exit)
gcloud functions logs read enquiry-notification-handler `
  --gen2 `
  --region=asia-south1 `
  --follow --limit=20
```

### Option B: Via Command Line

```powershell
# Find an enquiry
gcloud firestore documents list --collection-id=enquiries --project=locazar-f20b6

# Update its status (replace ENQ-123 with actual ID)
gcloud firestore documents update enquiries/ENQ-123 `
  --update-data status=resolved `
  --project=locazar-f20b6

# Check logs
gcloud functions logs read enquiry-notification-handler `
  --gen2 `
  --region=asia-south1 `
  --limit=10
```

---

## 🎯 Expected Log Output

If successful, you'll see logs like:

```
Starting ProcessEnquiryUpdate
Received event: event-123
Parsing event...
Detected changes: status, assignedTo
Finding FCM tokens for users...
Sent notification to 5 devices
Processing complete: SUCCESS
```

If there's an issue, you'll see:

```
ERROR: No recipients found
ERROR: Failed to send FCM notification
ERROR: Field parsing failed
```

---

## 📱 Next: Setup Mobile Notifications

For users to receive push notifications, add this to your mobile app:

### React Native / JavaScript:

```javascript
// Get and register FCM token
firebase.messaging().getToken()
  .then(token => {
    // Save to Firestore
    firebase.firestore()
      .collection('enquiry')
      .doc(userId)
      .collection('fcmTokens')
      .doc(token)
      .set({
        token: token,
        isActive: true,
        platform: 'android', // or 'ios'
        registeredAt: new Date()
      });
  });

// Handle foreground notifications
firebase.messaging().onMessage(message => {
  console.log('Notification:', message);
  // Show notification UI
});
```

### Flutter:

```dart
// Get and register FCM token
FirebaseMessaging messaging = FirebaseMessaging.instance;
String? token = await messaging.getToken();

// Save to Firestore
FirebaseFirestore.instance
    .collection('enquiry')
    .doc(userId)
    .collection('fcmTokens')
    .doc(token!)
    .set({
      'token': token,
      'isActive': true,
      'platform': 'ios', // or 'android'
      'registeredAt': DateTime.now(),
    });
```

---

## 🎨 How It Works

```
User Updates Enquiry
        ↓
Firestore Triggers Event (via Eventarc)
        ↓
Cloud Function Activates
        ↓
Function Checks Changed Fields
        ↓
If "status" or "assignedTo" Changed:
        ↓
Get FCM Tokens from Firestore
        ↓
Send Firebase Messaging
        ↓
Mobile Devices Receive Notification 📱
```

---

## 🔍 Useful Monitoring Commands

```powershell
# Watch logs in real-time
gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1 --follow

# See last 50 logs
gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1 --limit=50

# See only errors
gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1 | Select-String "ERROR"

# See function metrics
gcloud functions describe enquiry-notification-handler --gen2 --region=asia-south1 --format=json

# Update function memory (if too slow)
gcloud functions deploy enquiry-notification-handler --gen2 --region=asia-south1 --memory=1GB

# Delete function (if needed)
gcloud functions delete enquiry-notification-handler --gen2 --region=asia-south1
```

---

## 🆘 Troubleshooting

### "gcloud not found"
- Restart PowerShell (close and reopen)
- Check gcloud installed: https://cloud.google.com/sdk/install

### "Authentication failed"
```powershell
gcloud auth login
```

### "Project not found"
```powershell
gcloud config set project locazar-f20b6
gcloud config list
```

### "APIs not enabled"
- Run Step 4 again (enable APIs)
- Wait 1-2 minutes, then deploy

### "No recipients found" in logs
- Add FCM tokens to Firestore first
- See "Setup Mobile Notifications" section

### Function deployment takes too long
- Normal takes 2-5 minutes
- Check for errors: `gcloud functions log-read enquiry-notification-handler --gen2 --region=asia-south1`

---

## 🎯 Summary of What You Get

After completing these steps:

✅ Cloud Function deployed and active
✅ Listens to Firestore enquiry changes
✅ Sends Firebase notifications automatically
✅ Users get push notifications on mobile
✅ Real-time logging and monitoring
✅ Fully production-ready

---

## 📚 Full Documentation

- [Setup Instructions](SETUP_INSTRUCTIONS.md) - Detailed setup
- [Quick Deployment](QUICK_DEPLOYMENT.md) - Quick reference
- [Deployment Guide](DEPLOYMENT_GUIDE.md) - Complete guide

---

## ✨ You're All Set!

Run the deploy command in Step 5 and monitor the logs. When you see "completed successfully", your cloud function is live! 🎉
