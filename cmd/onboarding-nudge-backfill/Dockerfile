# Multi-stage build for the onboarding-nudge-backfill Cloud Run Job.
# A ONE-TIME manual tool (not scheduled) — anchors every currently-active
# shop into the onboarding-nudge sequence. Defaults to dry-run (CMD carries
# no -apply flag on purpose); pass it explicitly at execute time:
#   gcloud run jobs execute onboarding-nudge-backfill --args="-apply" --region=us-central2
# Mirrors cmd/enquiry-autoreject/Dockerfile.

FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /workspace

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -a -installsuffix cgo \
    -o onboarding-nudge-backfill \
    ./cmd/onboarding-nudge-backfill

FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN addgroup -S jobs && adduser -S jobs -G jobs

WORKDIR /

COPY --from=builder /workspace/onboarding-nudge-backfill ./onboarding-nudge-backfill

USER jobs

CMD ["./onboarding-nudge-backfill"]
