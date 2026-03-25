# Local Development & Testing Guide

## Quick Start

### Prerequisites

1. **Go 1.21+**: [Download](https://golang.org/dl/)
2. **Firebase CLI**: `npm install -g firebase-tools`
3. **Google Cloud SDK**: [Install](https://cloud.google.com/sdk/docs/install)
4. **Docker**: [Install](https://docs.docker.com/get-docker/)

### Setup Local Environment

```bash
# Clone the repository
cd mandi-backend

# Install dependencies
cd cmd/cloudfunctions/enquiry-notification-handler
go mod download

# Set environment variables
export GOOGLE_APPLICATION_CREDENTIALS="C:\Users\HP\Documents\Locazar\mandi-backend\locazar-f20b6-c125ff67e902.json"
export LOG_LEVEL=DEBUG
export GCP_PROJECT=your-project-id
```

### Run Locally

```bash
# Start the function server
go run main.go

# Server should start on http://localhost:8080

# In another terminal, test with a sample event
curl -X POST http://localhost:8080/ProcessEnquiryUpdate \
  -H "Content-Type: application/json" \
  -d @test_event.json
```

## Testing

### Sample Test Event

Create `test_event.json`:

```json
{
  "id": "test-event-001",
  "data": {
    "value": {
      "name": "projects/YOUR_PROJECT/databases/(default)/documents/enquiries/ENQ-123",
      "fields": {
        "id": {"stringValue": "ENQ-123"},
        "queryId": {"stringValue": "Q-123"},
        "userId": {"stringValue": "USER-001"},
        "status": {"stringValue": "in_progress"},
        "assignedTo": {"stringValue": "ADMIN-001"},
        "assignedToName": {"stringValue": "John Doe"},
        "priority": {"stringValue": "high"},
        "subject": {"stringValue": "Product Inquiry"},
        "description": {"stringValue": "I have a question about product X"},
        "createdAt": {"timestampValue": "2024-01-01T00:00:00Z"},
        "updatedAt": {"timestampValue": "2024-01-02T12:30:45Z"}
      },
      "createTime": "2024-01-01T00:00:00Z",
      "updateTime": "2024-01-02T12:30:45Z"
    },
    "oldValue": {
      "name": "projects/YOUR_PROJECT/databases/(default)/documents/enquiries/ENQ-123",
      "fields": {
        "id": {"stringValue": "ENQ-123"},
        "queryId": {"stringValue": "Q-123"},
        "userId": {"stringValue": "USER-001"},
        "status": {"stringValue": "new"},
        "assignedTo": {"nullValue": "NULL_VALUE"},
        "priority": {"stringValue": "medium"},
        "subject": {"stringValue": "Product Inquiry"},
        "description": {"stringValue": "I have a question about product X"},
        "createdAt": {"timestampValue": "2024-01-01T00:00:00Z"},
        "updatedAt": {"timestampValue": "2024-01-01T00:00:00Z"}
      },
      "createTime": "2024-01-01T00:00:00Z",
      "updateTime": "2024-01-01T00:00:00Z"
    },
    "updateMask": {
      "fieldPaths": ["status", "assignedTo", "assignedToName"]
    }
  }
}
```

### Test Cases

#### Test 1: Status Change Only

```bash
# Test with only status change
curl -X POST http://localhost:8080/ProcessEnquiryUpdate \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test-1",
    "data": {
      "value": {
        "name": "projects/test/databases/(default)/documents/enquiries/doc1",
        "fields": {"status": {"stringValue": "resolved"}},
        "updateTime": "2024-01-02T12:00:00Z"
      },
      "oldValue": {
        "name": "projects/test/databases/(default)/documents/enquiries/doc1",
        "fields": {"status": {"stringValue": "in_progress"}},
        "updateTime": "2024-01-02T11:00:00Z"
      },
      "updateMask": {"fieldPaths": ["status"]}
    }
  }'
```

#### Test 2: No Significant Changes

```bash
# Test with non-monitored field change
curl -X POST http://localhost:8080/ProcessEnquiryUpdate \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test-2",
    "data": {
      "value": {
        "name": "projects/test/databases/(default)/documents/enquiries/doc2",
        "fields": {"viewCount": {"integerValue": "10"}},
        "updateTime": "2024-01-02T12:00:00Z"
      },
      "oldValue": {
        "name": "projects/test/databases/(default)/documents/enquiries/doc2",
        "fields": {"viewCount": {"integerValue": "9"}},
        "updateTime": "2024-01-02T11:00:00Z"
      },
      "updateMask": {"fieldPaths": ["viewCount"]}
    }
  }'
```

#### Test 3: Multiple Field Changes

```bash
# Test with multiple significant changes
curl -X POST http://localhost:8080/ProcessEnquiryUpdate \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test-3",
    "data": {
      "value": {
        "name": "projects/test/databases/(default)/documents/enquiries/doc3",
        "fields": {
          "status": {"stringValue": "in_progress"},
          "assignedTo": {"stringValue": "admin1"},
          "assignedToName": {"stringValue": "Admin User"},
          "priority": {"stringValue": "high"}
        },
        "updateTime": "2024-01-02T12:00:00Z"
      },
      "oldValue": {
        "name": "projects/test/databases/(default)/documents/enquiries/doc3",
        "fields": {
          "status": {"stringValue": "new"},
          "assignedTo": {"nullValue": "NULL_VALUE"},
          "priority": {"stringValue": "low"}
        },
        "updateTime": "2024-01-02T11:00:00Z"
      },
      "updateMask": {"fieldPaths": ["status", "assignedTo", "assignedToName", "priority"]}
    }
  }'
```

### Unit Testing

```bash
# Test the firestore parser
cd pkg/utils/firestore
go test -v -run TestParseFields

# Test the comparator
go test -v -run TestDetectChanges

# Test with coverage
go test -v -cover ./...
```

### Integration Testing

```bash
# Build the container
docker build -f cmd/cloudfunctions/enquiry-notification-handler/Dockerfile \
  -t enquiry-notification-handler:latest .

# Run container locally
docker run -p 8080:8080 \
  -e LOG_LEVEL=DEBUG \
  -e GCP_PROJECT=test \
  -e GOOGLE_APPLICATION_CREDENTIALS=/config/key.json \
  -v ~/.config/gcloud:/config \
  enquiry-notification-handler:latest

# Test with event
curl -X POST http://localhost:8080/ProcessEnquiryUpdate \
  -H "Content-Type: application/json" \
  -d @test_event.json
```

## Debugging

### Enable Debug Logging

```bash
# Set debug level
export LOG_LEVEL=DEBUG

# Run with verbose output
go run main.go -v
```

### Common Issues & Solutions

#### Issue: "Failed to parse event"

Check the JSON structure:
```bash
# Validate JSON
cat test_event.json | jq .

# Check field format
jq '.data.value.fields' test_event.json
```

#### Issue: "No recipients found"

1. Verify FCM tokens collection exists:
```bash
firebase firestore --project=YOUR_PROJECT
> db.collection('enquiry').doc('enq_1774297150').collection('fcmTokens').get()
```

2. Check token validity:
```bash
# Test FCM token
curl -X POST https://fcm.googleapis.com/v1/projects/YOUR_PROJECT/messages:send \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $(gcloud auth print-access-token)" \
  -d '{
    "message": {
      "token": "TOKEN_HERE",
      "notification": {
        "title": "Test",
        "body": "Testing FCM"
      }
    }
  }'
```

#### Issue: Timeouts

1. Check Firestore latency:
```bash
# List Firestore indexes
gcloud firestore indexes --database=(default)

# Create index if needed (for fcmTokens)
gcloud firestore indexes composite create \
  --collection-ids=users \
  --field-config=field-path=fcmTokens
```

2. Increase timeout in code or environment

### Useful Commands

```bash
# View real-time logs (requires Cloud Function deployed)
gcloud functions logs read enquiry-notification-handler --follow

# Test Firebase connectivity
gcloud firebase test ios run --help

# Check service account permissions
gcloud projects get-iam-policy YOUR_PROJECT

# List deployed functions
gcloud functions list --gen2

# Delete function for cleanup
gcloud functions delete enquiry-notification-handler --gen2
```

## Performance Profiling

### CPU/Memory Profiling

```bash
# Add profiling to main.go and run with pprof
go tool pprof http://localhost:6060/debug/pprof/heap
```

### Benchmarking

```bash
# Create benchmark tests
# in pkg/utils/firestore/parser_test.go

go test -bench=. -benchmem ./pkg/utils/firestore
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Deploy Cloud Function

on:
  push:
    branches: [main]
    paths:
      - 'cmd/cloudfunctions/enquiry-notification-handler/**'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Setup Cloud SDK
        uses: google-github-actions/setup-gcloud@v0
        with:
          project_id: ${{ secrets.GCP_PROJECT_ID }}
          service_account_key: ${{ secrets.GCP_SA_KEY }}
      
      - name: Deploy Function
        run: |
          gcloud functions deploy enquiry-notification-handler \
            --source=cmd/cloudfunctions/enquiry-notification-handler \
            --gen2 \
            --region=us-central1 \
            --runtime=go121
```

## Development Tips

1. **Use updateMask**: Always provide updateMask for efficiency
2. **Mock Firebase**: Use Firebase emulator for local testing
3. **Test edge cases**: Null values, empty strings, nested objects
4. **Monitor performance**: Check Cold start times and latency

---

For more information, see [README.md](./README.md)
