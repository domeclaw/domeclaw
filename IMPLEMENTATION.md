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

### Key Components

```
pkg/tools/wallet_query.go      - query_wallet_balance tool
pkg/tools/wallet_transfer.go   - wallet_transfer tool
pkg/tools/wallet.go            - Contract tools (query_contract_call, execute_contract_write)
pkg/wallet/                    - Wallet service layer
pkg/commands/cmd_wallet.go     - /wallet slash commands
pkg/agent/loop.go              - Tool registration (lines 236-242)
```

## Files Modified from Upstream

### Core Files (Risk of Conflict on Sync)
```
pkg/agent/loop.go              - Added wallet tools registration
pkg/commands/builtin.go        - Added walletCommand()
pkg/config/config.go           - Added Wallet config section
pkg/config/defaults.go         - Added wallet defaults
cmd/picoclaw/main.go           - Removed CLI wallet command
cmd/picoclaw/main_test.go      - Updated tests
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
```

## Sync Strategy from Upstream

### Preparation
```bash
# 1. Create backup branch
git checkout domeclaw
git branch domeclaw-backup-$(date +%Y%m%d)

# 2. Fetch latest upstream
git fetch upstream main
```

### Sync Methods

#### Method 1: Rebase (Recommended)
```bash
git checkout domeclaw
git rebase upstream/main

# If conflicts occur, resolve them:
git add <conflicted-files>
git rebase --continue
```

#### Method 2: Merge
```bash
git checkout domeclaw
git merge upstream/main

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

#### 2. `pkg/commands/builtin.go`
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

#### 3. `pkg/config/config.go`
**Our additions:**
```go
type Config struct {
    // ... other fields ...
    Wallet WalletConfig `json:"wallet"`  // <-- Keep this
}
```

**Resolution:** Keep the Wallet field and related config types.

#### 4. `cmd/picoclaw/main.go`
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
- [ ] Config loads: Verify `pkg/config/config.go` has Wallet field
- [ ] Commands available: Check `/wallet` works in Telegram
- [ ] Tools work: Test natural language queries

## Testing Wallet Features

### Natural Language Queries
```
"wallet มีเหรียญอะไรบ้าง"     -> query_wallet_balance
"เช็คยอดเงิน"                   -> query_wallet_balance
"ส่ง 0.01 CLAW ให้ 0x..."      -> wallet_transfer
"โอนเงินให้เพื่อน"               -> wallet_transfer
```

### Slash Commands
```
/wallet create [password]
/wallet info
/wallet transfer [to] [amount]
/wallet transfer_token [to] [amount]
/wallet chain
/wallet call [contract] [abi] [method] [args...]
/wallet write [contract] [abi] [method] [value] [args...]
```

## Troubleshooting

### Issue: Tools not appearing
**Check:** `pkg/agent/loop.go` has wallet tools registration
**Verify:** `cfg.Wallet.Enabled` is true in config.json

### Issue: Transfer fails with PIN error
**Check:** `workspace/wallets/pin.json` exists with correct format:
```json
{"password": "your-pin"}
```

### Issue: Cannot connect to blockchain
**Check:** `config.json` has correct Wallet.Chains configuration
**Verify:** RPC endpoint is accessible

## Configuration Example

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

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 2.0.0 | 2026-03-10 | Added AI wallet tools (query_wallet_balance, wallet_transfer, contract tools) |
| 1.0.0 | 2026-03-09 | Initial wallet implementation with /wallet commands |

---

## Contact

For issues or questions about wallet implementation, refer to:
- SKILL.md - Usage instructions for AI
- This document - Implementation and sync guide
