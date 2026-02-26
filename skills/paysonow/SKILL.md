---
name: paysonow
description: "Payment gateway webhook handler for DomeClaw - Automatically process payments and transfer CLAW tokens to customers."
homepage: https://github.com/domeclaw/domeclaw
version: "1.0.0"
metadata: {"nanobot":{"emoji":"💳","requires":{"tools":["wallet_auto_transfer","query_wallet_balance"]}},"category":"payment","platform":"webhook","chain":"clawswift","chain_id":7441}
---

# PaySoNow Payment Gateway

Automated payment processing for merchants using DomeClaw as payment receiver. When a payment notification arrives via webhook, DomeClaw automatically transfers CLAW tokens to the customer's wallet.

## Overview

PaySoNow enables merchants to receive and process payments automatically. When an external payment gateway sends a webhook notification, DomeClaw parses the payment details and automatically transfers CLAW tokens to the customer's wallet address.

## Features

- Receive payment notifications via webhook
- Parse payment details from message text
- Automatic CLAW token transfers to customer wallets
- Support for EVM-compatible addresses
- Secure token transfers with PIN verification
- Transaction logging and audit trail

## Configuration

This skill requires:
1. Webhook channel enabled with proper authentication token
2. Wallet service with sufficient CLAW token balance
3. Proper PIN configuration for auto-transfers
4. Hotwallet mode enabled (for automatic transfers)

## Webhook Setup

Configure your payment gateway to send notifications to your DomeClaw instance:

```json
POST /webhook
Headers:
  Authorization: Bearer YOUR_TOKEN
  Content-Type: application/json
Body:
{
  "message": ">ชำระเงินจาก 0x5266Dfa5ae013674f8FdC832b7c601B838D94eE6 จำนวน 10 CLAW",
  "chat_id": "-5129639667",
  "metadata": {
    "target_channel": "telegram"
  }
}
```

## Message Format

The skill expects payment messages in the format:
```
>ชำระเงินจาก [WALLET_ADDRESS] จำนวน [AMOUNT] [TOKEN]
```

For example:
- `>ชำระเงินจาก 0x5266Dfa5ae013674f8FdC832b7c601B838D94eE6 จำนวน 10 CLAW`
- `>ชำระเงินจาก 0x123...abc จำนวน 5.5 CLAW`

## How It Works

1. Payment gateway sends webhook notification to DomeClaw
2. DomeClaw parses the message to extract wallet address and amount
3. DomeClaw automatically transfers the specified amount to the customer's wallet
4. Transaction is logged for audit purposes

## Security

- Webhook authentication via bearer token
- Transaction amounts limited by configuration
- Wallet PIN protection for transfers
- Audit logging for all transactions
- Hotwallet mode required (suitable for testnet only)

## Supported Tokens

- **CLAW**: 0x20c0000000000000000000000000000000000000 (Native Gas Token)
- **MTK**: 0x20C000000000000000000000550a7F768B9A78f3

## Network

- **Chain**: ClawSwift
- **Chain ID**: 7441
- **RPC**: https://exp.clawswift.net/rpc
- **Explorer**: https://exp.clawswift.net

## Example Flow

1. Customer pays via external payment gateway
2. Gateway sends webhook: 
   ```json
   {
     "message": ">ชำระเงินจาก 0x5266Dfa5ae013674f8FdC832b7c601B838D94eE6 จำนวน 10 CLAW",
     "chat_id": "-5129639667",
     "metadata": {"target_channel": "telegram"}
   }
   ```
3. DomeClaw automatically transfers 10 CLAW to 0x5266...4eE6
4. Confirmation sent to chat_id -5129639667