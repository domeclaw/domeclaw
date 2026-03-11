# DomeClaw Project Rules

This document contains important rules and guidelines for the DomeClaw project (PicoClaw fork with wallet & webhook support).

## Project Structure

DomeClaw is a fork of PicoClaw with additional features:
- **Ethereum Wallet Integration** - Hotwallet mode for AI agents
- **Webhook Channel** - HTTP webhook for external integrations
- **Custom Workflows** - Simplified CI/CD for DomeClaw-specific builds

## Protected Files (DO NOT DELETE)

The following files/directories are DomeClaw-specific and must be preserved during merges from upstream:

### CI/CD Workflows
```
.github/workflows/nightly.yml
.github/workflows/pr.yml
.github/workflows/release.yml
.github/workflows/upload-tos.yml
```

### Wallet Functionality
```
cmd/picoclaw/internal/wallet/
pkg/commands/cmd_wallet.go
pkg/tools/wallet*.go
pkg/wallet/
config-with-wallet.json
dc-hotwallet-SKILL.md
```

### Webhook Channel
```
pkg/channels/webhook/
```

### Docker Deployment
```
Dockerfile
docker-compose.yml
docker-compose.override.yml.example
```

### Documentation
```
IMPLEMENTATION.md
debug-conversion.go
```

### Workspace Skills
```
workspace/skills/dc-hotwallet/SKILL.md
```

### Development Files
```
.trae/
```

## Merge Strategy from Upstream

When merging changes from `main` (upstream PicoClaw):

1. **Always keep DomeClaw-specific files** listed above
2. **Resolve conflicts carefully** in:
   - `pkg/config/config.go` - Keep both `Wallet` and `Voice` configs
   - `pkg/agent/loop.go` - Keep wallet tools registration
   - `cmd/picoclaw/internal/gateway/helpers.go` - Keep webhook import
3. **Test after merge** - Ensure wallet and webhook features still work

## Key Configuration

### Wallet Support
Wallet is enabled via config:
```json
{
  "wallet": {
    "enabled": true,
    "chains": {
      "clawswift": {
        "chain_id": 7441,
        "rpc_url": "https://rpc.clawswift.net",
        "explorer_url": "https://explorer.clawswift.net"
      }
    }
  }
}
```

### Webhook Channel
Webhook receives HTTP POST on port 18795:
```json
{
  "channels": {
    "webhook": {
      "enabled": true,
      "port": 18795,
      "token": "your-secret-token"
    }
  }
}
```

### Exec Tool Allow Patterns
The following commands are allowed in exec tool by default:
- `curl` - HTTP requests
- `cast` - Foundry Ethereum CLI

## Rebrand Options

The following branding files are prepared for easy rebrand:

| File | Brand | Usage |
|------|-------|-------|
| `pkg/config/domeclaw_branding.go` | **DomeClaw** (current) | Default branding |
| `pkg/config/mvpclaw_branding.go` | MVP Claw | Future option |
| `pkg/config/jfinclaw_branding.go` | JFIN Claw | Future option |
| `pkg/config/tokclaw_branding.go` | TOK Claw | Future option |

### How to Rebrand

1. **Backup current branding:**
   ```bash
   cp pkg/config/domeclaw_branding.go pkg/config/domeclaw_branding.go.bak
   ```

2. **Replace with desired brand:**
   ```bash
   # For MVP Claw
   cp pkg/config/mvpclaw_branding.go pkg/config/domeclaw_branding.go
   
   # For JFIN Claw
   cp pkg/config/jfinclaw_branding.go pkg/config/domeclaw_branding.go
   ```

3. **Rebuild:**
   ```bash
   go build -o domeclaw ./cmd/picoclaw
   ```

4. **Verify:**
   ```bash
   ./domeclaw version
   ```

### Branding Components

Each branding file contains:
- `Banner` - ASCII art logo
- `AppNameDisplay` - Display name
- `AppShortDescription` - CLI short description
- `AppLongDescription` - Detailed description

## Docker Usage

### Binary Name
DomeClaw builds the binary as `domeclaw` (not `picoclaw`) to distinguish from upstream:

```bash
# Local build
go build -o domeclaw ./cmd/picoclaw

# Docker build
Dockerfile builds as /usr/local/bin/domeclaw
```

### Docker Commands

```bash
# Build
docker build -t domeclaw .

# Run with volume
docker run -d \
  -p 18790:18790 \
  -p 18795:18795 \
  -v ./data:/root/.picoclaw \
  ghcr.io/domeclaw/domeclaw:latest
```

## Release Process

1. Push tag: `git tag v0.x.x && git push origin v0.x.x`
2. Workflows automatically:
   - Build binaries (linux/amd64, linux/arm64, darwin/arm64, windows/amd64)
   - Build and push Docker image to `ghcr.io/domeclaw/domeclaw`
   - Create GitHub release with artifacts

## Important Notes

- **Wallet Security**: PIN stored in plaintext - testnet only
- **Webhook Auth**: Use strong Bearer token in production
- **Platform Support**: Docker builds for linux/amd64 and linux/arm64
