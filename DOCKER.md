# DomeClaw Docker Deployment Guide

## Quick Start

```bash
# 1. Build and start
docker compose up -d

# 2. Check logs
docker compose logs -f

# 3. Verify health
curl http://localhost:18790/health
```

## Setup Steps

### 1. Create Config Directory

```bash
mkdir -p data
```

### 2. Initial Config Setup

Run the onboard command to generate initial config:

```bash
docker compose run --rm picoclaw onboard
```

Or manually create `data/config.json`:

```json
{
  "llm": {
    "default": "openai",
    "openai": {
      "api_key": "YOUR_OPENAI_KEY",
      "model": "gpt-4o-mini"
    }
  },
  "channels": {
    "telegram": {
      "enabled": true,
      "bot_token": "YOUR_BOT_TOKEN"
    },
    "webhook": {
      "enabled": true,
      "port": 18795,
      "token": "YOUR_WEBHOOK_SECRET"
    }
  },
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

### 3. Setup Wallet (Hotwallet Mode)

```bash
# Create wallet
docker compose exec picoclaw picoclaw wallet create

# Or import existing
docker compose exec picoclaw picoclaw wallet import "0xYOUR_PRIVATE_KEY"
```

## Port Reference

| Port | Service | Description |
|------|---------|-------------|
| 18790 | Gateway | HTTP API for chat |
| 18795 | Webhook | External webhook receiver |

## Webhook Usage

Send messages to your bot via webhook:

```bash
curl -X POST http://localhost:18795/webhook \
  -H "Authorization: Bearer YOUR_WEBHOOK_SECRET" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Hello bot!",
    "chat_id": "user123",
    "metadata": {
      "target_channel": "telegram"
    }
  }'
```

## Common Commands

```bash
# Start
docker compose up -d

# Stop
docker compose down

# View logs
docker compose logs -f

# Restart
docker compose restart

# Shell access
docker compose exec picoclaw sh

# Wallet status
docker compose exec picoclaw picoclaw wallet status

# Check agent status
docker compose exec picoclaw wget -qO- http://localhost:18790/agents
```

## Data Persistence

All data is stored in `./data/`:

```
data/
├── config.json          # Main configuration
├── agents/              # Agent configs & skills
└── wallets/             # Wallet keystores
    ├── wallet.json
    └── pin.json
```

## Updating

```bash
# Pull latest code
git pull origin main

# Rebuild and restart
docker compose down
docker compose up -d --build
```
