# 📋 Cloud Function Deployment - Complete Summary

## What You Now Have ✅

I've set up the **Enquiry Notification Cloud Function** for you. This will automatically send push notifications to mobile users when enquiry status changes.

---

## 📁 Files Created for You

| File | Purpose |
|------|---------|
| **START_HERE.md** | 👈 **BEGIN HERE** - Quick copy-paste commands |
| **SETUP_INSTRUCTIONS.md** | Step-by-step installation guide |
| **QUICK_DEPLOYMENT.md** | Quick reference and testing guide |
| **DEPLOYMENT_GUIDE.md** | Detailed complete deployment guide |
| **Deploy-CloudFunction.ps1** | PowerShell deployment script |
| **deploy-cloud-function.sh** | Bash deployment script (for Linux/Mac) |
| **test_event.json** | Sample test event |
| **.env** | Configuration file |

---

## 🎯 How It Works

```
Enquiry Updated in Firestore
         ↓
Cloud Function Triggered (Eventarc)
         ↓
Function Detects Status/Priority Change
         ↓
Reads FCM Tokens from Firestore
         ↓
Sends Push Notification via Firebase
         ↓
User Receives Notification on Mobile 📱
```

---

## ⚡ Quick Start (3 Commands)

### 1. **Install Google Cloud SDK** (if not already installed)
Download from: https://cloud.google.com/sdk/docs/install-cloud-sdk
(Make sure to check "Add gcloud CLI to PATH")

### 2. **Authenticate**
```powershell
gcloud init
# Select: locazar-f20b6 project
```

### 3. **Deploy Function**
```powershell
cd C:\Users\HP\Documents\Locazar\mandi-backend

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
```

**Wait 2-5 minutes** ⏳

---

## ✅ After Deployment

### Test It
```powershell
# Update any enquiry status in Firebase Console:
# https://console.firebase.google.com/u/0/project/locazar-f20b6/firestore

# Watch logs
gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1 --follow
```

### Monitor Logs
```powershell
# View last 20 logs
gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1 --limit=20

# See only errors
gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1 | Select-String "ERROR"
```

---

## 📱 For Mobile Users to Get Notifications

Add this to your mobile app (React Native / Flutter / etc):

**JavaScript/React Native:**
```javascript
// Register FCM token
const token = await firebase.messaging().getToken();
firebase.firestore()
  .collection('users')
  .doc(userId)
  .collection('fcmTokens')
  .doc(token)
  .set({
    token,
    isActive: true,
    platform: 'android', // or 'ios'
    registeredAt: new Date()
  });
```

**Flutter:**
```dart
String? token = await FirebaseMessaging.instance.getToken();
FirebaseFirestore.instance
    .collection('users')
    .doc(userId)
    .collection('fcmTokens')
    .doc(token!)
    .set({
      'token': token,
      'isActive': true,
      'platform': 'ios',
      'registeredAt': DateTime.now(),
    });
```

---

## 🎯 Deployment Checklist

```
[ ] Google Cloud SDK installed (gcloud --version works)
[ ] Authenticated (gcloud auth list shows your email)
[ ] Project set to locazar-f20b6 (gcloud config get-value project)
[ ] Deploy command executed
[ ] Wait 2-5 minutes for deployment
[ ] Verify: gcloud functions describe enquiry-notification-handler --gen2 --region=asia-south1
[ ] Test by updating enquiry status in Firebase
[ ] Check logs for success
[ ] Add FCM token registration to mobile app
[ ] Test push notification on mobile device
```

---

## 🔗 Monitoring & Management

| Task | Command |
|------|---------|
| **View Real-Time Logs** | `gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1 --follow` |
| **View Last 50 Logs** | `gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1 --limit=50` |
| **Function Details** | `gcloud functions describe enquiry-notification-handler --gen2 --region=asia-south1` |
| **Increase Memory** | `gcloud functions deploy enquiry-notification-handler --gen2 --region=asia-south1 --memory=1GB` |
| **Increase Timeout** | `gcloud functions deploy enquiry-notification-handler --gen2 --region=asia-south1 --timeout=120s` |
| **Delete Function** | `gcloud functions delete enquiry-notification-handler --gen2 --region=asia-south1` |

---

## 🎛️ Configuration

The cloud function is already configured with:

```
GCP_PROJECT_ID=locazar-f20b6
LOG_LEVEL=INFO  (set to DEBUG for verbose logs)
REGION=asia-south1
MEMORY=512MB
TIMEOUT=60s

MONITORED_FIELDS:
├── status  (e.g., "new" → "assigned" → "in_progress" → "resolved")
├── assignedTo  (admin assignment)
├── priority  (urgency level)
├── resolutionDate
├── closedAt
└── tags
```

**To modify**, update the environment variables in the deploy command.

---

## 🆘 Common Issues & Fixes

| Issue | Solution |
|-------|----------|
| "gcloud not found" | Restart PowerShell or reinstall gcloud SDK |
| "Authentication failed" | Run `gcloud auth login` |
| "Project not found" | Run `gcloud config set project locazar-f20b6` |
| "APIs not enabled" | Enable them with gcloud: See SETUP_INSTRUCTIONS.md |
| "No recipients found" in logs | Register FCM tokens in Firebase (mobile app code) |
| Function takes too long | Increase memory to 1GB or timeout to 120s |
| "Permission denied" | You may need Editor role in GCP project |

---

## 📊 What Gets Sent as Notification

When an enquiry field changes, users receive:

**Example 1: Status Changed**
```
Title: "Enquiry Update"
Body: "Your enquiry ENQ-123 status is now: resolved"
```

**Example 2: Assigned**
```
Title: "Enquiry Assigned"
Body: "Your enquiry has been assigned to John Doe"
```

**Example 3: Priority Changed**
```
Title: "Priority Changed"
Body: "Your enquiry priority is now: High"
```

*(Exact messages depend on your payload configuration)*

---

## 🔐 Security

- ✅ Uses service account authentication
- ✅ Deploys to private Google Cloud
- ✅ Firestore security rules apply
- ✅ Only Firebase Console users can update enquiries
- ✅ Tokens are validated by Firebase before sending

---

## 📈 Next Steps (Recommended)

### Phase 1: Verify Deployment ✅
1. Run deploy command
2. Wait for completion
3. Check `gcloud functions logs read ...`

### Phase 2: Test Notifications 🧪
1. Add FCM token to Firestore (mobile app)
2. Update enquiry status in Firebase Console
3. Verify notification arrives
4. Check cloud function logs

### Phase 3: Production Ready 🚀
1. Customize notification messages (if needed)
2. Add more monitored fields (if needed)
3. Set up monitoring/alerts
4. Document for your team

---

## 📚 Full Documentation Links

- **START_HERE.md** - Copy-paste commands (easiest)
- **SETUP_INSTRUCTIONS.md** - Detailed step-by-step
- **QUICK_DEPLOYMENT.md** - Quick reference
- **DEPLOYMENT_GUIDE.md** - Complete guide with all options

---

## 🎓 Learning Resources

- [Google Cloud Functions Docs](https://cloud.google.com/functions/docs)
- [Firebase Cloud Messaging](https://firebase.google.com/docs/cloud-messaging)
- [Firestore Events](https://firebase.google.com/docs/firestore/other-products/eventarc)
- [gcloud CLI Reference](https://cloud.google.com/sdk/gcloud/reference/functions)

---

## 💡 Key Files in Your Project

```
cmd/cloudfunctions/enquiry-notification-handler/
├── main.go                 ← Cloud Function entry point
├── go.mod                  ← Go dependencies
├── Dockerfile              ← Container config
└── test_event.json         ← Sample test event

pkg/
├── service/notification/
│   ├── fcm_service.go      ← Firebase Cloud Messaging
│   └── payload_builder.go  ← Notification messages
└── utils/firestore/
    ├── parser.go           ← Parse Firestore events
    ├── comparator.go       ← Detect field changes
    └── event_handler.go    ← Process events
```

---

## 🎯 Success Metrics

After deployment, you'll know it's working when:

✅ **Deployment**: Command completes with "completed successfully"
✅ **Logs**: `gcloud functions logs read` shows no errors
✅ **Function**: `gcloud functions describe` shows STATUS = ACTIVE
✅ **Testing**: Update enquiry → logs show "sent successfully"
✅ **Mobile**: User receives notification when status changes

---

## 📞 Support

If you encounter issues:

1. **Check Logs**: `gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1`
2. **Restart**: Delete and redeploy the function
3. **Permissions**: Verify you have Editor role in GCP project
4. **Firestore**: Verify enquiry and user collections exist with proper schema

---

## ✨ You're Ready!

**Next Action**: Open **START_HERE.md** and follow the commands. The entire deployment takes about 5-10 minutes.

**Questions**: Refer to SETUP_INSTRUCTIONS.md or DEPLOYMENT_GUIDE.md for detailed information.

---

**Status**: ✅ All files prepared and ready for deployment
**Project**: locazar-f20b6
**Region**: asia-south1
**Runtime**: Go 1.21
**Trigger**: Firestore enquiries collection updates
