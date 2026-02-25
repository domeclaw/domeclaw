# Webhook and Chat API Guide

DomeClaw now supports receiving messages via HTTP webhook and chat API endpoints.

## Table of Contents
- [Gateway HTTP API](#gateway-http-api)
- [Webhook Channel](#webhook-channel)
- [Examples](#examples)

## Gateway HTTP API

The gateway server (port `18790` by default) now provides two main endpoints:

### `/chat` - Direct Chat Endpoint

Send a message to the AI agent and receive an immediate response.

**Request:**
```bash
curl -X POST http://localhost:18790/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "สวัสดีครับ มีอะไรให้ช่วยไหม?",
    "chat_id": "user123"
  }'
```

**Response:**
```json
{
  "response": "สวัสดีครับ! มีอะไรให้ผมช่วยไหมครับ?"
}
```

**Parameters:**
- `message` (required): The message to send to the AI
- `chat_id` (optional): User identifier for session management. Defaults to `curl_user`

**How it works:**
1. The request is processed by the agent loop
2. The response is generated using the configured AI model
3. The conversation history is maintained per `chat_id`
4. Response is returned in the same HTTP request

### `/webhook` - Webhook Endpoint

Send a message to the AI agent via the message bus. This is useful for asynchronous processing.

**Request:**
```bash
curl -X POST http://localhost:18790/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "message": "มีปัญหาเรื่องระบบ",
    "sender_id": "support_ticket_456",
    "chat_id": "support_channel"
  }'
```

**Response:**
```json
{
  "status": "message queued"
}
```

**Parameters:**
- `message` (required): The message to send
- `sender_id` (optional): Identifier of the sender. Defaults to `webhook`
- `chat_id` (optional): Chat identifier. Defaults to `webhook_chat`
- `metadata` (optional): Additional metadata as key-value pairs

**How it works:**
1. The message is published to the message bus
2. The agent processes it in the background
3. The response is sent via the configured channels (if any)

## Webhook Channel

The Webhook Channel is a dedicated channel that listens on a separate port (default: `18795`) for incoming webhook requests.

### Configuration

Add the following to your `config.json`:

```json
{
  "channels": {
    "webhook": {
      "enabled": true,
      "token": "your-secret-token-here",
      "host": "0.0.0.0",
      "port": 18795
    }
  }
}
```

**Configuration Options:**
- `enabled`: Enable or disable the webhook channel
- `token`: Authentication token (required if enabled)
- `host`: Host to bind to (default: `localhost`)
- `port`: Port to listen on (default: `18795`)

### Usage

Once configured, you can send messages to the webhook channel:

**With Token in Header:**
```bash
curl -X POST http://localhost:18795/webhook \
  -H "Authorization: Bearer your-secret-token-here" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "สวัสดีครับ",
    "sender_id": "web_user_1",
    "chat_id": "chat_123"
  }'
```

**With Token in Query String:**
```bash
curl -X POST "http://localhost:18795/webhook?token=your-secret-token-here" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "สวัสดีครับ"
  }'
```

**Response:**
```json
{
  "status": "message received"
}
```

### Sending Responses to Other Channels

The webhook channel can forward responses to other channels (e.g., Telegram, Discord) using the AI's built-in tools.

**Example: Using the `message` tool**

```bash
curl -X POST http://localhost:18795/webhook \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "ส่งข้อความแจ้งเตือนไปยังกลุ่ม Telegram",
    "sender_id": "system",
    "chat_id": "monitoring"
  }'
```

The AI can use the `message` tool to send the response to Telegram:

```json
{
  "tool": "message",
  "args": {
    "content": "ระบบทำงานปกติครับ",
    "channel": "telegram",
    "chat_id": "123456789"
  }
}
```

## Examples

### Example 1: Basic Chat Interaction

```bash
# First message
curl -X POST http://localhost:18790/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "สวัสดีครับ ผมชื่อจอห์น", "chat_id": "john"}'

# Response: "สวัสดีครับจอห์น! มีอะไรให้ผมช่วยไหมครับ?"

# Follow-up message (same chat_id maintains context)
curl -X POST http://localhost:18790/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "วันนี้อากาศเป็นอย่างไรบ้าง?", "chat_id": "john"}'

# Response: "ขอโทษครับ ผมไม่สามารถตรวจสอบสภาพอากาศได้โดยตรง..."
```

### Example 2: Sending Alert to Telegram

First, enable Telegram in your config:
```json
{
  "channels": {
    "telegram": {
      "enabled": true,
      "token": "YOUR_TELEGRAM_BOT_TOKEN"
    }
  }
}
```

Then send a webhook request:
```bash
curl -X POST http://localhost:18790/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "message": "ระบบมีปัญหา! กรุณาแจ้งทีมเทคนิคไปยังกลุ่ม Telegram รหัส 123456789 ว่าเซิร์ฟเวอร์ล่ม",
    "sender_id": "monitoring_system",
    "chat_id": "alert_channel"
  }'
```

The AI will use the `message` tool to send to Telegram:
```json
{
  "status": "message queued"
}
```

### Example 3: Integration with External Services

You can use DomeClaw as a webhook receiver for services like GitHub, Slack, etc.

**GitHub Webhook Example:**
```bash
# GitHub sends a webhook to your server
# Your server forwards it to DomeClaw
curl -X POST http://localhost:18795/webhook \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "GitHub webhook received: New commit pushed to main branch by user@example.com",
    "sender_id": "github",
    "chat_id": "dev_team"
  }'
```

The AI can then:
1. Analyze the commit message
2. Send notifications to the team via Telegram/Slack
3. Create a summary or take other actions

## Security

### Authentication

1. **Gateway API (`/chat`, `/webhook` on port 18790)**:
   - Currently does not require authentication
   - Should be protected by firewall or reverse proxy

2. **Webhook Channel (port 18795)**:
   - Requires token authentication
   - Token can be sent via:
     - `Authorization: Bearer <token>` header
     - `?token=<token>` query parameter

### Recommendations

- Use HTTPS in production
- Set up a reverse proxy (Nginx, Caddy) with authentication
- Use strong, random tokens
- Restrict access to trusted IPs
- Enable firewall rules

## Troubleshooting

### "Message is required" error

Make sure your JSON payload includes the `message` field:
```json
{
  "message": "Your message here"
}
```

### "Unauthorized" error (Webhook Channel)

Check that:
1. Your token matches the one in `config.json`
2. You're sending it correctly (header or query string)
3. The channel is enabled

### No response from `/chat` endpoint

Check:
1. DomeClaw gateway is running
2. The configured AI model is working
3. Logs for any errors

### Messages not appearing in Telegram/other channels

Verify:
1. The channel is enabled in config
2. Token/API keys are correct
3. The AI has permission to use the `message` tool
4. The target chat_id exists and is accessible

## Advanced Features

### Using Metadata

You can attach metadata to messages for later processing:

```bash
curl -X POST http://localhost:18795/webhook \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "แจ้งเตือนระบบ",
    "metadata": {
      "priority": "high",
      "source": "monitoring",
      "timestamp": "2026-02-25T10:00:00Z"
    }
  }'
```

### Targeting Specific Channels

The AI can be instructed to send responses to specific channels:

```
ส่งข้อความ "ระบบพร้อมใช้งาน" ไปยังช่องทาง Telegram ห้องแชท ID 123456789
```

This will trigger the `message` tool with:
```json
{
  "channel": "telegram",
  "chat_id": "123456789",
  "content": "ระบบพร้อมใช้งาน"
}
```
