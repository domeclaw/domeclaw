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

# Build as domeclaw (DomeClaw-specific binary name)
RUN go build -ldflags="-s -w" -o domeclaw ./cmd/picoclaw

# ============================================================
# Stage 2: Runtime (root user)
# ============================================================
FROM alpine:3.19

# Install packages and configure DNS
RUN apk add --no-cache ca-certificates tzdata curl openssl && \
    update-ca-certificates && \
    echo "nameserver 8.8.8.8" > /etc/resolv.conf && \
    echo "nameserver 8.8.4.4" >> /etc/resolv.conf && \
    echo "nameserver 1.1.1.1" >> /etc/resolv.conf

# Copy binary
COPY --from=builder /build/domeclaw /usr/local/bin/domeclaw

# Expose ports
# 18790 - Gateway HTTP API
# 18795 - Webhook channel
EXPOSE 18790 18795

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -q --spider http://localhost:18790/health || exit 1

ENTRYPOINT ["domeclaw"]
CMD ["gateway"]
