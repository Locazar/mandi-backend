#!/bin/bash
# Deployment script for the Enquiry Auto-Reject Cloud Function (Gen 2, HTTP)
# + its hourly Cloud Scheduler trigger.
#
# This is a SEPARATE, new function — it does not touch enquiry-notification-handler.
#
# Usage:
#   ./deploy.sh [PROJECT_ID] [REGION]
#
# Env overrides:
#   ENQUIRY_AUTO_REJECT_HOURS   window in hours before a stale enquiry is rejected (default 24)
#   SCHEDULE                    cron for the trigger (default "0 * * * *" = hourly)
#   SKIP_SCHEDULER=1            deploy the function only; do NOT create the live hourly trigger
#
# After deploy, verify safely without mutating data:
#   curl -H "Authorization: Bearer $(gcloud auth print-identity-token)" "<FUNCTION_URL>?dryRun=true"

set -e

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
log_info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# ── Configuration ─────────────────────────────────────────────────────────────
PROJECT_ID="${1:-$(gcloud config get-value project 2>/dev/null)}"
# asia-south1 (Mumbai) — same region as enquiry-notification-handler. us-central2
# is blocked by policy on this project; do not use it.
REGION="${2:-asia-south1}"
FUNCTION_NAME="enquiry-autoreject"
ENTRY_POINT="AutoRejectStaleEnquiries"
RUNTIME="go123"
MEMORY="256MB"
TIMEOUT="300"
AUTO_REJECT_HOURS="${ENQUIRY_AUTO_REJECT_HOURS:-24}"
SCHEDULE="${SCHEDULE:-0 * * * *}"
INVOKER_SA_NAME="enquiry-autoreject-invoker"

if [ -z "$PROJECT_ID" ]; then
    log_error "PROJECT_ID is required (pass as arg 1 or set a default gcloud project)"
    exit 1
fi
if ! command -v gcloud &>/dev/null; then
    log_error "gcloud CLI is not installed"; exit 1
fi

log_info "Deploying $FUNCTION_NAME"
log_info "Project: $PROJECT_ID | Region: $REGION | Window: ${AUTO_REJECT_HOURS}h"

gcloud config set project "$PROJECT_ID" >/dev/null

PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format="value(projectNumber)")
INVOKER_SA="${INVOKER_SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"

# ── Enable APIs ───────────────────────────────────────────────────────────────
log_info "Enabling required APIs..."
gcloud services enable \
    cloudfunctions.googleapis.com \
    run.googleapis.com \
    cloudbuild.googleapis.com \
    cloudscheduler.googleapis.com || log_warn "Some APIs might already be enabled"

# ── Stage source (monorepo replace directive must resolve inside the upload) ───
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT=$(cd "$SCRIPT_DIR/../../.." && pwd)
STAGING=$(mktemp -d)
cleanup() { rm -rf "$STAGING"; }
trap cleanup EXIT
log_info "Staging in $STAGING"

for f in "$SCRIPT_DIR"/*; do
    filename=$(basename "$f")
    [[ "$filename" == "go.mod.local" || "$filename" == "go.mod" || "$filename" == "deploy.sh" ]] && continue
    cp -r "$f" "$STAGING/"
done

# go.mod.local → go.mod, with the replace target rewritten to the bundled copy.
sed 's|replace github.com/rohit221990/mandi-backend => \.\./\.\./\.\.|replace github.com/rohit221990/mandi-backend => ./_mandi|' \
    "$SCRIPT_DIR/go.mod.local" > "$STAGING/go.mod"

# Bundle only the monorepo packages this function imports (service/notification)
# and their intra-repo deps (domain, utils/firestore) — same slim set the
# notification handler bundles.
mkdir -p "$STAGING/_mandi/pkg/domain" \
         "$STAGING/_mandi/pkg/service/notification" \
         "$STAGING/_mandi/pkg/utils/firestore"
cp "$REPO_ROOT/go.mod"                              "$STAGING/_mandi/"
cp "$REPO_ROOT/pkg/domain/"*.go                     "$STAGING/_mandi/pkg/domain/"
cp "$REPO_ROOT/pkg/service/notification/"*.go       "$STAGING/_mandi/pkg/service/notification/"
cp "$REPO_ROOT/pkg/utils/firestore/"*.go            "$STAGING/_mandi/pkg/utils/firestore/"
# Drop *_test.go from the bundle so test-only deps aren't required at build.
find "$STAGING/_mandi" -name '*_test.go' -delete

# ── Deploy the function (HTTP, auth-required) ─────────────────────────────────
log_info "Deploying Cloud Function (this triggers a Cloud Build, ~2-4 min)..."
gcloud functions deploy "$FUNCTION_NAME" \
    --gen2 \
    --runtime "$RUNTIME" \
    --region "$REGION" \
    --source "$STAGING" \
    --entry-point "$ENTRY_POINT" \
    --trigger-http \
    --no-allow-unauthenticated \
    --memory "$MEMORY" \
    --timeout "$TIMEOUT" \
    --set-env-vars "ENQUIRY_AUTO_REJECT_HOURS=${AUTO_REJECT_HOURS}" || {
    log_error "Function deployment failed"; exit 1
}

FUNCTION_URL=$(gcloud functions describe "$FUNCTION_NAME" --region "$REGION" --gen2 \
    --format="value(serviceConfig.uri)")
log_info "Function deployed: $FUNCTION_URL"

echo ""
log_info "Dry-run test (lists what WOULD be rejected, mutates nothing):"
echo "  curl -H \"Authorization: Bearer \$(gcloud auth print-identity-token)\" \"${FUNCTION_URL}?dryRun=true\""
echo ""

# ── Scheduler (go-live). Skip with SKIP_SCHEDULER=1 to test first. ────────────
if [ "$SKIP_SCHEDULER" = "1" ]; then
    log_warn "SKIP_SCHEDULER=1 set — function deployed but NOT yet scheduled."
    log_warn "It will not run automatically until you create the Scheduler trigger."
    exit 0
fi

log_info "Ensuring invoker service account: $INVOKER_SA"
gcloud iam service-accounts describe "$INVOKER_SA" >/dev/null 2>&1 || \
    gcloud iam service-accounts create "$INVOKER_SA_NAME" \
        --display-name="Cloud Scheduler invoker for $FUNCTION_NAME"

log_info "Granting the invoker permission to call the function..."
gcloud functions add-invoker-policy-binding "$FUNCTION_NAME" \
    --region "$REGION" \
    --member="serviceAccount:${INVOKER_SA}" >/dev/null || \
    log_warn "Could not add invoker binding (may already exist)"

log_info "Creating/updating the hourly Scheduler trigger..."
if gcloud scheduler jobs describe "${FUNCTION_NAME}-trigger" --location "$REGION" >/dev/null 2>&1; then
    gcloud scheduler jobs update http "${FUNCTION_NAME}-trigger" \
        --location "$REGION" \
        --schedule "$SCHEDULE" \
        --uri "$FUNCTION_URL" \
        --http-method POST \
        --oidc-service-account-email "$INVOKER_SA" \
        --oidc-token-audience "$FUNCTION_URL"
else
    gcloud scheduler jobs create http "${FUNCTION_NAME}-trigger" \
        --location "$REGION" \
        --schedule "$SCHEDULE" \
        --uri "$FUNCTION_URL" \
        --http-method POST \
        --oidc-service-account-email "$INVOKER_SA" \
        --oidc-token-audience "$FUNCTION_URL"
fi

echo ""
log_info "Done. The function runs on schedule: '$SCHEDULE' (default hourly)."
log_info "Trigger a run now:   gcloud scheduler jobs run ${FUNCTION_NAME}-trigger --location $REGION"
log_info "Live logs:           gcloud functions logs read $FUNCTION_NAME --region $REGION --gen2 --limit 50"
