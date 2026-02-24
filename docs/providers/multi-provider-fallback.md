# Multi-Provider Fallback Configuration

DomeClaw supports **automatic failover across multiple LLM providers**. If your primary provider fails (rate limit, timeout, server error), it automatically falls back to the next provider in the chain.

## 🎯 How It Works

```
User Request
    ↓
Primary Provider (kimi-code)
    ↓ ❌ Error (rate limit/timeout/server error)
Fallback 1 (qwen)
    ↓ ❌ Error
Fallback 2 (deepseek)
    ↓ ✅ Success
Response to User
```

## ⚙️ Configuration Methods

### Method 1: Using `model_list` with Multiple Entries (Recommended)

Configure multiple models with the same `model_name` for **load balancing + fallback**:

```json
{
  "model_list": [
    {
      "model_name": "my-ai",
      "model": "kimi-code/kimi-coding",
      "api_key": "sk-kimi-key",
      "api_base": "https://api.kimi.com/coding/v1"
    },
    {
      "model_name": "my-ai",
      "model": "qwen/qwen-max",
      "api_key": "sk-qwen-key",
      "api_base": "https://dashscope.aliyuncs.com/api/v1"
    },
    {
      "model_name": "my-ai",
      "model": "deepseek/deepseek-chat",
      "api_key": "sk-deepseek-key"
    }
  ],
  "agents": {
    "defaults": {
      "model": "my-ai"
    }
  }
}
```

**Behavior:**
- Uses **round-robin** load balancing across all models with the same name
- If one fails, automatically falls back to the next
- Each model can have different providers, API keys, and base URLs

---

### Method 2: Using `model_fallbacks` (Explicit Fallback List)

Configure primary + explicit fallback list:

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.domeclaw/workspace",
      "model": "kimi-code",
      "model_fallbacks": [
        "qwen",
        "deepseek",
        "groq/llama-3-70b"
      ]
    }
  },
  "model_list": [
    {
      "model_name": "kimi-code",
      "model": "kimi-code/kimi-coding",
      "api_key": "sk-kimi-key",
      "api_base": "https://api.kimi.com/coding/v1"
    },
    {
      "model_name": "qwen",
      "model": "qwen/qwen-max",
      "api_key": "sk-qwen-key"
    },
    {
      "model_name": "deepseek",
      "model": "deepseek/deepseek-chat",
      "api_key": "sk-deepseek-key"
    },
    {
      "model_name": "groq/llama-3-70b",
      "model": "groq/llama-3-70b-8192",
      "api_key": "gsk-key"
    }
  ]
}
```

**Behavior:**
- Always tries `kimi-code` first
- On failure, tries `qwen` → `deepseek` → `groq` in order
- Stops at first success

---

### Method 3: Structured Model Config (Advanced)

Use object format for fine-grained control:

```json
{
  "agents": {
    "defaults": {
      "model": {
        "primary": "kimi-code/kimi-coding",
        "fallbacks": [
          "qwen/qwen-max",
          "deepseek/deepseek-chat",
          "openai/gpt-4o-mini"
        ]
      }
    }
  },
  "model_list": [
    {
      "model_name": "kimi-code/kimi-coding",
      "model": "kimi-code/kimi-coding",
      "api_key": "sk-kimi-key",
      "api_base": "https://api.kimi.com/coding/v1"
    },
    {
      "model_name": "qwen/qwen-max",
      "model": "qwen/qwen-max",
      "api_key": "sk-qwen-key"
    },
    {
      "model_name": "deepseek/deepseek-chat",
      "model": "deepseek/deepseek-chat",
      "api_key": "sk-deepseek-key"
    },
    {
      "model_name": "openai/gpt-4o-mini",
      "model": "openai/gpt-4o-mini",
      "api_key": "sk-openai-key"
    }
  ]
}
```

---

## 🔧 Error Types That Trigger Fallback

DomeClaw automatically classifies errors and triggers fallback for **retriable errors**:

### ✅ **Retriable Errors (Triggers Fallback)**

| Error Type | Examples | Behavior |
|------------|----------|----------|
| **Rate Limit** | `429 Too Many Requests`, `rate_limit`, `quota exceeded` | Falls back to next provider |
| **Timeout** | `timeout`, `deadline exceeded`, `context deadline exceeded` | Falls back to next provider |
| **Server Error** | `500`, `502`, `503`, `529`, `overloaded` | Falls back to next provider |
| **Billing** | `402 Payment Required`, `insufficient credits`, `insufficient balance` | Falls back to next provider |
| **Auth** | `401 Unauthorized`, `invalid api key`, `expired token` | Falls back to next provider |

### ❌ **Non-Retriable Errors (No Fallback)**

| Error Type | Examples | Behavior |
|------------|----------|----------|
| **Format Error** | `invalid request format`, `tool_use.id`, `string should match pattern` | Returns error immediately |
| **Image Dimension** | `image dimensions exceed max` | Returns error immediately |
| **User Abort** | `context canceled` | Aborts immediately |

---

## 🛡️ Cooldown System

To prevent hammering failing providers, DomeClaw implements a **cooldown system**:

```json
{
  "agents": {
    "defaults": {
      "model": "kimi-code",
      "model_fallbacks": ["qwen", "deepseek"]
    }
  }
}
```

**Behavior:**
1. If `kimi-code` fails with rate limit → enters cooldown (e.g., 60 seconds)
2. Next request skips `kimi-code`, tries `qwen` directly
3. After cooldown expires, `kimi-code` is tried again
4. Success resets cooldown counter

**Cooldown Durations (by error type):**
- Rate Limit: 60 seconds
- Timeout: 30 seconds
- Server Error: 120 seconds
- Auth Error: 300 seconds (5 minutes)

---

## 📊 Load Balancing + Fallback

Combine load balancing with fallback for high availability:

```json
{
  "model_list": [
    {
      "model_name": "primary-model",
      "model": "kimi-code/kimi-coding",
      "api_key": "sk-kimi-key-1"
    },
    {
      "model_name": "primary-model",
      "model": "kimi-code/kimi-coding",
      "api_key": "sk-kimi-key-2"
    },
    {
      "model_name": "primary-model",
      "model": "kimi-code/kimi-coding",
      "api_key": "sk-kimi-key-3"
    },
    {
      "model_name": "backup-model",
      "model": "qwen/qwen-max",
      "api_key": "sk-qwen-key"
    }
  ],
  "agents": {
    "defaults": {
      "model": "primary-model",
      "model_fallbacks": ["backup-model"]
    }
  }
}
```

**Behavior:**
- Round-robin across 3 Kimi Code API keys
- If all Kimi keys fail → fallback to Qwen
- Maximum availability and rate limit distribution

---

## 🎯 Real-World Examples

### Example 1: Kimi Code + Qwen + DeepSeek (Cost Optimization)

```json
{
  "model_list": [
    {
      "model_name": "coding-assistant",
      "model": "kimi-code/kimi-coding",
      "api_key": "sk-kimi-key",
      "api_base": "https://api.kimi.com/coding/v1"
    },
    {
      "model_name": "coding-assistant",
      "model": "qwen/qwen-max",
      "api_key": "sk-qwen-key"
    },
    {
      "model_name": "coding-assistant",
      "model": "deepseek/deepseek-chat",
      "api_key": "sk-deepseek-key"
    }
  ],
  "agents": {
    "defaults": {
      "model": "coding-assistant"
    }
  }
}
```

**Strategy:**
- Primary: Kimi Code (best for coding, lowest cost)
- Fallback 1: Qwen (good balance)
- Fallback 2: DeepSeek (reliable backup)

---

### Example 2: Production High Availability

```json
{
  "model_list": [
    {
      "model_name": "production-model",
      "model": "openai/gpt-4o",
      "api_key": "sk-openai-key-1"
    },
    {
      "model_name": "production-model",
      "model": "openai/gpt-4o",
      "api_key": "sk-openai-key-2"
    },
    {
      "model_name": "production-model",
      "model": "anthropic/claude-sonnet-4.6",
      "api_key": "sk-ant-key-1"
    },
    {
      "model_name": "production-model",
      "model": "anthropic/claude-sonnet-4.6",
      "api_key": "sk-ant-key-2"
    }
  ],
  "agents": {
    "defaults": {
      "model": "production-model"
    }
  }
}
```

**Strategy:**
- 2 OpenAI keys (load balanced)
- 2 Anthropic keys (load balanced)
- Automatic failover between providers
- Zero downtime during outages

---

### Example 3: Budget Setup (Free Tier Friendly)

```json
{
  "model_list": [
    {
      "model_name": "free-model",
      "model": "groq/llama-3-70b",
      "api_key": "gsk-key"
    },
    {
      "model_name": "free-model",
      "model": "qwen/qwen-turbo",
      "api_key": "sk-qwen-key"
    }
  ],
  "agents": {
    "defaults": {
      "model": "free-model"
    }
  }
}
```

**Strategy:**
- Groq free tier (fast, limited quota)
- Qwen free tier as backup
- Maximizes free quota usage

---

## 🔍 Monitoring Fallback Behavior

When fallback occurs, you'll see logs like:

```
[INFO] [Agent] Primary provider kimi-code failed: rate limit exceeded
[INFO] [Agent] Falling back to qwen/qwen-max
[INFO] [Agent] Fallback successful: qwen/qwen-max (1.2s)
```

**Log Fields:**
- `Provider`: Which provider failed/succeeded
- `Model`: Which model was used
- `Reason`: Why fallback occurred (rate_limit, timeout, etc.)
- `Duration`: How long the attempt took
- `Attempt`: Which attempt number (1 = primary, 2 = first fallback, etc.)

---

## ⚠️ Best Practices

### ✅ Do's

1. **Always configure at least 1 fallback** for production use
2. **Use different providers** (not just different models from same provider)
3. **Test fallback behavior** by simulating failures
4. **Monitor fallback frequency** to identify problematic providers
5. **Use cooldown wisely** - don't set too short or too long

### ❌ Don'ts

1. **Don't configure too many fallbacks** (3-5 is optimal)
2. **Don't mix incompatible models** (e.g., very different context windows)
3. **Don't ignore fallback logs** - they indicate provider issues
4. **Don't use fallback for non-retriable errors** (format errors won't help)

---

## 🧪 Testing Fallback

### Test 1: Simulate Rate Limit

```bash
# Configure with fake rate-limited key
{
  "model_list": [
    {
      "model_name": "test-model",
      "model": "kimi-code/kimi-coding",
      "api_key": "sk-invalid-key"
    },
    {
      "model_name": "test-model",
      "model": "qwen/qwen-max",
      "api_key": "sk-valid-key"
    }
  ]
}

# Run agent
domeclaw agent -m "Hello"

# Expected: Falls back to qwen after kimi-code fails
```

### Test 2: Monitor Fallback in Logs

```bash
# Enable debug logging
export PICOCLAW_LOG_LEVEL=debug

# Run agent
domeclaw gateway

# Watch for fallback events in logs
tail -f ~/.domeclaw/domeclaw.log | grep -i fallback
```

---

## 📈 Performance Considerations

### Latency

- **Primary Success**: ~1-3s (normal)
- **1 Fallback**: ~3-6s (primary fail + fallback success)
- **2 Fallbacks**: ~6-9s (2 fails + final success)

### Cost

- Fallback to cheaper models can **reduce costs**
- Example: Kimi Code ($0.01/M tokens) → Qwen ($0.005/M tokens)

### Reliability

- **1 Provider**: 99% uptime
- **2 Providers**: 99.99% uptime
- **3+ Providers**: 99.999% uptime

---

## 🆘 Troubleshooting

### Issue: Fallback Not Triggering

**Check:**
1. Error type is retriable (see table above)
2. Fallback models are properly configured
3. API keys are valid for fallback models

**Solution:**
```json
{
  "agents": {
    "defaults": {
      "model": "kimi-code",
      "model_fallbacks": ["qwen", "deepseek"]  // Ensure this exists
    }
  }
}
```

---

### Issue: All Fallbacks Fail

**Symptoms:**
```
Error: fallback exhausted: all providers failed
- Attempt 1: kimi-code (rate_limit, 2.1s)
- Attempt 2: qwen (timeout, 30.0s)
- Attempt 3: deepseek (server_error, 5.2s)
```

**Solutions:**
1. Add more fallback providers
2. Increase timeout duration
3. Check network connectivity
4. Verify all API keys are valid

---

### Issue: Fallback Too Slow

**Symptoms:** Requests take 10+ seconds

**Solutions:**
1. Reduce number of fallbacks (2-3 is optimal)
2. Set shorter timeout per provider
3. Use faster providers (Groq, local models)
4. Configure aggressive cooldown

---

## 📚 Related Documentation

- [Kimi Code Provider](kimi_code.md)
- [Qwen Provider](qwen.md)
- [Error Classification](../pkg/providers/error_classifier.go)
- [Fallback Implementation](../pkg/providers/fallback.go)
- [Configuration Guide](../../README.md#configuration)

---

## 🎓 Advanced: Custom Fallback Logic

For advanced users, you can implement custom fallback logic:

```go
// Example: Custom fallback based on model capabilities
func customFallback(primary string, fallbacks []string) string {
    if strings.Contains(primary, "coding") {
        // Prefer coding-optimized models
        return "kimi-code"
    }
    // Default fallback chain
    return fallbacks[0]
}
```

---

**Summary:** DomeClaw's multi-provider fallback ensures your AI assistant stays online even when individual providers fail. Configure 2-3 providers from different vendors for maximum reliability! 🚀
