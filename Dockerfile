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

# Build domeclaw binary
RUN go build -ldflags="-s -w" -o domeclaw ./cmd/picoclaw

# Build domeclaw-launcher binary (web console)
RUN apk add --no-cache nodejs npm gcc musl-dev && \
    npm install -g pnpm && \
    if [ -f web/frontend/package.json ]; then \
        cd web/frontend && CI=true pnpm install --no-frozen-lockfile && CI=true pnpm build:backend; \
    fi && \
    cd /build && \
    CGO_ENABLED=1 go build -ldflags="-s -w" -o domeclaw-launcher ./web/backend

# ============================================================
# Stage 2: Runtime (root user)
# ============================================================
FROM alpine:3.23.3

# Install packages including Node.js, npm, Python
RUN apk add --no-cache ca-certificates tzdata curl openssl nodejs npm python3 py3-pip && \
    update-ca-certificates && \
    # Install TypeScript globally
    npm install -g typescript && \
    # Create a symbolic link for python to python3 (common convention)
    ln -sf python3 /usr/bin/python

# Copy binaries
COPY --from=builder /build/domeclaw /usr/local/bin/domeclaw
COPY --from=builder /build/domeclaw-launcher /usr/local/bin/domeclaw-launcher

# Expose ports
# 18790 - Gateway HTTP API
# 18795 - Webhook channel
EXPOSE 18790 18795

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -q --spider http://localhost:18790/health || exit 1

ENTRYPOINT ["domeclaw"]
CMD ["gateway"]
