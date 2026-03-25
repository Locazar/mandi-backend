#!/bin/bash
# Automated Deployment Script for Enquiry Notification Cloud Function
# This script deploys the cloud function to GCP Cloud Functions (Gen 2)

set -e

# Configuration
PROJECT_ID="locazar-f20b6"
REGION="asia-south1"
FUNCTION_NAME="enquiry-notification-handler"
RUNTIME="go121"
MEMORY="512MB"
TIMEOUT="60"
SOURCE_DIR="cmd/cloudfunctions/enquiry-notification-handler"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}✓${NC} $1"
}

log_step() {
    echo -e "${BLUE}→${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

# Print header
echo -e "\n${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}GCP Cloud Function Deployment - Enquiry Notifications${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"

# Step 1: Verify prerequisites
log_step "Checking prerequisites..."
if ! command -v gcloud &> /dev/null; then
    log_error "gcloud CLI is not installed"
    echo "Install from: https://cloud.google.com/sdk/docs/install"
    exit 1
fi
log_info "gcloud CLI found"

if ! command -v go &> /dev/null; then
    log_error "Go is not installed"
    echo "Install from: https://golang.org/dl/ (Go 1.21+)"
    exit 1
fi
go_version=$(go version | awk '{print $3}')
log_info "Go $go_version found"

# Step 2: Verify GCP project
log_step "Verifying GCP project..."
current_project=$(gcloud config get-value project 2>/dev/null || echo "")

if [ "$current_project" != "$PROJECT_ID" ]; then
    log_step "Setting GCP project to $PROJECT_ID..."
    gcloud config set project $PROJECT_ID
fi
log_info "Project set to: $PROJECT_ID"

# Step 3: Verify service account exists
log_step "Verifying service account permissions..."
service_account="firebase-adminsdk@${PROJECT_ID}.iam.gserviceaccount.com"

if ! gcloud iam service-accounts describe "$service_account" &> /dev/null; then
    log_warn "Service account not found, creating one..."
    gcloud iam service-accounts create firebase-adminsdk \
        --display-name="Firebase Admin SDK Service Account"
    
    # Grant necessary roles
    gcloud projects add-iam-policy-binding $PROJECT_ID \
        --member="serviceAccount:$service_account" \
        --role="roles/firebase.admin"
    
    gcloud projects add-iam-policy-binding $PROJECT_ID \
        --member="serviceAccount:$service_account" \
        --role="roles/firestore.serviceAgent"
    
    gcloud projects add-iam-policy-binding $PROJECT_ID \
        --member="serviceAccount:$service_account" \
        --role="roles/cloudfunctions.developer"
else
    log_info "Service account verified: $service_account"
fi

# Step 4: Enable required APIs
log_step "Enabling required Google Cloud APIs..."
gcloud services enable cloudfunctions.googleapis.com
gcloud services enable cloudbuild.googleapis.com
gcloud services enable firestore.googleapis.com
gcloud services enable eventarc.googleapis.com
gcloud services enable logging.googleapis.com
gcloud services enable pubsub.googleapis.com
log_info "APIs enabled"

# Step 5: Deploy the function
log_step "Deploying Cloud Function..."
echo ""

gcloud functions deploy $FUNCTION_NAME \
    --gen2 \
    --runtime=$RUNTIME \
    --region=$REGION \
    --source=$SOURCE_DIR \
    --entry-point=ProcessEnquiryUpdate \
    --service-account=$service_account \
    --memory=$MEMORY \
    --timeout=${TIMEOUT}s \
    --set-env-vars="GCP_PROJECT_ID=$PROJECT_ID,LOG_LEVEL=INFO,MONITORED_FIELDS=status,assignedTo,priority" \
    --trigger-event-filters="type=google.cloud.firestore.document.v1.updated" \
    --trigger-event-filters="database=(default)" \
    --trigger-event-filters-path-pattern="document=enquiries/{enquiryId}"

# Step 6: Verify deployment
log_step "Verifying deployment..."
if gcloud functions describe $FUNCTION_NAME --gen2 --region=$REGION &> /dev/null; then
    log_info "Function deployed successfully!"
else
    log_error "Function deployment verification failed"
    exit 1
fi

# Step 7: Display deployment details
echo ""
log_step "Deployment Details:"
gcloud functions describe $FUNCTION_NAME \
    --gen2 \
    --region=$REGION \
    --format="table(
        name.scope(functions):label=FUNCTION,
        runtime:label=RUNTIME,
        status:label=STATUS,
        updateTime:label=DEPLOYED
    )"

# Step 8: Get function logs
log_step "Getting recent logs..."
echo ""
gcloud functions logs read $FUNCTION_NAME --gen2 --region=$REGION --limit=5

# Step 9: Success message
echo ""
echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✓ Deployment Complete!${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}\n"

log_info "Function Name: $FUNCTION_NAME"
log_info "Region: $REGION"
log_info "Project: $PROJECT_ID"
log_info "Status: DEPLOYED"

echo ""
log_step "Next Steps:"
echo "1. Update an enquiry document in Firestore Console"
echo "2. Monitor logs with:"
echo "   gcloud functions logs read $FUNCTION_NAME --gen2 --region=$REGION --follow"
echo ""
log_step "View in Console:"
echo "   https://console.cloud.google.com/functions/details/$REGION/$FUNCTION_NAME?gen2=true&project=$PROJECT_ID"
echo ""
log_step "Firebase Console:"
echo "   https://console.firebase.google.com/u/0/project/$PROJECT_ID"
echo ""
