# Multi-stage build for the enquiry-autoreject Cloud Run Job.
# This is a one-shot batch binary (not an HTTP service) — triggered on a
# schedule by Cloud Scheduler via "gcloud run jobs execute", runs once, exits.

FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /workspace

# Copy the entire mandi-backend module (root context)
COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -a -installsuffix cgo \
    -o enquiry-autoreject \
    ./cmd/enquiry-autoreject

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN addgroup -S jobs && adduser -S jobs -G jobs

WORKDIR /

COPY --from=builder /workspace/enquiry-autoreject ./enquiry-autoreject

USER jobs

# -apply is required — without it the job only logs a dry-run and never
# writes/notifies, which would silently do nothing when run on a schedule.
CMD ["./enquiry-autoreject", "-apply"]
