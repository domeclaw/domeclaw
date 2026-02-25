# Webhook Channel

The webhook channel allows you to send messages to DomeClaw via HTTP POST requests.

## Configuration

Add to your config:

```json
{
  "channels": {
    "webhook": {
      "enabled": true,
      "host": "0.0.0.0",
      "port": 18794,
      "path": "/webhook",
      "auth_token": "your-secret-token"
    }
  }
}
```

Or using environment variables:

```bash
DOMECLAW_CHANNELS_WEBHOOK_ENABLED=true
DOMECLAW_CHANNELS_WEBHOOK_HOST=0.0.0.0
DOMECLAW_CHANNELS_WEBHOOK_PORT=18794
DOMECLAW_CHANNELS_WEBHOOK_PATH=/webhook
DOMECLAW_CHANNELS_WEBHOOK_AUTH_TOKEN=your-secret-token
```

## Usage

### Send a message

```bash
curl -X POST http://localhost:18794/webhook \
  -H "Authorization: Bearer your-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Hello, how are you?",
    "chat_id": "123456789",
    "metadata": {
      "target_channel": "telegram"
    }
  }'
```

### Response

```json
{
  "status": "success",
  "message_id": "webhook-12345"
}
```

### Error Responses

**Unauthorized (401):**
```json
{
  "status": "error",
  "error": "Invalid or missing authorization token"
}
```

**Bad Request (400):**
```json
{
  "status": "error",
  "error": "Message is required"
}
```

## Metadata Fields

| Field | Required | Description |
|-------|----------|-------------|
| `target_channel` | Optional | Target channel to send response (telegram, discord, etc.) |
| Any custom fields | No | Additional metadata for routing |

## Supported Channels

The following channels can be used as `target_channel`:

- `telegram` (default)
- `discord`
- `whatsapp`
- `feishu`
- `dingtalk`
- `slack`
- `line`
- `qq`
- `wecom`
- `wecom_app`
- `onebot`
- `maixcam`

## Full Example with Telegram

```bash
curl -X POST http://localhost:18794/webhook \
  -H "Authorization: Bearer test_token" \
  -H "Content-Type: application/json" \
  -d '{
    "message": ">balance เหลือเท่าไหร่",
    "chat_id": "188576201",
    "metadata": {
      "target_channel": "telegram"
    }
  }'
```

This will:
1. Receive the webhook request
2. Process it through the LLM
3. Send the response back to Telegram chat `188576201`
