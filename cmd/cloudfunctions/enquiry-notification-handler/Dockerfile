# Multi-stage build for Enquiry Notification Cloud Function
# This Dockerfile builds a production-ready Cloud Function (Gen 2)

FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /workspace

# Copy the entire mandi-backend module (root context)
COPY . .

# Download dependencies using root go.mod
RUN go mod download

# Build the binary from the cloud function main package
# Cloud Functions expect the main package to expose a public function
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -a -installsuffix cgo \
    -o enquiry-notification-handler \
    ./cmd/cloudfunctions/enquiry-notification-handler

# Runtime stage
FROM alpine:latest

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user for security
RUN addgroup -S functions && adduser -S functions -G functions

WORKDIR /

# Copy binary from builder
COPY --from=builder /workspace/enquiry-notification-handler ./enquiry-notification-handler

# Use non-root user
USER functions

# Set working directory
WORKDIR /

# Cloud Functions Framework expects PORT environment variable
ENV PORT=8080

# Expose the port
EXPOSE 8080

# Start the function server
CMD ["./enquiry-notification-handler"]
