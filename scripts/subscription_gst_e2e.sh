#!/bin/bash
#
# End-to-end check for the GST snapshot on subscription orders.
#
# Creates a real subscription order via the running API (which in turn creates a
# Razorpay TEST order) and verifies the response carries the GST breakdown that
# is snapshotted onto the order:
#   - gst_rate_basis_points  (e.g. 1800 = 18.00%)
#   - gst_amount             (GST portion already included in `amount`, in paise)
#
# The server must be running with a Razorpay TEST key and GST_PERCENT_BASIS_POINTS
# set (defaults to 1800). Provide a valid user bearer token via USER_TOKEN — the
# token is read from the environment and is never stored in this script.
#
# Usage:
#   BASE_URL=http://localhost:3000 \
#   USER_TOKEN=<user_jwt> \
#   [PLAN_ID=<plan_id>] \
#   [EXPECTED_GST_BPS=1800] \
#   ./scripts/subscription_gst_e2e.sh

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:3000}"
EXPECTED_GST_BPS="${EXPECTED_GST_BPS:-1800}"

if [ -z "${USER_TOKEN:-}" ]; then
  echo "✗ USER_TOKEN is required (a user bearer token). Aborting." >&2
  exit 1
fi

AUTH_HEADER="Authorization: Bearer ${USER_TOKEN}"

echo "=== Subscription GST E2E ==="
echo "Base URL: ${BASE_URL}"

# 1. Resolve a plan id (first paid plan) unless one was supplied.
PLAN_ID="${PLAN_ID:-}"
if [ -z "${PLAN_ID}" ]; then
  echo "→ Fetching paid plans..."
  PLANS_JSON=$(curl -sS -H "${AUTH_HEADER}" "${BASE_URL}/api/subscriptions/plans")
  PLAN_ID=$(echo "${PLANS_JSON}" | jq -r '.data[0].id // .data[0].ID // empty')
  if [ -z "${PLAN_ID}" ]; then
    echo "✗ Could not resolve a plan id from /subscriptions/plans response:" >&2
    echo "${PLANS_JSON}" >&2
    exit 1
  fi
fi
echo "→ Using plan_id: ${PLAN_ID}"

# 2. Create the order (this hits Razorpay's test API).
echo "→ Creating subscription order..."
ORDER_JSON=$(curl -sS -X POST \
  -H "${AUTH_HEADER}" \
  -H "Content-Type: application/json" \
  -d "{\"plan_id\":\"${PLAN_ID}\"}" \
  "${BASE_URL}/api/subscriptions/create-order")

echo "Response:"
echo "${ORDER_JSON}" | jq . 2>/dev/null || echo "${ORDER_JSON}"

# 3. Assert the GST breakdown is present and correct.
AMOUNT=$(echo "${ORDER_JSON}" | jq -r '.data.amount // empty')
GST_BPS=$(echo "${ORDER_JSON}" | jq -r '.data.gst_rate_basis_points // empty')
GST_AMOUNT=$(echo "${ORDER_JSON}" | jq -r '.data.gst_amount // empty')

if [ -z "${GST_BPS}" ] || [ -z "${GST_AMOUNT}" ]; then
  echo "✗ Response is missing gst_rate_basis_points / gst_amount." >&2
  exit 1
fi

echo ""
echo "amount (inclusive) : ${AMOUNT} paise"
echo "gst_rate           : ${GST_BPS} bps"
echo "gst_amount         : ${GST_AMOUNT} paise"

FAIL=0
if [ "${GST_BPS}" != "${EXPECTED_GST_BPS}" ]; then
  echo "✗ gst_rate_basis_points = ${GST_BPS}, expected ${EXPECTED_GST_BPS}" >&2
  FAIL=1
fi
# Expected inclusive GST = amount * bps / (10000 + bps), rounded to nearest paise.
if [ -n "${AMOUNT}" ] && [ "${GST_BPS}" -gt 0 ]; then
  EXPECTED_GST=$(( (AMOUNT * GST_BPS + (10000 + GST_BPS) / 2) / (10000 + GST_BPS) ))
  if [ "${GST_AMOUNT}" != "${EXPECTED_GST}" ]; then
    echo "✗ gst_amount = ${GST_AMOUNT}, expected ${EXPECTED_GST} (= ${AMOUNT} * ${GST_BPS} / (10000 + ${GST_BPS}))" >&2
    FAIL=1
  fi
fi

if [ "${FAIL}" -ne 0 ]; then
  echo "=== E2E FAILED ===" >&2
  exit 1
fi

echo ""
echo "✓ GST snapshot verified on the created order."
echo "=== E2E PASSED ==="
