# Kimi Code Provider

DomeClaw supports Kimi Code (Moonshot's coding-focused LLM) via the new `kimi_code` provider.

## Quick Start

### 1. Get API Key

Visit [Kimi Code Platform](https://platform.kimi.com/) to get your API key.

### 2. Configure

Add to your `~/.domeclaw/config.json`:

**Option A: Using `model_list` (Recommended)**

```json
{
  "model_list": [
    {
      "model_name": "kimi-code",
      "model": "kimi-code/kimi-coding",
      "api_key": "sk-your-kimi-code-key",
      "api_base": "https://api.kimi.com/coding/v1"
    }
  ],
  "agents": {
    "defaults": {
      "model": "kimi-code"
    }
  }
}
```

**Option B: Using `providers` section (Legacy)**

```json
{
  "agents": {
    "defaults": {
      "provider": "kimi_code",
      "model": "kimi-coding"
    }
  },
  "providers": {
    "kimi_code": {
      "api_key": "sk-your-kimi-code-key",
      "api_base": "https://api.kimi.com/coding/v1"
    }
  }
}
```

### 3. Run

```bash
domeclaw agent -m "Hello, help me write a Python function"
```

## Configuration Options

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `api_key` | string | Yes | - | Your Kimi Code API key |
| `api_base` | string | No | `https://api.kimi.com/coding/v1` | API base URL |
| `proxy` | string | No | - | HTTP proxy URL (e.g., `http://localhost:7890`) |

## Model Names

The following model name formats are supported:

- `kimi-coding` - Direct model name
- `kimi-code/kimi-coding` - With provider prefix
- `kimi_code/kimi-coding` - Alternative provider prefix

## Environment Variables

You can also configure via environment variables:

```bash
export PICOCLAW_PROVIDERS_KIMI_CODE_API_KEY="sk-your-kimi-code-key"
export PICOCLAW_PROVIDERS_KIMI_CODE_API_BASE="https://api.kimi.com/coding/v1"
```

## Example: Full Configuration

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.domeclaw/workspace",
      "model": "kimi-code",
      "max_tokens": 8192,
      "temperature": 0.7,
      "max_tool_iterations": 20
    }
  },
  "model_list": [
    {
      "model_name": "kimi-code",
      "model": "kimi-code/kimi-coding",
      "api_key": "sk-your-kimi-code-key",
      "api_base": "https://api.kimi.com/coding/v1"
    },
    {
      "model_name": "gpt4-backup",
      "model": "openai/gpt-4o",
      "api_key": "sk-openai-key"
    }
  ],
  "tools": {
    "web": {
      "duckduckgo": {
        "enabled": true,
        "max_results": 5
      }
    }
  }
}
```

## Fallback Configuration

You can configure fallbacks for high availability:

```json
{
  "agents": {
    "defaults": {
      "model": {
        "primary": "kimi-code",
        "fallbacks": ["gpt4-backup", "deepseek"]
      }
    }
  }
}
```

## Troubleshooting

### API Key Issues

Make sure your API key is valid and has sufficient credits. Check your Kimi Code dashboard.

### Rate Limiting

If you encounter rate limits, consider:
1. Using the `cooldown` configuration
2. Setting up multiple API keys with load balancing
3. Implementing retry logic in your workflow

### Connection Issues

If you're behind a firewall, use the `proxy` configuration:

```json
{
  "model_list": [
    {
      "model_name": "kimi-code",
      "model": "kimi-code/kimi-coding",
      "api_key": "sk-your-kimi-code-key",
      "api_base": "https://api.kimi.com/coding/v1",
      "proxy": "http://localhost:7890"
    }
  ]
}
```

## Differences from Moonshot Provider

| Feature | Kimi Code Provider | Moonshot Provider |
|---------|-------------------|-------------------|
| API Base | `https://api.kimi.com/coding/v1` | `https://api.moonshot.cn/v1` |
| Focus | Coding tasks | General purpose |
| Models | `kimi-coding` | `moonshot-v1-*` |

## Related Links

- [Kimi Code Documentation](https://platform.kimi.com/docs)
- [Moonshot AI](https://www.moonshot.cn/)
- [DomeClaw Provider Configuration](../../README.md#providers)
