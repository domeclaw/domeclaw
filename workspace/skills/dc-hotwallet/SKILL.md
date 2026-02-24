---
name: dc-hotwallet
description: "Telegram Wallet Commands for DomeClaw - EVM Wallet & Blockchain Integration on ClawSwift chain."
homepage: https://github.com/domeclaw/domeclaw
metadata: {"nanobot":{"emoji":"🦐","requires":{"tools":["wallet_auto_transfer","query_wallet_balance","query_contract_call","execute_contract_write"]}},"category":"blockchain","platform":"telegram","chain":"clawswift","chain_id":7441}
---

# Wallet Skill (Hotwallet Mode)

EVM wallet operations via Telegram Bot on ClawSwift chain (Chain ID: 7441).

**⚠️ Hotwallet Mode**: PIN stored in plaintext. AI can execute transactions automatically. Convenience over security - use for testnet only.

## Quick Start

```
# 1. Create wallet
/wallet create 1234

# 2. Check balance
/wallet balance

# 3. Transfer tokens
/wallet transfer 0xRecipient 100 1234

# 4. Check specific token
/wallet balance 0x20c0000000000000000000000000000000000000
```

## AI-Powered Operations

Ask naturally in any language. AI will execute via tools:

| You say | AI does |
|---------|---------|
| "เรามี balance เท่าไหร่" | Queries balance directly |
| "โอน 0.01 CLAW ให้ 0xABC..." | Executes transfer automatically |
| "เช็ค balanceOf ของ 0xABC..." | Calls contract function |
| "approve ให้ 0xDEF..." | Writes to contract |

## Telegram Commands

### Wallet Management
```
/wallet create [PIN]     # Create new wallet
/wallet info             # View address & balance
/wallet unlock [PIN]     # Unlock for transactions
/wallet lock             # Lock wallet
```

### Token Operations
```
/wallet balance [token]              # Check balance (default: CLAW)
/wallet transfer <to> <amt> <pin>    # Send CLAW tokens
/wallet transfertoken <token> <to> <amt> <pin>  # Send ERC20
```

### Smart Contracts
```
/wallet abilist                       # List uploaded ABIs
/wallet abiupload <name>              # Upload ABI (reply to JSON)
/wallet call <contract> <abi> <method> [args]   # Read contract
/wallet write <c> <abi> <m> <val> <pin> [args]  # Write contract
```

## Examples

### Transfer CLAW
```
/wallet transfer 0xA3570FCDA303F55e0978be450f87F885d80a3758 0.01 1234
```

### Call Contract (Read)
```
/wallet call 0x20c0000000000000000000000000000000000000 erc20 balanceOf 0x44c2db1fc0986ca3c173403701c909874badc0d0
```

### Write Contract
```
/wallet write 0x20c0000000000000000000000000000000000000 erc20 transfer 0 1234 0xRecipientAddress 1000000000000000000
```

## Network

- **Chain**: ClawSwift
- **Chain ID**: 7441
- **RPC**: https://exp.clawswift.net/rpc
- **Explorer**: https://exp.clawswift.net
- **Gas Token**: 0x20c0000000000000000000000000000000000000 (CLAW)

## Security Warning

This is a **hotwallet** implementation:
- PIN stored at `~/.domeclaw/workspace/wallet/pin.json` (plaintext)
- AI has direct keystore access
- Suitable for **testnet/development only**
- **NOT for mainnet or real funds**
