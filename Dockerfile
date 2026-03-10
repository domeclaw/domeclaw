# DomeClaw - PicoClaw with Wallet & Webhook Support
# Multi-stage build for production deployment

# ============================================================
# Stage 1: Build
# ============================================================
FROM golang:1.25.7-alpine AS builder

RUN apk add --no-cache git make build-base

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Generate embedded files (workspace for onboard command)
RUN go generate ./...

# Build
RUN go build -ldflags="-s -w" -o picoclaw ./cmd/picoclaw

# ============================================================
# Stage 2: Runtime (root user)
# ============================================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata curl

# Copy binary
COPY --from=builder /build/picoclaw /usr/local/bin/picoclaw

# Expose ports
# 18790 - Gateway HTTP API
# 18795 - Webhook channel
EXPOSE 18790 18795

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -q --spider http://localhost:18790/health || exit 1

ENTRYPOINT ["picoclaw"]
CMD ["gateway"]
