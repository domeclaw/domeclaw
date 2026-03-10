---
name: dc-hotwallet
version: 1.1.0
description: "Telegram Wallet Commands for DomeClaw - EVM Wallet & Blockchain Integration on ClawSwift chain."
homepage: https://github.com/domeclaw/domeclaw
metadata: {"nanobot":{"emoji":"🦐","requires":{"tools":["wallet_auto_transfer","query_wallet_balance","query_contract_call","execute_contract_write"]},"natural_language":{"patterns":{"transfer":"Transfer {amount} {token} to {address}","balance":"Check {token} balance of {address}"},"mapping":{"transfer":{"tool":"wallet_auto_transfer","parameters":{"amount":"{amount}","token":"{token}","to":"{address}"},"pin_required":true,"cwd":"/Users/dome/project/domeclaw/picoclaw-domeclaw/picoclaw","command":"picoclaw --config /Users/dome/project/domeclaw/picoclaw-domeclaw/picoclaw/config-with-wallet.json wallet transfer"},"balance":{"tool":"query_wallet_balance","parameters":{"token":"{token}","address":"{address}"}}}}},"category":"blockchain","platform":"telegram","chain":"clawswift","chain_id":7441}
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
| "What's my balance" | Queries balance directly |
| "Transfer 0.01 CLAW to 0xABC..." | Executes transfer automatically |
| "Check balanceOf of 0xABC..." | Calls contract function |
| "Approve for 0xDEF..." | Writes to contract |

## Natural Language Commands

### Supported Patterns
- **Transfer Pattern**: "Transfer {amount} {token} to {address}"
  - {amount}: The amount of tokens to transfer (e.g., 0.1, 100)
  - {token}: The token name (e.g., CLAW)
  - {address}: The recipient address (e.g., 0xA3570FCDA303F55e0978be450f87F885d80a3758)

### Parameter Mapping
- Natural language commands will be converted to parameters for the wallet_auto_transfer tool
- Mapping: amount → amount, token → token, address → to

### Examples
| Natural Language Command | Equivalent CLI Command | Tool Parameters |
|---------------------------|-----------------------|----------------------|
| Transfer 0.1 CLAW to 0xA3570FCDA303F55e0978be450f87F885d80a3758 | /wallet transfer 0xA3570FCDA303F55e0978be450f87F885d80a3758 0.1 1234 | to=0xA3570FCDA303F55e0978be450f87F885d80a3758, amount=0.1, token=CLAW |
| Check CLAW balance of 0xA3570FCDA303F55e0978be450f87F885d80a3758 | /wallet balance 0xA3570FCDA303F55e0978be450f87F885d80a3758 | address=0xA3570FCDA303F55e0978be450f87F885d80a3758, token=CLAW |
| Transfer 100 CLAW to 0x44c2db1fc0986ca3c173403701c909874badc0d0 | /wallet transfer 0x44c2db1fc0986ca3c173403701c909874badc0d0 100 1234 | to=0x44c2db1fc0986ca3c173403701c909874badc0d0, amount=100, token=CLAW |

### Parsing Rules
- **Amount Extraction**: Use regex `\d+\.?\d*` to extract the token amount (e.g., 0.1, 100)
- **Token Extraction**: Use regex `[A-Za-z]+` to extract the token name (e.g., CLAW)
- **Address Extraction**: Use regex `0x[0-9a-fA-F]{40}` to extract the recipient address
- **PIN Handling**: Will prompt the user to enter their PIN after successful parameter extraction
- **Fallback Logic**: If any parameter is missing, will prompt the user to provide additional details

### Regex Patterns
```regex
# Transfer Sentence Pattern Extraction
^Transfer\s+(\d+\.?\d*)\s+([A-Za-z]+)\s+to\s+(0x[0-9a-fA-F]{40})$```

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



## Security Warning

This is a **hotwallet** implementation:
- PIN stored at `workspace/wallet/pin.json` (plaintext)
- AI has direct keystore access
