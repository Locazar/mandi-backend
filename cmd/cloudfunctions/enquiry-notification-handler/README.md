# Enquiry Notification Cloud Function (Gen 2) - Go Implementation

## Overview

This is a production-ready Google Cloud Function (Gen 2) that:

1. **Listens to Firestore Events**: Receives `google.cloud.firestore.document.v1.updated` events via Eventarc for the enquiries collection
2. **Compares Field Changes**: Safely parses and compares Firestore typed values
3. **Detects Significant Changes**: Only triggers notifications when monitored fields change (e.g., `status`, `assignedTo`)
4. **Sends FCM Notifications**: Uses Firebase Admin SDK to send push notifications to relevant users
5. **Implements Idempotency**: Prevents duplicate notifications with optional notification history tracking
6. **Production-Grade**: Includes comprehensive logging, error handling, and environment-based configuration

## Project Structure

```
cmd/cloudfunctions/enquiry-notification-handler/
├── main.go                    # Cloud Function entry point
├── go.mod                      # Go module definition
├── go.sum                      # Dependency locks
├── Dockerfile                  # Container build configuration
├── cloudbuild.yaml            # GCP Cloud Build configuration
└── README.md                   # This file

pkg/
├── domain/
│   └── firestore_event.go     # Event and notification domain types
├── service/notification/
│   ├── fcm_service.go         # FCM notification service
│   └── payload_builder.go     # Notification payload builder
└── utils/firestore/
    ├── parser.go               # Firestore field value parser
    ├── comparator.go           # Field comparison and change detection
    └── event_handler.go        # Event parsing and processing
```

## Architecture

### Event Flow

```
Firestore Document Update
        ↓
    Eventarc
        ↓
Cloud Function (Gen 2)
        ↓
    Event Parser (firestore/event_handler.go)
        ↓
    Field Value Parser (firestore/parser.go)
        ↓
    Field Comparator (firestore/comparator.go)
        ↓
    Payload Builder (notification/payload_builder.go)
        ↓
    FCM Service (notification/fcm_service.go)
        ↓
Firebase Messaging API
        ↓
User Devices (iOS, Android, Web)
```

### Key Components

#### 1. Event Handler (`pkg/utils/firestore/event_handler.go`)
- Parses raw Firestore events into structured format
- Extracts document ID and path from full resource names
- Detects field changes using updateMask or full comparison

#### 2. Field Parser (`pkg/utils/firestore/parser.go`)
- Safely extracts Firestore typed values:
  - stringValue, integerValue, doubleValue
  - booleanValue, timestampValue, nullValue
  - arrayValue, mapValue, geoPointValue, bytesValue
- Handles nested structures and fallback values

#### 3. Field Comparator (`pkg/utils/firestore/comparator.go`)
- Compares old and new field values
- Supports configurable monitored fields
- Implements change detection with idempotency
- Returns list of significant changes

#### 4. Notification Service (`pkg/service/notification/fcm_service.go`)
- Initializes Firebase Admin SDK
- Retrieves FCM tokens from Firestore
- Constructs platform-specific notifications (Android, iOS, Web)
- Sends messages with proper error handling
- Records notification history for idempotency checking

#### 5. Payload Builder (`pkg/service/notification/payload_builder.go`)
- Generates user-friendly notification titles and bodies
- Creates context-aware messages based on field changes
- Extracts enquiry metadata (ID, assignee, user)
- Formats action URLs for in-app navigation

## Deployment

### Prerequisites

1. **Google Cloud Project** with:
   - Firebase enabled
   - Firestore database
   - Cloud Build enabled
   - Cloud Run enabled (for Gen 2 functions)
   - Service account with permissions:
     - Cloud Functions Developer
     - Cloud Run Admin
     - Firebase Admin
     - Eventarc Event Receiver

2. **Firestore Structure**:
   ```
   enquiries/{docId}
   └── status, assignedTo, priority, etc.
   
   users/{userId}
   ├── fcmTokens/{tokenDoc}
   │   └── token, isActive, createdAt
   └── ...
   ```

3. **Environment Setup**:
   ```bash
   gcloud auth login
   gcloud config set project YOUR_PROJECT_ID
   ```

### Quick Deploy (Recommended)

#### Option 1: Using Google Cloud Console

1. Go to Cloud Functions in GCP Console
2. Create new function:
   - **Environment**: 2nd gen
   - **Region**: Choose appropriate region (e.g., `us-central1`)
   - **Trigger Type**: Cloud Pub/Sub
   - **Authentication**: Require authentication
   
3. For source code:
   - **Runtime**: Go 1.21
   - **Entry point**: `ProcessEnquiryUpdate`
   - Copy the ZIP of the `cmd/cloudfunctions/enquiry-notification-handler/` directory

4. **Runtime settings**:
   - **Memory**: 512 MB (minimum)
   - **Timeout**: 60 seconds
   - **Runtime service account**: Your Firebase service account
   - **Concurrency**: 100

#### Option 2: Using gcloud CLI

```bash
# Set variables
PROJECT_ID="your-project-id"
REGION="us-central1"
SERVICE_ACCOUNT="firebase-admin@${PROJECT_ID}.iam.gserviceaccount.com"

# Deploy the function
gcloud functions deploy enquiry-notification-handler \
  --runtime go121 \
  --trigger-event google.cloud.firestore.document.v1.updated \
  --trigger-resource "projects/${PROJECT_ID}/databases/(default)/documents/enquiries/{docId}" \
  --entry-point ProcessEnquiryUpdate \
  --source ./cmd/cloudfunctions/enquiry-notification-handler \
  --service-account ${SERVICE_ACCOUNT} \
  --region ${REGION} \
  --memory 512MB \
  --timeout 60s \
  --gen2 \
  --set-env-vars LOG_LEVEL=INFO,ENABLE_IDEMPOTENCY_CHECK=true,MONITORED_FIELDS="status,assignedTo,priority"
```

#### Option 3: Using Terraform (IaC)

```hcl
resource "google_cloudfunctions_function" "enquiry_notification" {
  name      = "enquiry-notification-handler"
  runtime   = "go121"
  
  event_type       = "google.cloud.firestore.document.v1.updated"
  event_filters {
    attribute = "database"
    value     = "(default)"
    operator  = "="
  }
  event_filters {
    attribute = "document"
    value     = "enquiries/{docId}"
    operator  = "="
  }
  
  source_repository {
    url = google_sourcerepo_repository.repo.clone_https_uri
  }
  
  service_account_email = google_service_account.cf_sa.email
  
  environment_variables = {
    LOG_LEVEL                   = "INFO"
    ENABLE_IDEMPOTENCY_CHECK    = "true"
    MONITORED_FIELDS            = "status,assignedTo,priority"
  }
  
  timeout             = 60
  available_memory_mb = 512
  gen2               = true
}
```

#### Option 4: Using Cloud Build

```bash
# Push to Cloud Source Repository (or GitHub)
git push

# Trigger build and deployment
gcloud builds submit \
  --config=cmd/cloudfunctions/enquiry-notification-handler/cloudbuild.yaml
```

### Setting up Eventarc Trigger

**Important**: While deploying via CLI/Terraform, set up Eventarc trigger:

```bash
# Create Eventarc trigger to listen to Firestore updates
gcloud eventarc triggers create enquiry-notification-trigger \
  --location=${REGION} \
  --destination-cf=enquiry-notification-handler \
  --destination-cf-region=${REGION} \
  --event-filters="type=google.cloud.firestore.document.v1.updated" \
  --event-filters="database=(default)" \
  --event-filters="document=enquiries/*" \
  --service-account=${SERVICE_ACCOUNT}
```

## Configuration

### Environment Variables

```bash
# Log level: DEBUG, INFO, WARN, ERROR (default: INFO)
LOG_LEVEL=INFO

# Enable notification history tracking for idempotency (default: false)
ENABLE_IDEMPOTENCY_CHECK=true

# Monitored fields (comma-separated, default: status,assignedTo,priority,resolutionDate,closedAt,tags)
MONITORED_FIELDS=status,assignedTo,priority,resolutionDate,closedAt

# GCP Project ID (auto-detected if not set)
GCP_PROJECT=your-project-id

# Firebase Realtime Database URL (optional)
FIREBASE_DB_URL=https://your-project.firebaseio.com
```

### Monitored Fields

Default monitored fields trigger notifications:
- `status` - Enquiry status changes
- `assignedTo` - Assignment changes
- `assignedToName` - Assignment name changes
- `priority` - Priority level changes
- `resolutionDate` - Expected resolution date
- `closedAt` - Enquiry closure
- `tags` - Tag changes
- `category`, `type`, `respondedAt`, `department`, `customStatus`

Override via `MONITORED_FIELDS` environment variable:
```bash
MONITORED_FIELDS="status,assignedTo,customField1,customField2"
```

### Firebase Firestore Structure

**Required collection structure for tokens**:
```
users/{userId}/fcmTokens/{tokenId}
├── token: "fcm_token_here"
├── isActive: true
├── platform: "android" | "ios" | "web"
├── createdAt: Timestamp
└── updatedAt: Timestamp

notificationHistory/{docId}
├── userId: "user_id"
├── documentId: "enquiry_id"
├── messageId: "fcm_message_id"
├── sentAt: Timestamp
└── expireAt: Timestamp
```

## Testing

### Local Testing

1. **Mock Event**:
```bash
cat > test_event.json << 'EOF'
{
  "id": "test-1",
  "data": {
    "value": {
      "name": "projects/YOUR_PROJECT/databases/(default)/documents/enquiries/doc-123",
      "fields": {
        "status": {"stringValue": "in_progress"},
        "assignedTo": {"stringValue": "admin-001"},
        "assignedToName": {"stringValue": "John Doe"},
        "priority": {"stringValue": "high"},
        "createdAt": {"timestampValue": "2024-01-01T00:00:00Z"}
      },
      "createTime": "2024-01-01T00:00:00Z",
      "updateTime": "2024-01-02T12:30:45Z"
    },
    "oldValue": {
      "name": "projects/YOUR_PROJECT/databases/(default)/documents/enquiries/doc-123",
      "fields": {
        "status": {"stringValue": "new"},
        "assignedTo": {"nullValue": "NULL_VALUE"},
        "priority": {"stringValue": "medium"}
      },
      "createTime": "2024-01-01T00:00:00Z",
      "updateTime": "2024-01-01T00:00:00Z"
    },
    "updateMask": {
      "fieldPaths": ["status", "assignedTo", "assignedToName"]
    }
  }
}
EOF
```

2. **Run Locally**:
```bash
cd cmd/cloudfunctions/enquiry-notification-handler

# Set up local environment
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"
export LOG_LEVEL=DEBUG
export ENABLE_IDEMPOTENCY_CHECK=false

# Run the function locally
go run main.go

# In another terminal, test with local event
curl -X POST http://localhost:8080/ProcessEnquiryUpdate \
  -H "Content-Type: application/json" \
  -d @test_event.json
```

### Production Testing (Dry Run)

1. **Create test document** in Firestore enquiries collection
2. **Monitor Cloud Function logs**:
```bash
gcloud functions logs read enquiry-notification-handler \
  --limit 50 \
  --region ${REGION}
```

3. **Update a monitored field**:
```javascript
// Firebase console or SDK
db.collection('enquiry').doc('test-doc').update({
  status: 'in_progress'  // Will trigger notification
});
```

4. **Check notification delivery**:  
   - View Firebase Console → Cloud Messaging → Sent messages
   - Check device notifications if tokens are registered

### Unit Tests (Optional)

```bash
# Run tests for utility functions
cd pkg/utils/firestore
go test -v

# Run tests for notification service
cd pkg/service/notification
go test -v
```

## Monitoring & Debugging

### View Function Logs

```bash
# Stream live logs
gcloud functions logs read enquiry-notification-handler \
  --region ${REGION} \
  --follow

# View specific time range
gcloud functions logs read enquiry-notification-handler \
  --region ${REGION} \
  --limit 100 \
  --filter 'timestamp>="2024-01-15T09:00:00Z"'
```

### Check Function Metrics

```bash
# View invocations count
gcloud monitoring read \
  --filter 'resource.type="cloud_function" AND metric.type="cloudfunctions.googleapis.com/execution_count"' \
  --format=table

# View error count
gcloud monitoring read \
  --filter 'resource.type="cloud_function" AND metric.type="cloudfunctions.googleapis.com/execution_times" AND metric.status="ERROR"'
```

### CloudTrace Integration

Calls are automatically traced in Cloud Trace. View in:
- GCP Console → Cloud Trace → Trace List
- Filter by function name: `enquiry-notification-handler`

### Error Troubleshooting

**Problem**: "Failed to parse event"
- **Cause**: Event structure mismatch
- **Solution**: Check Firestore document format, ensure fields are properly typed

**Problem**: "No recipients found"
- **Cause**: FCM tokens missing for user
- **Solution**: Ensure user has registered FCM tokens in `users/{userId}/fcmTokens/`

**Problem**: "No significant changes detected"
- **Cause**: Updated fields not in monitored list
- **Solution**: Add fields to `MONITORED_FIELDS` env var or update default list

**Problem**: Timeout errors
- **Cause**: Firestore queries taking too long
- **Solution**: Increase function timeout, add indexes to fcmTokens collection

**Problem**: "Invalid CloudEvent"
- **Cause**: Wrong trigger type
- **Solution**: Ensure trigger is `google.cloud.firestore.document.v1.updated`, not Pub/Sub

## Performance & Scalability

### Performance Characteristics

- **Cold start**: ~2-3 seconds (Go functions with Firebase SDK)
- **Warm execution**: ~200-500ms for typical event
- **Token lookup**: ~100-300ms (varies by Firestore latency)
- **FCM send**: ~50-200ms per token

### Scaling

- **Concurrency**: Default 100, can be increased
- **Memory**: 512 MB (recommended), can be increased to 1GB for faster execution
- **Timeout**: 60 seconds (sufficient for ~100 tokens)

### Cost Optimization

1. **Use updateMask**: Significantly reduces comparison time
2. **Batch queries**: Combine token collections
3. **Selective monitoring**: Only monitor necessary fields
4. **Regional deployment**: Choose region closest to your Firestore

## Security Best Practices

1. **Service Account**: Use dedicated service account with minimal permissions
2. **Secrets**: Store API keys in Secret Manager, not environment variables
3. **Authentication**: Enable authentication on Cloud Function
4. **Encryption**: Use TLS for all network communication
5. **Audit**: Enable Cloud Audit Logs for function execution
6. **Rate Limiting**: Implement backpressure for notification floods

### Minimal IAM Permissions

```yaml
roles/firebase.admin              # For Firebase Admin SDK
roles/cloudfunctions.developer    # For function execution
roles/logging.logWriter           # For logging
roles/monitoring.metricWriter     # For metrics
roles/cloudkms.cryptoKeyDecrypter # For encryption keys (if used)
```

## Known Limitations & Future Enhancements

### Limitations
- Single document updates only (batch updates handled individually)
- Maximum 100 concurrent executions
- 540-second maximum timeout for event processing

### Planned Enhancements
- [ ] Batch notification aggregation
- [ ] Custom notification templates
- [ ] A/B testing for notification content
- [ ] Delivery time optimization
- [ ] Analytics and performance metrics
- [ ] Webhook integration for custom handlers
- [ ] Template-based multilingual notifications

## Troubleshooting Checklist

- [ ] Cloud Function deployed and active
- [ ] Eventarc trigger is enabled and listening to correct resource
- [ ] Service account has Firebase admin permissions
- [ ] Firestore database exists with enquiries collection
- [ ] Users have fcmTokens collection and valid tokens
- [ ] Monitored fields match your Firestore schema
- [ ] Log level is set appropriately for debugging
- [ ] Firebase SDK credentials are properly configured
- [ ] Cloud Functions API is enabled
- [ ] necessary quotas are available

## Related Documentation

- [Google Cloud Functions](https://cloud.google.com/functions/docs)
- [Eventarc Documentation](https://cloud.google.com/eventarc/docs)
- [Firebase Admin SDK (Go)](https://firebase.google.com/docs/database/admin/start)
- [Cloud Pub/Sub to Firestore Events](https://cloud.google.com/firestore/docs/audit-logs)

## Support & Contribution

For issues, questions, or contributions:
1. Check the troubleshooting section above
2. Review Cloud Function logs
3. Open an issue with detailed logs and configuration

---

**Version**: 1.0.0  
**Last Updated**: 2024-03-24  
**Status**: Production Ready
