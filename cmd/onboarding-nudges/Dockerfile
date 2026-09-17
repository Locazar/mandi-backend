# Multi-stage build for the onboarding-nudges Cloud Run Job.
# This is a one-shot batch binary (not an HTTP service) — triggered on a
# schedule by Cloud Scheduler via "gcloud run jobs execute", runs one sweep,
# exits. Mirrors cmd/enquiry-autoreject/Dockerfile.

FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /workspace

# Copy the entire mandi-backend module (root context)
COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -a -installsuffix cgo \
    -o onboarding-nudges \
    ./cmd/onboarding-nudges

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN addgroup -S jobs && adduser -S jobs -G jobs

WORKDIR /

COPY --from=builder /workspace/onboarding-nudges ./onboarding-nudges

USER jobs

CMD ["./onboarding-nudges"]
