# DomeClaw

🦐 DomeClaw is an ultra-lightweight personal AI Assistant. It is the official continuation and rebrand of [PicoClaw](https://github.com/sipeed/picoclaw), which was inspired by [nanobot](https://github.com/HKUDS/nanobot). DomeClaw is refactored from the ground up in Go through a self-bootstrapping process, where the AI agent itself drove the entire architectural migration and code optimization.

⚡️ Runs on $10 hardware with <10MB RAM: That's 99% less memory than OpenClaw and 98% cheaper than a Mac mini!

**📦 Project Info:**
- **Source Code:** https://github.com/domeclaw/domeclaw
- **Issues:** https://github.com/domeclaw/domeclaw/issues
- **Discussions:** https://github.com/domeclaw/domeclaw/discussions

---

## 🆕 New Features: EVM Wallet & Blockchain Support

DomeClaw now supports **EVM-compatible blockchains** with built-in wallet management!

### **🔐 Wallet Features:**
- Create and manage Ethereum wallets
- Check token balances (native & ERC20)
- Transfer tokens
- Interact with smart contracts (read/write)
- Upload and manage contract ABIs
- PIN-protected wallet security

### **🔗 Multi-Chain Support:**
- Configure multiple EVM chains
- Auto-detect token decimals
- Support for both native and ERC20 tokens
- Built-in RPC client with failover

### **📱 Telegram Commands:**
```
/wallet create 1234          # Create wallet with PIN
/wallet info                 # View balance and wallet info
/wallet unlock 1234          # Unlock wallet for transactions
/wallet lock                 # Lock wallet
/wallet abi upload [name]    # Upload contract ABI
/wallet abi list             # List available ABIs
```

### **⛓️ Example Configuration (ClawSwift Chain):**

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
        "gas_token_name": "CLAW"
      }
    ]
  }
}
```

**Configuration Options:**
- `name`: Chain name (e.g., "ClawSwift")
- `chain_id`: EVM chain ID (e.g., 7441)
- `rpc`: RPC endpoint URL
- `explorer`: Block explorer URL
- `currency`: Token symbol
- `is_native`: `true` for native token, `false` for ERC20
- `gas_token`: ERC20 contract address (if `is_native: false`)
- `gas_token_name`: Token name/symbol

---

## 🔥 Hotwallet Mode (AI-Powered Wallet)

DomeClaw includes an **optional hotwallet mode** that allows the AI agent to execute blockchain transactions directly without manual user intervention.

### **Features:**
- **Natural Language Queries**: Ask "What's my balance?" instead of typing commands
- **AI-Executed Transfers**: Say "Send 0.01 CLAW to 0xABC..." and AI handles it
- **Smart Contract Interaction**: Query or write to contracts via conversation
- **Auto PIN Reading**: PIN stored in `~/.domeclaw/workspace/wallet/pin.json`

### **AI Wallet Tools:**
| Tool | Description | Example Query |
|------|-------------|---------------|
| `query_wallet_balance` | AI checks balance directly | "เรามี balance เท่าไหร่" |
| `wallet_auto_transfer` | AI executes transfers | "โอน 0.01 CLAW ให้ 0xABC..." |
| `query_contract_call` | AI reads contract data | "เช็ค balanceOf ที่อยู่นี้" |
| `execute_contract_write` | AI writes to contracts | "approve ให้ 0xDEF..." |

### **⚠️ Security Warning:**

> **HOTWALLET MODE PRIORITIZES CONVENIENCE OVER SECURITY**
> 
> - PIN is stored in **plaintext** at `~/.domeclaw/workspace/wallet/pin.json`
> - AI has direct access to wallet keystore and can sign transactions
> - Suitable for **testnet/development ONLY**
> - **NOT RECOMMENDED** for mainnet or wallets holding significant funds
> 
> Use at your own risk. For production use, disable wallet or use manual PIN entry only.

---
