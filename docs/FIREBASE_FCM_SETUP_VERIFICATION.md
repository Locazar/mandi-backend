# Firebase FCM 404 Error - Complete Debugging Guide

## Current Status
**Error**: `404 Not Found` on Firebase `/batch` endpoint
**Project**: `locazar-f20b6`  
**Service Account**: `firebase-adminsdk-fbsvc@locazar-f20b6.iam.gserviceaccount.com`

---

## Root Cause Analysis

The 404 error on `/batch` endpoint indicates **one of these issues**:

1. ❌ **FCM API Not Enabled** (MOST COMMON)
   - Firebase Cloud Messaging API is disabled in GCP project
   - Result: Requests to `/batch` endpoint fail with 404

2. ❌ **Service Account Lacks Permissions**
   - Missing "Firebase Cloud Messaging Admin" role
   - Service account can't access FCM API
   
3. ❌ **Invalid/Expired Credentials**
   - Private key in JSON is corrupted or expired
   - Authentication fails silently

4. ❌ **API Quota Exceeded**
   - Project has hit daily/monthly limits
   - Firebase Cloud Messaging quota exhausted

5. ❌ **Billing Issue**
   - Project billing account not valid
   - API access restricted for unpaid projects

---

## Step-by-Step Verification

### ✅ Step 1: Verify Local Credentials File

```powershell
# PowerShell
$cred = Get-Content "C:\Users\HP\Documents\Locazar\mandi-backend\locazar-f20b6-f024a4597849.json" | ConvertFrom-Json

Write-Host "Project ID: $($cred.project_id)"
Write-Host "Service Account: $($cred.client_email)"
Write-Host "Key ID: $($cred.private_key_id)"
Write-Host "Expires: $($cred.client_x509_cert_url)" # Note: this URL won't change
```

**Expected Output:**
```
Project ID: locazar-f20b6
Service Account: firebase-adminsdk-fbsvc@locazar-f20b6.iam.gserviceaccount.com
Key ID: f024a4597849d0a5f362e8732c98d208b2214a65
```

---

### ✅ Step 2: Verify FCM API is Enabled

**In Google Cloud Console:**

1. Go to: https://console.cloud.google.com/
2. Select Project: **`locazar-f20b6`**
3. Navigation → **APIs & Services** → **Enabled APIs & Services**
4. Search for: **`Cloud Messaging`** or **`Firebase Cloud Messaging API`**

**Expected Result:**
- ✅ Should show **"Firebase Cloud Messaging"** or **"Firebase Cloud Messaging API"** with status **"ENABLED"**

**If NOT found:**
1. Click **"+ Enable APIs and Services"** at top
2. Search: "Cloud Messaging" or "Firebase Cloud Messaging"
3. Click the result
4. Click **"Enable"** button
5. Wait 2-3 minutes for activation

---

### ✅ Step 3: Verify Service Account Permissions

**In Google Cloud Console:**

1. Go to: https://console.cloud.google.com/
2. Select Project: **`locazar-f20b6`**
3. Navigation → **IAM & Admin** → **IAM**
4. Look for service account: **`firebase-adminsdk-fbsvc@locazar-f20b6.iam.gserviceaccount.com`**

**Required Roles (one of these):**
- ✅ `Editor` (has all permissions - easiest for testing)
- ✅ `Firebase Cloud Messaging Admin`
- ✅ `Service Accounts Token Creator`

**If permissions are missing:**
1. Click **"Grant Access"** button
2. Search for service account email: `firebase-adminsdk-fbsvc@locazar-f20b6.iam.gserviceaccount.com`
3. Assign role: **`Editor`** or **`Firebase Cloud Messaging Admin`**
4. Click **"Save"**
5. Wait 1-2 minutes for permissions to propagate

---

### ✅ Step 4: Verify in Firebase Console

**In Firebase Console:**

1. Go to: https://console.firebase.google.com/
2. Select Project: **`locazar-f20b6`**
3. Left menu → **Project settings** (gear icon)
4. Tab → **Service Accounts**
5. Click your service account link or **"Go to Cloud Console"**

**Verify:**
- Service account is still active in GCP
- No errors or warnings displayed
- Private key exists and is valid

---

### ✅ Step 5: Check Firebase Cloud Messaging Settings

**In Firebase Console:**

1. Go to: https://console.firebase.google.com/project/locazar-f20b6
2. Left menu → **Cloud Messaging**
3. Verify tab is accessible (no permission denied errors)

**Check:**
- ✅ Server API Key exists
- ✅ No quota warnings
- ✅ No billing warnings

---

## Testing the Fix

### Test 1: Verify Environment Variable

```powershell
# PowerShell
$env:FIREBASE_CREDENTIALS_FILE = "C:\Users\HP\Documents\Locazar\mandi-backend\locazar-f20b6-f024a4597849.json"
echo $env:FIREBASE_CREDENTIALS_FILE

# Test it's persisted
[Environment]::GetEnvironmentVariable("FIREBASE_CREDENTIALS_FILE", [EnvironmentVariableTarget]::User)
```

### Test 2: Check Application Logs

When you run `make run`, watch for logs starting with `[FCM]`:

**Success Logs:**
```
[FCM] Retrieved 2 FCM tokens for shop 1
[FCM] Token[0] added (length: 152)
[FCM] Token[1] added (length: 152)
[FCM] Calling SendMulticast with 2 tokens...
[FCM SUCCESS] Notification sent - Successful: 2, Failed: 0
```

**Failure Logs (Root Cause):**
```
[FCM ERROR] SendMulticast failed: unexpected http response with status: 404
[FCM DIAGNOSTIC] Troubleshooting steps:
  1. Verify FCM API is enabled in GCP project: locazar-f20b6
  2. Check service account has 'Firebase Cloud Messaging Admin' role
  3. Inspect GCP project quota and billing status
  4. Validate credentials file is not corrupted
  5. Ensure tokens are valid FCM device tokens from Firebase
```

### Test 3: Validate FCM Tokens

**The tokens being sent must be:**
- ✅ Generated by Firebase from actual Android/iOS/Web clients
- ✅ NOT expired (tokens expire after ~60 days of inactivity)
- ✅ From the SAME Firebase project (`locazar-f20b6`)

**Check token validity:**
1. Verify tokens are ~152 characters long
2. Each token should start with letters/numbers (not special chars except `:`)
3. Ensure app is actually registering tokens with Firebase

---

## Common Mistakes

### ❌ Mistake 1: Wrong Project ID
```
Credentials Project: locazar-f20b6
But sending to: locazar-f20b6-prod
Result: 404 error
```
**Solution**: Ensure credentials match the Firebase project you're using

### ❌ Mistake 2: API Not Enabled
```
Error: 404 /batch not found
Cause: Firebase Cloud Messaging API is disabled
```
**Solution**: Enable FCM API in Google Cloud Console (see Step 2 above)

### ❌ Mistake 3: Service Account Deleted
```
Error: Authentication failed
Cause: Service account was deleted or disabled
```
**Solution**: Create new service account key in Firebase Console → Project Settings → Service Accounts

### ❌ Mistake 4: Expired Credentials
```
Error: Invalid authentication token
Cause: Private key expired (after ~2-3 years)
```
**Solution**: Regenerate key in Google Cloud Console → Service Accounts

### ❌ Mistake 5: Invalid FCM Tokens
```
Error: 400 Bad Request or silent failures
Cause: Tokens don't match Firebase project or are expired
```
**Solution**: Verify app is registering tokens with Firebase SDK

---

## Credentials Regeneration (If Needed)

### Create New Service Account Key

1. Go to: https://console.cloud.google.com/iam-admin/serviceaccounts
2. Select Project: `locazar-f20b6`
3. Find: `firebase-adminsdk-fbsvc@locazar-f20b6.iam.gserviceaccount.com`
4. Click on it
5. Tab → **Keys**
6. Click **Add Key** → **Create new key**
7. Choose **JSON** format
8. Click **Create**
9. Replace your `locazar-f20b6-f024a4597849.json` file

---

## Post-Fix Checklist

Before retesting after making changes:

- [ ] FCM API enabled in Google Cloud Console
- [ ] Service account has `Editor` or `Firebase Cloud Messaging Admin` role
- [ ] `FIREBASE_CREDENTIALS_FILE` environment variable is set to absolute path
- [ ] Credentials file exists and is not corrupted
- [ ] At least one valid FCM token in database for test shop
- [ ] Application restarted after environment variable change
- [ ] Check application logs for `[FCM]` prefix messages

---

## Detailed Logs to Collect

When troubleshooting, enable and collect these logs:

1. **Application startup logs** (Firebase initialization)
   ```
   [FCM] Initializing Firebase app with credentials file: ...
   [FCM] Credentials validated - Project: locazar-f20b6, ...
   [FCM] Firebase app initialized successfully
   [FCM] Firebase messaging client initialized and ready for use
   ```

2. **Notification send logs** (actual message sending)
   ```
   [FCM] Retrieved N FCM tokens for shop ...
   [FCM] Token[0] added (length: 152)
   [FCM] Calling SendMulticast with N tokens...
   ```

3. **Error logs** (if something fails)
   ```
   [FCM ERROR] SendMulticast failed: ...
   [FCM DIAGNOSTIC] Troubleshooting steps:
   ```

---

## Still Not Working?

If you've verified all steps above and still getting 404:

### Option 1: Check GCP Quotas
1. Google Cloud Console → Project `locazar-f20b6`
2. Navigation → **APIs & Services** → **Quotas**
3. Search: "Cloud Messaging"
4. Check if any quotas are exceeded

### Option 2: Check Billing
1. Google Cloud Console → Project `locazar-f20b6`
2. Navigation → **Billing** (Linked Billing Account)
3. Verify account is active and payment method is valid

### Option 3: Regenerate Credentials
- Delete old service account key
- Create new JSON key (see section above)
- Replace credentials file
- Restart application

### Option 4: Check Firewall/Network
```powershell
# Test connectivity to Firebase endpoint
Test-NetConnection -ComputerName fcm.googleapis.com -Port 443

# For Linux/WSL
curl -I https://fcm.googleapis.com/
```

---

## Reference Documentation

- [Firebase Admin SDK (Go)](https://firebase.google.com/docs/admin/setup)
- [Firebase Cloud Messaging](https://firebase.google.com/docs/cloud-messaging)
- [FCM HTTP Protocol](https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages/send)
- [GCP IAM Roles](https://cloud.google.com/iam/docs/understanding-roles)
- [FCM Troubleshooting](https://firebase.google.com/docs/cloud-messaging/troubleshooting)
