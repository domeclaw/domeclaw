# DomeClaw - PicoClaw with Wallet & Webhook Support
# Multi-stage build for production deployment
# Using Ubuntu 24.04 (glibc 2.39) for tempo wallet compatibility

# ============================================================
# Stage 1: Build
# ============================================================
FROM ubuntu:24.04 AS builder

# Install Go, Node.js 20, and build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates curl git make build-essential gcc \
    && rm -rf /var/lib/apt/lists/*

# Install Go 1.23
RUN curl -fsSL https://go.dev/dl/go1.23.4.linux-$(dpkg --print-architecture).tar.gz | tar -C /usr/local -xz \
    && ln -s /usr/local/go/bin/go /usr/local/bin/go \
    && ln -s /usr/local/go/bin/gofmt /usr/local/bin/gofmt

ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH=/go
ENV GOBIN=/go/bin

# Install Node.js 20 from NodeSource (Vite requires Node 20+)
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

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
RUN npm install -g pnpm \
    && if [ -f web/frontend/package.json ]; then \
        cd web/frontend && CI=true pnpm install --no-frozen-lockfile && CI=true pnpm build:backend; \
    fi \
    && cd /build \
    && CGO_ENABLED=1 go build -ldflags="-s -w" -o domeclaw-launcher ./web/backend

# ============================================================
# Stage 2: Runtime (root user)
# ============================================================
FROM ubuntu:24.04

# Install packages including Node.js, npm, Python, and glibc runtime for tempo
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    curl \
    wget \
    openssl \
    python3 \
    python3-pip \
    && update-ca-certificates \
    && ln -sf python3 /usr/bin/python \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js 20 from NodeSource for runtime
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && npm install -g typescript \
    && rm -rf /var/lib/apt/lists/*

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
