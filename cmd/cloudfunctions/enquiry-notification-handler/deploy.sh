#!/bin/bash
# Deployment script for Enquiry Notification Cloud Function (Gen 2)
#
# Usage: ./deploy.sh [PROJECT_ID] [REGION]
# Example: ./deploy.sh my-project us-central1

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ID="${1:-}"
REGION="${2:-us-central1}"
FUNCTION_NAME="enquiry-notification-handler"
RUNTIME="go121"
MEMORY="512MB"
TIMEOUT="60"

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Validate inputs
if [ -z "$PROJECT_ID" ]; then
    log_error "PROJECT_ID is required"
    echo "Usage: $0 PROJECT_ID [REGION]"
    echo "Example: $0 my-project us-central1"
    exit 1
fi

log_info "Deploying Enquiry Notification Handler"
log_info "Project: $PROJECT_ID"
log_info "Region: $REGION"
log_info "Function: $FUNCTION_NAME"

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    log_error "gcloud CLI is not installed"
    exit 1
fi

# Authenticate
log_info "Setting up gcloud authentication..."
gcloud auth login || {
    log_error "Authentication failed"
    exit 1
}

# Set project
log_info "Setting project to $PROJECT_ID..."
gcloud config set project "$PROJECT_ID" || {
    log_error "Failed to set project"
    exit 1
}

# Get service account email
SERVICE_ACCOUNT=$(gcloud iam service-accounts list \
    --filter="displayName:Firebase" \
    --format="value(email)" | head -1)

if [ -z "$SERVICE_ACCOUNT" ]; then
    log_warn "No Firebase service account found, using default"
    SERVICE_ACCOUNT="${PROJECT_ID}@appspot.gserviceaccount.com"
fi

log_info "Using service account: $SERVICE_ACCOUNT"

# Enable required APIs
log_info "Enabling required APIs..."
gcloud services enable \
    cloudfunctions.googleapis.com \
    cloudbuild.googleapis.com \
    cloudrun.googleapis.com \
    eventarc.googleapis.com \
    firebase.googleapis.com || log_warn "Some APIs might already be enabled"

# Deploy the function
log_info "Deploying Cloud Function..."

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

gcloud functions deploy "$FUNCTION_NAME" \
    --runtime "$RUNTIME" \
    --gen2 \
    --region "$REGION" \
    --source "$SCRIPT_DIR" \
    --entry-point ProcessEnquiryUpdate \
    --memory "$MEMORY" \
    --timeout "$TIMEOUT" \
    --service-account "$SERVICE_ACCOUNT" \
    --set-env-vars "LOG_LEVEL=INFO,ENABLE_IDEMPOTENCY_CHECK=true" \
    --ingress-settings internal-only \
    --allow-unauthenticated || {
    log_error "Function deployment failed"
    exit 1
}

log_info "Cloud Function deployed successfully!"

# Create Eventarc trigger
log_info "Setting up Eventarc trigger..."

TRIGGER_NAME="${FUNCTION_NAME}-trigger"

# Check if trigger already exists
if gcloud eventarc triggers describe "$TRIGGER_NAME" --location "$REGION" &>/dev/null; then
    log_warn "Trigger already exists, skipping creation"
else
    gcloud eventarc triggers create "$TRIGGER_NAME" \
        --location "$REGION" \
        --destination-cloud-function "$FUNCTION_NAME" \
        --destination-cloud-function-region "$REGION" \
        --event-filters "type=google.cloud.firestore.document.v1.updated" \
        --event-filters "database=(default)" \
        --event-filters "document=enquiries/*" \
        --service-account "$SERVICE_ACCOUNT" || log_warn "Trigger creation might have failed, verify in console"
fi

# Verify deployment
log_info "Verifying deployment..."
if gcloud functions describe "$FUNCTION_NAME" --region "$REGION" &>/dev/null; then
    log_info "Function found!"
    
    # Display function details
    echo ""
    log_info "Function Details:"
    gcloud functions describe "$FUNCTION_NAME" --region "$REGION" --gen2 --format=json | jq '{
        name: .displayName,
        status: .status,
        runtime: .runtime,
        memory: .availableMemoryMb,
        timeout: .timeout,
        serviceAccount: .serviceConfig.serviceAccountEmail,
        region: .location
    }' || gcloud functions describe "$FUNCTION_NAME" --region "$REGION" --gen2
else
    log_error "Function verification failed"
    exit 1
fi

# Function logs URL
LOGS_URL="https://console.cloud.google.com/functions/details/$REGION/$FUNCTION_NAME?project=$PROJECT_ID"
echo ""
log_info "View function in Cloud Console:"
echo "$LOGS_URL"

# Logs command
echo ""
log_info "To view live logs, run:"
echo "gcloud functions logs read $FUNCTION_NAME --region $REGION --follow"

echo ""
log_info "Deployment completed successfully!"
echo ""
