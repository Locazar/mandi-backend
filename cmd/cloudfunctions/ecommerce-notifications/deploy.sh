#!/usr/bin/env bash
# ---------------------------------------------------------------------------
# deploy-ecommerce-notifications.sh
#
# Deploys three Cloud Function instances (Gen 2) via Eventarc:
#   1. order-notification    – watches the "orders" collection
#   2. product-notification  – watches the "products" collection
#   3. shop-notification     – watches the "shops" collection
#
# Prerequisites:
#   gcloud CLI authenticated with a service account that has:
#     - roles/cloudfunctions.developer
#     - roles/eventarc.admin
#     - roles/firebase.admin (or roles/datastore.user + messaging.sender)
#
# Usage:
#   export PROJECT_ID=your-gcp-project-id
#   export REGION=asia-south1           # or any Cloud Functions region
#   bash deploy-ecommerce-notifications.sh
# ---------------------------------------------------------------------------

set -euo pipefail

PROJECT_ID="${PROJECT_ID:?Set PROJECT_ID}"
REGION="${REGION:-asia-south1}"
RUNTIME="go123"
FUNCTION_DIR="$(dirname "$0")"
ENTRY_POINT="ProcessEcommerceUpdate"

deploy_function() {
  local NAME=$1
  local COLLECTION_TYPE=$2
  local MONITORED_FIELDS=$3
  local NOTIFY_USER=$4
  local NOTIFY_SELLER=$5
  local FIRESTORE_COLLECTION=$6   # e.g. "orders"

  echo "-------------------------------------------------------------------"
  echo "Deploying: $NAME  (collection=$FIRESTORE_COLLECTION)"
  echo "-------------------------------------------------------------------"

  gcloud functions deploy "$NAME" \
    --gen2 \
    --runtime="$RUNTIME" \
    --region="$REGION" \
    --project="$PROJECT_ID" \
    --source="$FUNCTION_DIR" \
    --entry-point="$ENTRY_POINT" \
    --trigger-event-filters="type=google.cloud.firestore.document.v1.updated" \
    --trigger-event-filters="database=(default)" \
    --trigger-event-filters-path-pattern="document=${FIRESTORE_COLLECTION}/{docId}" \
    --set-env-vars="COLLECTION_TYPE=${COLLECTION_TYPE},MONITORED_FIELDS=${MONITORED_FIELDS},NOTIFY_USER=${NOTIFY_USER},NOTIFY_SELLER=${NOTIFY_SELLER},LOG_LEVEL=INFO" \
    --service-account="firebase-notifier@${PROJECT_ID}.iam.gserviceaccount.com" \
    --memory=256MB \
    --timeout=60s \
    --min-instances=0 \
    --max-instances=10
}

# 1. Order notifications – notify customer + seller on status changes
deploy_function \
  "order-notifications" \
  "order" \
  "status,orderStatus,paymentStatus" \
  "true" \
  "true" \
  "orders"

# 2. Product notifications – notify seller on verification/block/stock changes
deploy_function \
  "product-notifications" \
  "product" \
  "verificationStatus,blockStatus,stockStatus,stockQuantity" \
  "false" \
  "true" \
  "products"

# 3. Shop notifications – notify seller on shop status changes
deploy_function \
  "shop-notifications" \
  "shop" \
  "verificationStatus,blockStatus,isActive" \
  "false" \
  "true" \
  "shops"

echo ""
echo "All three Cloud Functions deployed successfully."
