# ⚙️ Complete Setup & Installation Guide

## Prerequisites Check

Before deploying, ensure you have these installed:

### 1. Google Cloud SDK (gcloud CLI) ✓ **REQUIRED**
This is the main tool needed for deployment.

#### Install on Windows:
```powershell
# Option A: Using Chocolatey (if installed)
choco install google-cloud-sdk

# Option B: Direct Download (Recommended)
# 1. Download installer: https://dl.google.com/dl/cloudsdk/channels/rapid/GoogleCloudSDKInstaller.exe
# 2. Run the installer and follow the wizard
# 3. Choose "Add Gcloud CLI to PATH" ✓ (important!)
# 4. Restart PowerShell/Command Prompt

# Verify installation:
gcloud --version
```

#### Don't have Chocolatey? Install it first:
```powershell
# Run as Administrator
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
```

### 2. Go Programming Language (Go 1.21+)
```powershell
# Download from https://golang.org/dl/
# Run installer and add to PATH
# Verify:
go version

# Should show: go version go1.21 or higher
```

### 3. Firebase CLI (Optional - for local testing)
```powershell
# Requires Node.js/npm
npm install -g firebase-tools

# Verify:
firebase --version
```

---

## Setup Instructions (Step by Step)

### Step 1: Install Google Cloud SDK (Windows)

1. **Download Installer**:
   - Go to: https://cloud.google.com/sdk/docs/install-cloud-sdk
   - Click "Windows 64-bit (with Python included)" installer

2. **Run Installer**:
   - Execute `GoogleCloudSDKInstaller.exe`
   - Accept default installation directory
   - **IMPORTANT**: Check "Add gcloud CLI to PATH" ✓

3. **Initialize gcloud**:
   ```powershell
   # After installation, restart PowerShell as Administrator
   gcloud init
   
   # Follow prompts:
   # 1. Select "Y" to log in
   # 2. Choose brower to authenticate
   # 3. Approve permissions
   # 4. Select project: locazar-f20b6
   ```

4. **Verify**:
   ```powershell
   gcloud --version
   gcloud config list
   ```

### Step 2: Authenticate with Google Cloud

```powershell
# Set current project
gcloud config set project locazar-f20b6

# Verify authentication
gcloud auth list

# You should see your email with a checkmark (✓ ACTIVE)
```

### Step 3: Verify Service Account Key

The credential file already exists:
```powershell
# Check if file exists
ls C:\Users\HP\Documents\Locazar\mandi-backend\locazar-f20b6-c125ff67e902.json

# Should display the file details
```

### Step 4: Enable Required APIs

```powershell
# Enable Cloud Functions API
gcloud services enable cloudfunctions.googleapis.com

# Enable Cloud Build API
gcloud services enable cloudbuild.googleapis.com

# Enable Firestore API
gcloud services enable firestore.googleapis.com

# Enable Eventarc API
gcloud services enable eventarc.googleapis.com

# Verify (should take 1-2 minutes)
gcloud services list --enabled | Select-String "functions"
```

---

##  🚀 Now Deploy!

### Quick Deploy (One Command)

Once all prerequisites are installed, run this:

```powershell
cd c:\Users\HP\Documents\Locazar\mandi-backend

# Deploy using direct gcloud command
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
  --set-env-vars="GCP_PROJECT_ID=locazar-f20b6,LOG_LEVEL=INFO"
```

**Wait 2-5 minutes for deployment** ⏳

### Monitor Deployment

```powershell
# View deployment progress
gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1 --follow

# View function details
gcloud functions describe enquiry-notification-handler --gen2 --region=asia-south1
```

---

## ✅ Verify Deployment Success

Once deployment completes, verify with:

```powershell
# Check function status
gcloud functions describe enquiry-notification-handler `
  --gen2 `
  --region=asia-south1 `
  --format="table(status, runtime, memoryMb)"

# Should show STATUS = ACTIVE
```

---

## 🧪 Test the Function

### Option 1: Update Firestore Document

1. Go to: https://console.firebase.google.com/u/0/project/locazar-f20b6
2. Click **Firestore Database** → **enquiries** collection
3. Select any enquiry and update the `status` field
4. Return to PowerShell and check logs:

```powershell
gcloud functions logs read enquiry-notification-handler `
  --gen2 `
  --region=asia-south1 `
  --limit=10
```

### Option 2: Update via gcloud CLI

```powershell
# List enquiries
gcloud firestore documents list --collection-id=enquiries --project=locazar-f20b6

# Update one (example)
gcloud firestore documents update enquiries/ENQ-123 `
  --update-data status=resolved `
  --project=locazar-f20b6

# Check logs
gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1
```

---

## 📱 Enable Mobile Notifications

For users to receive push notifications, register FCM tokens:

### In Your Mobile App (JavaScript/React Native/Flutter):

```javascript
// Get FCM token from device
const token = await messaging.getToken();

// Register in Firestore
firebase.firestore()
  .collection('users')
  .doc(userId)
  .collection('fcmTokens')
  .doc(token)
  .set({
    token: token,
    isActive: true,
    platform: 'android', // or 'ios'
    registeredAt: new Date()
  });
```

### Verify Tokens in Firestore:
```powershell
# List users with FCM tokens
gcloud firestore documents list `
  --collection-id=users `
  --project=locazar-f20b6

# Check if specific user has tokens
gcloud firestore documents list `
  --collection-id=fcmTokens `
  --parent-document="users/USER-001" `
  --project=locazar-f20b6
```

---

## 🐛 Troubleshooting

### Issue: "gcloud command not found"
```
Solution:
1. Restart PowerShell (restart the terminal)
2. If still not found, verify installation:
   - Go to: C:\Program Files\Google\Cloud SDK
   - If folder doesn't exist, reinstall from: https://cloud.google.com/sdk/docs/install
```

### Issue: "Authentication failed"
```
Solution:
Run: gcloud auth login
- This opens a browser
- Log in with your Google account
- Grant permissions
- Close browser and return to terminal
```

### Issue: "Project not found"
```
Solution:
1. Verify project ID: gcloud projects list
2. Set correct project: gcloud config set project locazar-f20b6
3. Check you have permissions in that project
```

### Issue: "Permission denied"
```
Solution:
The service account might need more permissions. Run:
gcloud projects get-iam-policy locazar-f20b6

If needed, grant additional roles:
gcloud projects add-iam-policy-binding locazar-f20b6 `
  --member="user:your-email@gmail.com" `
  --role="roles/editor"
```

### Issue: "APIs not enabled"
```
Solution:
gcloud services enable \
  cloudfunctions.googleapis.com \
  cloudbuild.googleapis.com \
  firestore.googleapis.com \
  eventarc.googleapis.com
```

---

## 📋 Deployment Checklist

- [ ] Google Cloud SDK installed (`gcloud --version` works)
- [ ] Authenticated with Google (`gcloud auth list` shows active account)
- [ ] Project set to locazar-f20b6 (`gcloud config get-value project`)
- [ ] Required APIs enabled (run enable commands above)
- [ ] Go 1.21+ installed (`go version`)
- [ ] Run deployment command
- [ ] Wait 2-5 minutes for deployment
- [ ] Verify function deployed (`gcloud functions describe enquiry-notification-handler --gen2 --region=asia-south1`)
- [ ] Update test enquiry in Firestore
- [ ] Check logs for success (`gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1`)

---

## 🎯 What Happens After Deployment

1. **Function Active**: Cloud Function `enquiry-notification-handler` is running
2. **Firestore Trigger**: Listens to all updates in `enquiries` collection
3. **Status Change Detection**: When `status`, `assignedTo`, or `priority` changes:
   - Cloud Function triggers automatically
   - Reads FCM tokens from Firestore
   - Sends push notifications to registered devices
4. **Mobile Users Get Notifications**: Users see push notification on their phones/tablets

---

## 🔗 Useful Links

- **Google Cloud Console**: https://console.cloud.google.com/functions?project=locazar-f20b6
- **Firebase Console**: https://console.firebase.google.com/u/0/project/locazar-f20b6
- **Cloud Functions Logs**: https://console.cloud.google.com/logs/query?project=locazar-f20b6
- **Firestore**: https://console.firebase.google.com/u/0/project/locazar-f20b6/firestore
- **SDK Install**: https://cloud.google.com/sdk/docs/install-cloud-sdk

---

## ⭐ Next Steps

1. **Install gcloud SDK** (if not already done)
2. **Run authentication** (`gcloud init`)
3. **Enable APIs** (run enable commands)
4. **Deploy function** (run deployment command)
5. **Test with Firestore update**
6. **Check mobile notifications**

---

**Need Help?**
- Check QUICK_DEPLOYMENT.md for quick reference
- Check DEPLOYMENT_GUIDE.md for detailed information
- Review logs with: `gcloud functions logs read enquiry-notification-handler --gen2 --region=asia-south1 --follow`
