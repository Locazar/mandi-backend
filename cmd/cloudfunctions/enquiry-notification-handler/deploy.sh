#!/bin/bash
# Deployment script for Enquiry Notification Cloud Function (Gen 2)
#
# Usage: ./deploy.sh [PROJECT_ID] [REGION]
# Example: ./deploy.sh my-project asia-south1 asia-south2

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ID="${1:-}"
# Cloud Functions Gen 2 region; asia-south1 (Mumbai) is the nearest supported region to our Firestore location
REGION="${2:-asia-south1}"
# Must match your Firestore database location exactly (our DB is in asia-south2 / Delhi)
FIRESTORE_LOCATION="${3:-asia-south2}"
FUNCTION_NAME="enquiry-notification-handler"
RUNTIME="go123"
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
    echo "Example: $0 my-project asia-south1"
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

# Resolve project number (needed for service agent references)
PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format="value(projectNumber)") || {
    log_error "Failed to resolve project number"
    exit 1
}
log_info "Project number: $PROJECT_NUMBER"

# Enable required APIs
log_info "Enabling required APIs..."
gcloud services enable \
    cloudfunctions.googleapis.com \
    cloudbuild.googleapis.com \
    cloudrun.googleapis.com \
    eventarc.googleapis.com \
    firebase.googleapis.com || log_warn "Some APIs might already be enabled"

# Grant IAM roles required for Eventarc-triggered Gen 2 Cloud Functions
log_info "Granting required IAM roles..."

# The function's service account must be able to receive Eventarc events
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:$SERVICE_ACCOUNT" \
    --role="roles/eventarc.eventReceiver" \
    --condition=None \
    && log_info "Granted roles/eventarc.eventReceiver" \
    || log_warn "Could not grant eventarc.eventReceiver (may already exist)"

# Gen 2 functions run on Cloud Run; the SA needs permission to be invoked
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:$SERVICE_ACCOUNT" \
    --role="roles/run.invoker" \
    --condition=None \
    && log_info "Granted roles/run.invoker" \
    || log_warn "Could not grant run.invoker (may already exist)"

# Eventarc Firestore triggers use Pub/Sub internally; the Pub/Sub service agent
# must be able to create tokens for the function's service account
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:service-${PROJECT_NUMBER}@gcp-sa-pubsub.iam.gserviceaccount.com" \
    --role="roles/iam.serviceAccountTokenCreator" \
    --condition=None \
    && log_info "Granted roles/iam.serviceAccountTokenCreator to Pub/Sub SA" \
    || log_warn "Could not grant iam.serviceAccountTokenCreator to Pub/Sub SA (may already exist)"

# Deploy the function
log_info "Deploying Cloud Function..."

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Cloud Build only receives the --source directory; a local replace directive
# like "../../.." cannot resolve in the remote build environment.
# We stage the function together with the required monorepo packages so that
# the replace target (_mandi/) is contained within the uploaded directory.
STAGING=$(mktemp -d)
cleanup() { rm -rf "$STAGING"; }
trap cleanup EXIT

log_info "Creating deployment staging directory: $STAGING"

# Copy all function source files (go.mod.local is rewritten below).
for f in "$SCRIPT_DIR"/*; do
    filename=$(basename "$f")
    [[ "$filename" == "go.mod.local" || "$filename" == "go.mod" ]] && continue
    cp -r "$f" "$STAGING/"
done

# Rewrite the replace directive so it resolves inside the staging dir.
# Original:  replace github.com/rohit221990/mandi-backend => ../../..
# New:       replace github.com/rohit221990/mandi-backend => ./_mandi
sed 's|replace github.com/rohit221990/mandi-backend => \.\./\.\./\.\.|replace github.com/rohit221990/mandi-backend => ./_mandi|' \
    "$SCRIPT_DIR/go.mod.local" > "$STAGING/go.mod"

# Bundle the monorepo packages imported by this function.
REPO_ROOT=$(cd "$SCRIPT_DIR/../../.." && pwd)
log_info "Bundling monorepo packages from: $REPO_ROOT"

mkdir -p "$STAGING/_mandi/pkg/domain"
mkdir -p "$STAGING/_mandi/pkg/service/notification"
mkdir -p "$STAGING/_mandi/pkg/utils/firestore"

# The replace target needs a go.mod so Go can confirm the module identity.
cp "$REPO_ROOT/go.mod" "$STAGING/_mandi/"

# Copy the specific packages imported by this function and their transitive
# intra-repo dependencies (notification → domain, firestore/utils → domain).
cp "$REPO_ROOT/pkg/domain/"*.go         "$STAGING/_mandi/pkg/domain/"
cp "$REPO_ROOT/pkg/service/notification/"*.go "$STAGING/_mandi/pkg/service/notification/"
cp "$REPO_ROOT/pkg/utils/firestore/"*.go     "$STAGING/_mandi/pkg/utils/firestore/"

gcloud functions deploy "$FUNCTION_NAME" \
    --runtime "$RUNTIME" \
    --gen2 \
    --region "$REGION" \
    --source "$STAGING" \
    --entry-point ProcessEnquiryUpdate \
    --memory "$MEMORY" \
    --timeout "$TIMEOUT" \
    --service-account "$SERVICE_ACCOUNT" \
    --set-env-vars "LOG_LEVEL=INFO,ENABLE_IDEMPOTENCY_CHECK=true" \
    --ingress-settings internal-only \
    --trigger-event-filters "type=google.cloud.firestore.document.v1.updated" \
    --trigger-event-filters "database=(default)" \
    --trigger-event-filters-path-pattern "document=enquiry/*" \
    --trigger-location "$FIRESTORE_LOCATION" || {
    log_error "Function deployment failed"
    exit 1
}

log_info "Cloud Function (update handler) deployed successfully!"

# Deploy the create handler as a separate Cloud Function entry point
CREATE_FUNCTION_NAME="${FUNCTION_NAME}-create"

log_info "Deploying Cloud Function (create handler): $CREATE_FUNCTION_NAME"

gcloud functions deploy "$CREATE_FUNCTION_NAME" \
    --runtime "$RUNTIME" \
    --gen2 \
    --region "$REGION" \
    --source "$STAGING" \
    --entry-point ProcessEnquiryCreate \
    --memory "$MEMORY" \
    --timeout "$TIMEOUT" \
    --service-account "$SERVICE_ACCOUNT" \
    --set-env-vars "LOG_LEVEL=INFO,ENABLE_IDEMPOTENCY_CHECK=true" \
    --ingress-settings internal-only \
    --trigger-event-filters "type=google.cloud.firestore.document.v1.created" \
    --trigger-event-filters "database=(default)" \
    --trigger-event-filters-path-pattern "document=enquiry/*" \
    --trigger-location "$FIRESTORE_LOCATION" || {
    log_error "Create handler function deployment failed"
    exit 1
}

log_info "Cloud Function (create handler) deployed successfully!"
log_info "Eventarc triggers were configured inline with each function deployment above."

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
