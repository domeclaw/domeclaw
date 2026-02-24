---
name: wallet
version: 1.0.0
description: Telegram Wallet Commands for DomeClaw - EVM Wallet & Blockchain Integration
homepage: https://github.com/domeclaw/domeclaw
metadata: {"category":"blockchain","platform":"telegram","chain":"clawswift","chain_id":7441,"tags":["wallet","telegram","evm","blockchain","erc20","smart-contract"]}
---

# Telegram Wallet Skill for DomeClaw

## Overview

Wallet management and blockchain interaction commands for DomeClaw via Telegram Bot. This skill provides comprehensive EVM wallet operations including token transfers, balance checking, and smart contract interactions directly from Telegram.

**Tagline**: "Your Web3 Wallet in Telegram"

---

## Network Configuration

### Supported Chains

| Property | Value |
|----------|-------|
| **Chain Name** | ClawSwift |
| **Chain ID** | `7441` |
| **Native Currency** | CLAW |
| **Decimals** | 18 |
| **RPC URL** | `https://exp.clawswift.net/rpc` |
| **Explorer** | `https://exp.clawswift.net` |
| **Gas Token** | `0x20c0000000000000000000000000000000000000` |

---

## Wallet Commands

### 1. Wallet Management

#### Create New Wallet
```
/wallet create [PIN]
```
Creates a new Ethereum wallet with optional 4-digit PIN.

**Example:**
```
/wallet create 1234
```

**Output:**
- Wallet address (e.g., `0x44c2db1fc0986ca3c173403701c909874badc0d0`)
- Success confirmation

---

#### View Wallet Info
```
/wallet info
```
Displays wallet address, balance, and chain information.

**Example:**
```
/wallet info
```

**Output:**
```
🦐 DomeClaw Wallet

📍 Address: 0x44c2...
💰 Balance: 1000 CLAW
🔗 Chain: ClawSwift (7441)

🔒 Status: Locked
```

---

#### Unlock Wallet
```
/wallet unlock [PIN]
```
Unlocks wallet for transactions requiring signing.

**Example:**
```
/wallet unlock 1234
```

---

#### Lock Wallet
```
/wallet lock
```
Locks the wallet for security.

**Example:**
```
/wallet lock
```

---

### 2. Token Operations

#### Check Balance
```
/wallet balance [token_address]
```
Checks token balance. Defaults to CLAW token if no address provided.

**Examples:**
```
# Check CLAW balance (default)
/wallet balance

# Check specific token balance
/wallet balance 0x20c0000000000000000000000000000000000000
```

**Output:**
```
💰 Token Balance

👛 Wallet: 0x44c2...
🪙 Token: 0x20c0...0000
🏷️ Symbol: CLAW
📊 Decimals: 18

💵 Balance: 1234.5678 CLAW
```

---

#### Transfer CLAW Tokens (Default)
```
/wallet transfer <to_address> <amount> <pin>
```
Transfers CLAW tokens (native/gas token) to another address.

**Parameters:**
- `to_address` - Recipient address (0x...)
- `amount` - Amount to send (e.g., 100, 0.5)
- `pin` - Wallet PIN (4 digits)

**Example:**
```
/wallet transfer 0xA3570FCDA303F55e0978be450f87F885d80a3758 100 1234
```

**Output:**
```
✅ Transfer Successful!

📤 Transaction Hash:
0xabc123...
```

---

#### Transfer ERC20 Tokens
```
/wallet transfertoken <token_address> <to_address> <amount> <pin>
```
Transfers any ERC20 token.

**Parameters:**
- `token_address` - Token contract address
- `to_address` - Recipient address
- `amount` - Amount to send
- `pin` - Wallet PIN

**Example:**
```
/wallet transfertoken 0x20c000000000000000000000550a7f768b9a78f3 0xA3570FCDA303F55e0978be450f87F885d80a3758 0.01 1234
```

---

### 3. Smart Contract Operations

#### List Uploaded ABIs
```
/wallet abilist
```
Shows all uploaded contract ABIs.

**Example:**
```
/wallet abilist
```

**Output:**
```
📋 Available ABIs:

1. erc20
2. uniswap-v2
3. my-custom-contract
```

---

#### Upload ABI
```
/wallet abiupload <name>
```
Uploads an ABI JSON file for contract interaction.

**How to use:**
1. Send a JSON file containing the ABI in Telegram
2. Reply to that message with: `/wallet abiupload <name>`

**Example:**
```
[Reply to JSON file]
/wallet abiupload erc20
```

**Sample ABI JSON:**
```json
[
  {
    "constant": true,
    "inputs": [{"name": "_owner", "type": "address"}],
    "name": "balanceOf",
    "outputs": [{"name": "balance", "type": "uint256"}],
    "type": "function"
  },
  {
    "constant": false,
    "inputs": [
      {"name": "_to", "type": "address"},
      {"name": "_value", "type": "uint256"}
    ],
    "name": "transfer",
    "outputs": [{"name": "", "type": "bool"}],
    "type": "function"
  }
]
```

---

#### Call Contract (Read)
```
/wallet call <contract_address> <abi_name> <method> [args...]
```
Calls a read-only function on a smart contract.

**Parameters:**
- `contract_address` - Contract address (0x...)
- `abi_name` - Name of uploaded ABI
- `method` - Method name to call
- `args` - Optional arguments

**Examples:**
```
# Check token balance
/wallet call 0x20c0000000000000000000000000000000000000 erc20 balanceOf 0x44c2db1fc0986ca3c173403701c909874badc0d0

# Get total supply
/wallet call 0x20c0000000000000000000000000000000000000 erc20 totalSupply

# Get token symbol
/wallet call 0x20c0000000000000000000000000000000000000 erc20 symbol
```

**Output:**
```
📤 Contract Call Result

Contract: 0x20c0...0000
Method: balanceOf

Result: 1000000000000000000000
```

---

#### Write Contract (State-Changing)
```
/wallet write <contract> <abi> <method> <value> <pin> [args...]
```
Executes a state-changing transaction on a smart contract.

**Parameters:**
- `contract` - Contract address
- `abi` - ABI name
- `method` - Method name
- `value` - ETH value to send (0 for token transfers)
- `pin` - Wallet PIN
- `args` - Method arguments

**Examples:**
```
# Transfer tokens via contract
/wallet write 0x20c0000000000000000000000000000000000000 erc20 transfer 0 1234 0xA3570FCDA303F55e0978be450f87F885d80a3758 1000000000000000000

# Approve spender
/wallet write 0x20c0000000000000000000000000000000000000 erc20 approve 0 1234 0xSpenderAddress 1000000000000000000000
```

**Output:**
```
✅ Transaction Sent!

📤 Transaction Hash:
0xabc123...
```

---

## Complete Command Reference

| Command | Syntax | Description |
|---------|--------|-------------|
| `/wallet create` | `[PIN]` | Create new wallet |
| `/wallet info` | - | View wallet info |
| `/wallet unlock` | `[PIN]` | Unlock wallet |
| `/wallet lock` | - | Lock wallet |
| `/wallet balance` | `[token]` | Check token balance |
| `/wallet transfer` | `<to> <amt> <pin>` | Send CLAW tokens |
| `/wallet transfertoken` | `<token> <to> <amt> <pin>` | Send ERC20 tokens |
| `/wallet abilist` | - | List ABIs |
| `/wallet abiupload` | `<name>` | Upload ABI (reply to JSON) |
| `/wallet call` | `<c> <abi> <method> [args]` | Read contract |
| `/wallet write` | `<c> <abi> <m> <val> <pin> [args]` | Write to contract |

---

## Common Workflows

### Workflow 1: Complete Setup
```
# 1. Create wallet
/wallet create 1234

# 2. Check address
/wallet info

# 3. Get CLAW from faucet
# Visit: https://exp.clawswift.net/faucet

# 4. Check balance
/wallet balance
```

### Workflow 2: Transfer Tokens
```
# Transfer CLAW tokens
/wallet transfer 0xRecipientAddress 100 1234

# Transfer specific ERC20 token
/wallet transfertoken 0xTokenAddress 0xRecipientAddress 50 1234
```

### Workflow 3: Smart Contract Interaction
```
# 1. Upload ABI first
[Send erc20.json file]
/wallet abiupload erc20

# 2. Check balance via contract
/wallet call 0xTokenAddress erc20 balanceOf 0xYourAddress

# 3. Transfer via contract
/wallet write 0xTokenAddress erc20 transfer 0 1234 0xRecipient 1000000000000000000
```

---

## Security Notes

### PIN Protection
- PIN is required for all transaction operations
- PIN is stored encrypted in wallet directory
- Default PIN: `1234` (change recommended)

### Wallet Storage
- Keystore location: `~/.domeclaw/workspace/wallet/`
- Private keys encrypted with scrypt
- Never share keystore files or private keys

### Transaction Safety
- Always verify recipient address before sending
- Check transaction details in explorer: `https://exp.clawswift.net/tx/<hash>`
- Start with small amounts when testing

---

## Troubleshooting

### "Invalid PIN"
- Ensure PIN is correct (default: 1234)
- Check wallet was created successfully: `/wallet info`

### "Failed to estimate gas"
- Check token contract address is correct
- Ensure you have enough CLAW for gas fees
- Verify amount format (use decimals correctly)

### "Execution reverted"
- Insufficient token balance
- Contract paused or restricted
- Invalid method parameters

### "ABI not found"
- Upload ABI first with `/wallet abiupload`
- Check ABI name matches exactly

---

## Resources

- **Explorer**: https://exp.clawswift.net
- **Faucet**: https://exp.clawswift.net/faucet
- **GitHub**: https://github.com/domeclaw/domeclaw
- **Documentation**: See README.md in project root

---

## Version Info

| Property | Value |
|----------|-------|
| **Skill Version** | 1.0.0 |
| **Compatible With** | DomeClaw v1.0+ |
| **Last Updated** | 2026-02-24 |
| **Platform** | Telegram Bot |

---

## Quick Reference

```
# Essentials
/wallet create 1234          # Create wallet
/wallet info                 # View info
/wallet balance              # Check CLAW balance
/wallet transfer 0x... 10 1234   # Send CLAW

# Smart Contracts
/wallet abiupload erc20      # Upload ABI
/wallet call 0x... erc20 balanceOf 0x...   # Read
/wallet write 0x... erc20 transfer 0 1234 0x... 100   # Write
```

---

*This skill enables Web3 wallet operations directly within Telegram chat. All transactions are signed locally and broadcast to the ClawSwift blockchain.*
