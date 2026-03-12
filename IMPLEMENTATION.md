# DomeClaw Wallet Implementation Guide

## Overview

This document describes the AI-powered wallet tools implementation for DomeClaw (picoclaw fork), enabling natural language wallet operations on ClawSwift chain.

## Implemented Features

### AI Wallet Tools

| Tool | Description | Use Case |
|------|-------------|----------|
| `query_wallet_balance` | Query wallet balance with explorer link | "wallet มีเหรียญอะไรบ้าง" |
| `wallet_transfer` | Transfer tokens with auto PIN read | "ส่ง 0.01 CLAW ให้ 0x..." |
| `query_contract_call` | Read smart contract data | "เช็ค balanceOf" |
| `execute_contract_write` | Write to smart contract | "approve ให้..." |

### Webhook Channel

Inbound webhook for external integrations - allows external services to POST messages that get processed and broadcast to configured channels.

| Feature | Description |
|---------|-------------|
| HTTP POST endpoint | `/webhook` on port 18795 |
| Authentication | Bearer token |
| Broadcast | To target_channel (default: telegram) |

**Usage:**
```bash
curl -X POST http://localhost:18795/webhook \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Hello from webhook",
    "chat_id": "-123456789",
    "metadata": {"target_channel": "telegram"}
  }'
```

### Key Components

```
pkg/tools/wallet_query.go      - query_wallet_balance tool
pkg/tools/wallet_transfer.go   - wallet_transfer tool
pkg/tools/wallet.go            - Contract tools (query_contract_call, execute_contract_write)
pkg/wallet/                    - Wallet service layer
pkg/commands/cmd_wallet.go     - /wallet slash commands
pkg/agent/loop.go              - Tool registration (lines 236-242)
pkg/channels/webhook/          - Webhook channel for external integrations
```

## Files Modified from Upstream

### Core Files (Risk of Conflict on Sync)
```
pkg/agent/loop.go              - Added wallet tools registration
cmd/picoclaw/internal/gateway/helpers.go - Added webhook channel import
pkg/commands/builtin.go        - Added walletCommand()
pkg/config/config.go           - Added Wallet config section + WebhookConfig
pkg/config/defaults.go         - Added wallet defaults + webhook defaults
cmd/picoclaw/main.go           - Removed CLI wallet command
cmd/picoclaw/main_test.go      - Updated tests
pkg/channels/manager.go        - Added webhook channel initialization
```

### New Files (No Conflict)
```
pkg/tools/wallet_query.go
pkg/tools/wallet_transfer.go
pkg/wallet/config.go
pkg/wallet/erc20.go
pkg/wallet/errors.go
pkg/wallet/service.go
pkg/wallet/service_test.go
pkg/wallet/types.go
pkg/channels/webhook/webhook.go
pkg/channels/webhook/init.go
```

## Configuration

### Webhook Config

```json
{
  "channels": {
    "webhook": {
      "enabled": true,
      "token": "your_secret_token",
      "host": "0.0.0.0",
      "port": 18795,
      "path": "/webhook"
    }
  }
}
```

### Wallet Config

```json
{
  "wallet": {
    "enabled": true,
    "chains": [
      {
        "name": "ClawSwift",
        "chain_id": 7441,
        "rpc": "https://exp.clawswift.net/rpc",
        "explorer": "https://exp.clawswift.net",
        "currency": "CLAW",
        "is_native": false,
        "gas_token": "0x20c0000000000000000000000000000000000000",
        "gas_token_name": "CLAW",
        "decimal": 16
      }
    ]
  }
}
```

## Sync Strategy from Upstream

### Preparation
```bash
# 1. Create backup branch
git checkout domeclaw
git branch domeclaw-backup-$(date +%Y%m%d)

# 2. Fetch latest upstream
git fetch origin main
```

### Sync Methods

#### Method 1: Rebase (Recommended)
```bash
git checkout domeclaw
git rebase origin/main

# If conflicts occur, resolve them:
git add <conflicted-files>
git rebase --continue
```

#### Method 2: Merge
```bash
git checkout domeclaw
git merge origin/main

# Resolve conflicts if any
git add <conflicted-files>
git commit
```

### Expected Conflict Areas

#### 1. `pkg/agent/loop.go` (Most Likely)
**Our additions:**
```go
// Around line 236
// Wallet tools (hotwallet mode - AI can query, transfer, and interact with contracts)
if cfg.Wallet.Enabled {
    agent.Tools.Register(tools.NewWalletQueryTool(agent.Workspace, cfg))
    agent.Tools.Register(tools.NewWalletTransferTool(agent.Workspace, cfg))
    agent.Tools.Register(tools.NewQueryContractCallTool(cfg))
    agent.Tools.Register(tools.NewExecuteContractWriteTool(cfg))
}
```

**Resolution:** Keep our wallet tools registration block.

#### 2. `cmd/picoclaw/internal/gateway/helpers.go`
**Our addition:**
```go
_ "github.com/sipeed/picoclaw/pkg/channels/webhook" // Register webhook channel factory
```

**Resolution:** Keep the webhook import.

#### 3. `pkg/commands/builtin.go`
**Our additions:**
```go
func BuiltinDefinitions() []Definition {
    return []Definition{
        // ... other commands ...
        walletCommand(),  // <-- Keep this
    }
}
```

**Resolution:** Ensure `walletCommand()` remains in the list.

#### 4. `pkg/config/config.go`
**Our additions:**
```go
type Config struct {
    // ... other fields ...
    Wallet WalletConfig `json:"wallet"`  // <-- Keep this
}

type ChannelsConfig struct {
    // ... other channels ...
    Webhook WebhookConfig `json:"webhook"`  // <-- Keep this
}
```

**Resolution:** Keep the Wallet and Webhook fields.

#### 5. `pkg/config/defaults.go`
**Our additions:**
```go
Webhook: WebhookConfig{
    Enabled: false,
    Token:   "",
    Host:    "0.0.0.0",
    Port:    18795,
    Path:    "/webhook",
},
Wallet: WalletConfig{
    Enabled: false,
    Chains: [...]
}
```

**Resolution:** Keep both Webhook and Wallet default configs.

#### 6. `pkg/channels/manager.go`
**Our addition:**
```go
if m.config.Channels.Webhook.Enabled {
    m.initChannel("webhook", "Webhook")
}
```

**Resolution:** Keep the webhook channel initialization.

#### 7. `cmd/picoclaw/main.go`
**Our change:**
```go
// REMOVED: wallet import and command registration
// Keep it removed - we use tools instead of CLI
```

**Resolution:** Ensure wallet CLI command stays removed (we use Telegram/tools only).

## Post-Sync Checklist

After syncing from upstream:

- [ ] Build succeeds: `go build -o picoclaw ./cmd/picoclaw`
- [ ] Wallet tools registered: Check `pkg/agent/loop.go`
- [ ] Webhook channel registered: Check `cmd/picoclaw/internal/gateway/helpers.go`
- [ ] Config loads: Verify `pkg/config/config.go` has Wallet and Webhook fields
- [ ] Commands available: Check `/wallet` works in Telegram
- [ ] Tools work: Test natural language queries
- [ ] Webhook works: Test POST to port 18795

## Testing

### Wallet Features

#### Natural Language Queries
```
"wallet มีเหรียญอะไรบ้าง"     -> query_wallet_balance
"เช็คยอดเงิน"                   -> query_wallet_balance
"ส่ง 0.01 CLAW ให้ 0x..."      -> wallet_transfer
"โอนเงินให้เพื่อน"               -> wallet_transfer
```

#### Slash Commands
```
/wallet create [password]
/wallet info
/wallet transfer [to] [amount]
/wallet transfer_token [to] [amount]
/wallet chain
/wallet call [contract] [abi] [method] [args...]
/wallet write [contract] [abi] [method] [value] [args...]
```

### Webhook

```bash
# Test webhook
curl -X POST http://localhost:18795/webhook \
  -H "Authorization: Bearer test_token" \
  -H "Content-Type: application/json" \
  -d '{
    "message": ">balance เหลือเท่าไหร่",
    "chat_id": "-5129639667",
    "metadata": {"target_channel": "telegram"}
  }'
```

## Troubleshooting

### Issue: Tools not appearing
**Check:** `pkg/agent/loop.go` has wallet tools registration
**Verify:** `cfg.Wallet.Enabled` is true in config.json

### Issue: Webhook not starting
**Check:** Port 18795 is not in use: `lsof -i :18795`
**Check:** `cmd/picoclaw/internal/gateway/helpers.go` has webhook import
**Verify:** `cfg.Channels.Webhook.Enabled` is true in config.json

### Issue: Transfer fails with PIN error
**Check:** `workspace/wallets/pin.json` exists with correct format:
```json
{"password": "your-pin"}
```

### Issue: Cannot connect to blockchain
**Check:** `config.json` has correct Wallet.Chains configuration
**Verify:** RPC endpoint is accessible

## Network Details

- **Chain**: ClawSwift
- **Chain ID**: 7441
- **RPC**: https://exp.clawswift.net/rpc
- **Explorer**: https://exp.clawswift.net
- **Gas Token**: 0x20c0000000000000000000000000000000000000
- **Decimals**: 16

## Security Notes

⚠️ **Hotwallet Mode:**
- PIN stored in plaintext at `workspace/wallets/pin.json`
- AI has direct keystore access
- Use for testnet only
- Never store large amounts

⚠️ **Webhook:**
- Token should be kept secret
- Use HTTPS in production
- Validate sender IPs if possible

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 3.0.0 | 2026-03-11 | Added Webhook channel for external integrations |
| 2.0.0 | 2026-03-10 | Added AI wallet tools (query_wallet_balance, wallet_transfer, contract tools) |
| 1.0.0 | 2026-03-09 | Initial wallet implementation with /wallet commands |

---

## Contact

For issues or questions about wallet implementation, refer to:
- workspace/skills/dc-hotwallet/SKILL.md - Usage instructions for AI
- This document - Implementation and sync guide
