#!/bin/bash

echo "=== Firebase Setup Verification ==="
echo ""

# Check if credentials file exists
CRED_FILE="${FIREBASE_CREDENTIALS_FILE:-locazar-f20b6-f024a4597849.json}"
echo "Looking for credentials file: $CRED_FILE"

if [ -f "$CRED_FILE" ]; then
    echo "✓ Credentials file found"
    echo ""
    echo "Credentials file contents (with sensitive data masked):"
    cat "$CRED_FILE" | jq '{type, project_id, client_email}' 2>/dev/null || cat "$CRED_FILE"
else
    echo "✗ Credentials file NOT found at: $CRED_FILE"
    echo "  Current directory: $(pwd)"
    echo "  Available JSON files:"
    find . -name "*.json" -type f 2>/dev/null | head -10
fi

echo ""
echo "=== Checking environment variables ==="
echo "FIREBASE_CREDENTIALS_FILE=${FIREBASE_CREDENTIALS_FILE:-not set}"
echo ""

echo "=== Recommended fixes ==="
echo ""
echo "1. Ensure credentials file path:"
echo "   export FIREBASE_CREDENTIALS_FILE=/absolute/path/to/locazar-f20b6-f024a4597849.json"
echo ""
echo "2. Verify service account has these permissions:"
echo "   - Firebase Cloud Messaging Admin"
echo "   - Service Accounts Admin"
echo ""
echo "3. Check that the project_id in credentials matches your Firebase project"
echo ""
echo "4. In Firebase Console, ensure the service account has Editor role"
