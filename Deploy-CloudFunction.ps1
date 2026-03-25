# Automated Deployment Script for Enquiry Notification Cloud Function - PowerShell Version
# This script deploys the cloud function to GCP Cloud Functions (Gen 2)

param(
    [string]$ProjectID = "locazar-f20b6",
    [string]$Region = "asia-south1",
    [string]$FunctionName = "enquiry-notification-handler"
)

# Configuration
$Runtime = "go121"
$Memory = "512MB"
$Timeout = "60"
$SourceDir = "cmd/cloudfunctions/enquiry-notification-handler"

# Colors
$Green = "`e[32m"
$Yellow = "`e[33m"
$Blue = "`e[34m"
$Red = "`e[31m"
$Reset = "`e[0m"

function Write-Step {
    param([string]$Message)
    Write-Host "${Blue}→${Reset} $Message" -ForegroundColor Cyan
}

function Write-Info {
    param([string]$Message)
    Write-Host "${Green}✓${Reset} $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "${Yellow}⚠${Reset} $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "${Red}✗${Reset} $Message" -ForegroundColor Red
}

# Header
Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "GCP Cloud Function Deployment - Enquiry Notifications" -ForegroundColor Cyan
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

# Step 1: Check prerequisites
Write-Step "Checking prerequisites..."

$gcloudCmd = Get-Command gcloud -ErrorAction SilentlyContinue
if (-not $gcloudCmd) {
    Write-Error "gcloud CLI is not installed"
    Write-Host "Install from: https://cloud.google.com/sdk/docs/install"
    exit 1
}
Write-Info "gcloud CLI found"

$goCmd = Get-Command go -ErrorAction SilentlyContinue
if (-not $goCmd) {
    Write-Error "Go is not installed"
    Write-Host "Install from: https://golang.org/dl/ (Go 1.21+)"
    exit 1
}
$goVersion = (go version) -split ' ' | Select-Object -Index 2
Write-Info "Go $goVersion found"

# Step 2: Set GCP project
Write-Step "Setting GCP project..."
gcloud config set project $ProjectID
Write-Info "Project set to: $ProjectID"

# Step 3: Verify service account
Write-Step "Verifying service account..."
$ServiceAccount = "firebase-adminsdk@${ProjectID}.iam.gserviceaccount.com"

try {
    gcloud iam service-accounts describe $ServiceAccount | Out-Null
    Write-Info "Service account verified: $ServiceAccount"
} catch {
    Write-Warning "Service account not found, creating one..."
    gcloud iam service-accounts create firebase-adminsdk --display-name="Firebase Admin SDK Service Account"
    
    Write-Step "Granting IAM roles..."
    gcloud projects add-iam-policy-binding $ProjectID `
        --member="serviceAccount:$ServiceAccount" `
        --role="roles/firebase.admin" | Out-Null
    
    gcloud projects add-iam-policy-binding $ProjectID `
        --member="serviceAccount:$ServiceAccount" `
        --role="roles/firestore.serviceAgent" | Out-Null
    
    Write-Info "Service account created and roles assigned"
}

# Step 4: Enable APIs
Write-Step "Enabling required Google Cloud APIs..."
@(
    "cloudfunctions.googleapis.com",
    "cloudbuild.googleapis.com",
    "firestore.googleapis.com",
    "eventarc.googleapis.com",
    "logging.googleapis.com",
    "pubsub.googleapis.com"
) | ForEach-Object {
    gcloud services enable $_ 2>&1 | Out-Null
}
Write-Info "APIs enabled"

# Step 5: Deploy function
Write-Step "Deploying Cloud Function to $Region..."
Write-Host ""

gcloud functions deploy $FunctionName `
    --gen2 `
    --runtime=$Runtime `
    --region=$Region `
    --source=$SourceDir `
    --entry-point=ProcessEnquiryUpdate `
    --service-account=$ServiceAccount `
    --memory=$Memory `
    --timeout="${Timeout}s" `
    --set-env-vars="GCP_PROJECT_ID=$ProjectID,LOG_LEVEL=INFO,MONITORED_FIELDS=status,assignedTo,priority" `
    --trigger-event-filters="type=google.cloud.firestore.document.v1.updated" `
    --trigger-event-filters="database=(default)" `
    --trigger-event-filters-path-pattern="document=enquiries/{enquiryId}"

Write-Host ""

# Step 6: Verify deployment
Write-Step "Verifying deployment..."
$deploymentStatus = gcloud functions describe $FunctionName --gen2 --region=$Region 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Info "Function deployed successfully!"
} else {
    Write-Error "Function deployment verification failed"
    exit 1
}

# Step 7: Display details
Write-Host ""
Write-Step "Deployment Details:"
gcloud functions describe $FunctionName `
    --gen2 `
    --region=$Region `
    --format="table(name.scope(functions):label=FUNCTION, runtime:label=RUNTIME, status:label=STATUS, updateTime:label=DEPLOYED)"

# Step 8: Show logs
Write-Host ""
Write-Step "Recent Function Logs:"
Write-Host ""
gcloud functions logs read $FunctionName --gen2 --region=$Region --limit=5

# Step 9: Success
Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host "✓ Deployment Complete!" -ForegroundColor Green
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host ""

Write-Info "Function Name: $FunctionName"
Write-Info "Region: $Region"
Write-Info "Project: $ProjectID"
Write-Info "Status: DEPLOYED"

Write-Host ""
Write-Step "Next Steps:"
Write-Host "1. Update an enquiry document in Firestore Console"
Write-Host "2. Monitor logs with:"
Write-Host "   gcloud functions logs read $FunctionName --gen2 --region=$Region --follow"
Write-Host ""

Write-Step "View in Console:"
Write-Host "   https://console.cloud.google.com/functions/details/$Region/$FunctionName?gen2=true&project=$ProjectID"
Write-Host ""

Write-Step "Firebase Console:"
Write-Host "   https://console.firebase.google.com/u/0/project/$ProjectID"
Write-Host ""
