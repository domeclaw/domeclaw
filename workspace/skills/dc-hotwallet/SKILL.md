---
name: dc-hotwallet
version: 2.0.0
description: "Telegram Wallet Tools for DomeClaw - AI-powered EVM Wallet & Blockchain Integration on ClawSwift chain."
homepage: https://github.com/domeclaw/domeclaw
metadata: {"nanobot":{"emoji":"🦐","requires":{"tools":["query_wallet_balance","wallet_transfer","query_contract_call","execute_contract_write"]}},"category":"blockchain","platform":"telegram","chain":"clawswift","chain_id":7441}
---

# Wallet Skill (Hotwallet Mode)

EVM wallet operations via AI Tools on ClawSwift chain (Chain ID: 7441).

**⚠️ Hotwallet Mode**: PIN stored in plaintext at `workspace/wallets/pin.json`. AI can execute transactions automatically. Convenience over security - use for testnet only.

## AI Tools Available

When users ask about wallet operations in natural language, you MUST use the appropriate tool:

### 1. query_wallet_balance
**Use when:** User asks about their wallet balance, coins, or tokens.

**Examples:**
- "wallet มีเหรียญอะไรบ้าง" → Use `query_wallet_balance`
- "เช็คยอดเงิน" → Use `query_wallet_balance`
- "balance เท่าไหร่" → Use `query_wallet_balance`
- "มีกี่เหรียญ" → Use `query_wallet_balance`
- "ดู wallet" → Use `query_wallet_balance`
- "check balance" → Use `query_wallet_balance`
- "มีเงินเท่าไหร่" → Use `query_wallet_balance`
- "เหลือกี่บาท" → Use `query_wallet_balance`

**Returns:** Wallet address, balance, symbol, and explorer link.

---

### 2. wallet_transfer
**Use when:** User wants to send/transfer tokens.

**Examples:**
- "ส่ง 0.01 CLAW ให้ 0xABC..." → Use `wallet_transfer`
- "โอน 100 tokens ให้เพื่อน" → Use `wallet_transfer`
- "transfer 0.5 CLAW to 0x..." → Use `wallet_transfer`
- "ส่งเงินให้ 0x..." → Use `wallet_transfer`

**Parameters:**
- `to_address`: Recipient address (0x...)
- `amount`: Amount to transfer (e.g., "0.01", "100")
- `token_address`: (Optional) ERC20 token address. If not provided, sends native token (CLAW).

**Returns:** Transaction hash with explorer link.

---

### 3. query_contract_call
**Use when:** User wants to read data from a smart contract.

**Examples:**
- "เช็ค balanceOf ของ 0x... บนสัญญา 0x..." → Use `query_contract_call`
- "ดู totalSupply" → Use `query_contract_call`
- "call balanceOf on token contract" → Use `query_contract_call`
- "เช็ค allowance" → Use `query_contract_call`

**Parameters:**
- `contract_address`: Smart contract address
- `abi_type`: ABI name (e.g., "erc20", "erc721")
- `method`: Method name (e.g., "balanceOf", "totalSupply")
- `params`: (Optional) Array of arguments

---

### 4. execute_contract_write
**Use when:** User wants to write/execute a function on a smart contract (requires PIN).

**Examples:**
- "approve ให้ 0x... ใช้ 100 tokens" → Use `execute_contract_write`
- "transfer tokens บนสัญญา" → Use `execute_contract_write`
- "write to contract" → Use `execute_contract_write`

**Parameters:**
- `contract_address`: Smart contract address
- `abi_type`: ABI name
- `method`: Method name (e.g., "transfer", "approve")
- `value`: ETH value to send (use "0" for token transfers)
- `params`: Array of arguments

---

## Quick Reference

| User Intent | Tool to Use |
|-------------|-------------|
| Check balance | `query_wallet_balance` |
| Transfer/Send tokens | `wallet_transfer` |
| Read contract data | `query_contract_call` |
| Write to contract | `execute_contract_write` |

## Network Configuration

- **Chain**: ClawSwift
- **Chain ID**: 7441
- **RPC**: https://exp.clawswift.net/rpc
- **Explorer**: https://exp.clawswift.net
- **Gas Token**: 0x20c0000000000000000000000000000000000000 (CLAW)
- **Decimals**: 16

## Security Warning

This is a **hotwallet** implementation:
- PIN stored at `workspace/wallets/pin.json` (plaintext)
- AI has direct keystore access and can sign transactions automatically
- Use for testnet only
- Never store large amounts in this wallet

## Examples

### Example 1: Check Balance
**User:** "wallet มีเหรียญอะไรบ้าง"
**AI Action:** Call `query_wallet_balance` tool
**Result:** Shows wallet address, CLAW balance, and explorer link

### Example 2: Transfer Tokens
**User:** "ส่ง 0.01 CLAW ให้ 0xA3570FCDA303F55e0978be450f87F885d80a3758"
**AI Action:** Call `wallet_transfer` with to_address and amount
**Result:** Transaction submitted with tx hash and explorer link

### Example 3: Query Contract
**User:** "เช็ค balanceOf ของ 0xABC บนสัญญา 0x20c0000000000000000000000000000000000000"
**AI Action:** Call `query_contract_call` with contract address and method
**Result:** Returns balance from contract

### Example 4: Execute Contract Write
**User:** "approve ให้ 0xDEF ใช้ 1000 tokens"
**AI Action:** Call `execute_contract_write` for approve function
**Result:** Transaction submitted
