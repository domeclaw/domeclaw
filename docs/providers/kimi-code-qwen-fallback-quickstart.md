# Quick Start: Kimi Code + Qwen Fallback

This guide shows you how to configure DomeClaw with **Kimi Code as primary** and **Qwen as fallback** provider for high availability.

## 🎯 Why This Setup?

- **Kimi Code**: Excellent for coding tasks, competitive pricing
- **Qwen**: Reliable backup, good general-purpose model
- **Automatic Fallback**: If Kimi Code fails (rate limit, timeout), automatically uses Qwen
- **Zero Downtime**: Your AI assistant stays online even during provider outages

## ⚡ Quick Setup (5 minutes)

### Step 1: Get API Keys

**Kimi Code:**
1. Visit [Kimi Code Platform](https://platform.kimi.com/)
2. Create account / login
3. Go to API Keys section
4. Create new API key
5. Copy the key (starts with `sk-`)

**Qwen (Alibaba DashScope):**
1. Visit [DashScope Console](https://dashscope.console.aliyun.com/)
2. Create account / login
3. Go to API Key Management
4. Create new API key
5. Copy the key (starts with `sk-`)

---

### Step 2: Create Config File

Create or edit `~/.domeclaw/config.json`:

```json
{
  "model_list": [
    {
      "model_name": "my-coding-assistant",
      "model": "kimi-code/kimi-coding",
      "api_key": "sk-your-kimi-code-key-here",
      "api_base": "https://api.kimi.com/coding/v1"
    },
    {
      "model_name": "my-coding-assistant",
      "model": "qwen/qwen-max",
      "api_key": "sk-your-qwen-key-here",
      "api_base": "https://dashscope.aliyuncs.com/api/v1"
    }
  ],
  "agents": {
    "defaults": {
      "workspace": "~/.domeclaw/workspace",
      "model": "my-coding-assistant",
      "max_tokens": 8192,
      "temperature": 0.7,
      "max_tool_iterations": 20
    }
  },
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

**Replace:**
- `sk-your-kimi-code-key-here` with your actual Kimi Code API key
- `sk-your-qwen-key-here` with your actual Qwen API key

---

### Step 3: Test Configuration

```bash
# Test Kimi Code (primary)
domeclaw agent -m "Hello, can you help me write a Python function?"

# Expected output:
# 🦞 DomeClaw is thinking... (using kimi-code/kimi-coding)
# ✅ Response received from Kimi Code
```

---

### Step 4: Test Fallback (Optional)

To verify fallback works, temporarily break the Kimi Code config:

```json
{
  "model_list": [
    {
      "model_name": "my-coding-assistant",
      "model": "kimi-code/kimi-coding",
      "api_key": "invalid-key",
      "api_base": "https://api.kimi.com/coding/v1"
    },
    {
      "model_name": "my-coding-assistant",
      "model": "qwen/qwen-max",
      "api_key": "sk-your-qwen-key-here",
      "api_base": "https://dashscope.aliyuncs.com/api/v1"
    }
  ]
}
```

```bash
# Test with invalid Kimi key
domeclaw agent -m "Test fallback"

# Expected output:
# 🦞 DomeClaw is thinking...
# ⚠️  Primary provider (kimi-code) failed: invalid api key
# 🔄 Falling back to qwen/qwen-max
# ✅ Response received from Qwen (fallback)
```

**Don't forget to restore the valid Kimi key after testing!**

---

## 📊 How It Works

```
User Request
    ↓
Try: kimi-code/kimi-coding
    ↓ ❌ (if rate limit / timeout / server error)
Try: qwen/qwen-max
    ↓ ✅
Return Response
```

**Fallback Triggers:**
- ✅ Rate limit (429)
- ✅ Timeout
- ✅ Server error (500, 502, 503)
- ✅ Insufficient credits
- ✅ API key invalid

**No Fallback:**
- ❌ Invalid request format (client error)
- ❌ User cancellation

---

## 🔧 Configuration Options

### Option 1: Load Balancing + Fallback (Recommended)

Both models share the same name → round-robin + fallback:

```json
{
  "model_list": [
    {
      "model_name": "my-model",
      "model": "kimi-code/kimi-coding",
      "api_key": "sk-kimi-key"
    },
    {
      "model_name": "my-model",
      "model": "qwen/qwen-max",
      "api_key": "sk-qwen-key"
    }
  ],
  "agents": {
    "defaults": {
      "model": "my-model"
    }
  }
}
```

**Behavior:**
- 50% requests to Kimi Code, 50% to Qwen (load balancing)
- If one fails, uses the other (fallback)

---

### Option 2: Primary + Explicit Fallback

Kimi Code always first, Qwen only on failure:

```json
{
  "model_list": [
    {
      "model_name": "kimi-primary",
      "model": "kimi-code/kimi-coding",
      "api_key": "sk-kimi-key"
    },
    {
      "model_name": "qwen-backup",
      "model": "qwen/qwen-max",
      "api_key": "sk-qwen-key"
    }
  ],
  "agents": {
    "defaults": {
      "model": "kimi-primary",
      "model_fallbacks": ["qwen-backup"]
    }
  }
}
```

**Behavior:**
- Always tries Kimi Code first
- Only uses Qwen if Kimi fails

---

### Option 3: Multiple Fallbacks

Add more providers for maximum reliability:

```json
{
  "model_list": [
    {
      "model_name": "kimi-code",
      "model": "kimi-code/kimi-coding",
      "api_key": "sk-kimi-key"
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
    }
  ],
  "agents": {
    "defaults": {
      "model": "kimi-code",
      "model_fallbacks": ["qwen", "deepseek"]
    }
  }
}
```

**Fallback Chain:**
1. Kimi Code (primary)
2. Qwen (first fallback)
3. DeepSeek (second fallback)

---

## 📈 Monitoring

### Check Logs

```bash
# View real-time logs
tail -f ~/.domeclaw/domeclaw.log

# Filter for fallback events
grep -i "fallback" ~/.domeclaw/domeclaw.log
```

**Example Log Output:**
```
[2026-02-24T10:30:15Z] [INFO] [Agent] Primary provider kimi-code failed: rate limit exceeded
[2026-02-24T10:30:15Z] [INFO] [Agent] Falling back to qwen/qwen-max
[2026-02-24T10:30:17Z] [INFO] [Agent] Fallback successful: qwen/qwen-max (2.1s)
```

### Check Status

```bash
domeclaw status
```

**Output:**
```
DomeClaw Status
===============
Primary Provider: kimi-code (active)
Fallback Providers: qwen (standby)
Total Requests: 150
Fallback Events: 5 (3.3%)
```

---

## 💰 Cost Estimation

**Pricing (as of 2026-02):**

| Provider | Input (per 1M tokens) | Output (per 1M tokens) |
|----------|----------------------|------------------------|
| Kimi Code | ~$0.01 | ~$0.03 |
| Qwen | ~$0.005 | ~$0.01 |

**Monthly Cost Example:**
- 100,000 requests/day
- Average 500 tokens/request
- 95% Kimi Code, 5% fallback to Qwen

```
Kimi Code: 100,000 × 0.95 × 500 × 30 = 1.425B tokens/month
Qwen: 100,000 × 0.05 × 500 × 30 = 75M tokens/month

Cost:
- Kimi: 1.425B × $0.02 (avg) = $28.50/month
- Qwen: 75M × $0.0075 (avg) = $0.56/month
Total: ~$29/month
```

---

## ⚠️ Troubleshooting

### Issue: "Invalid API Key"

**Solution:**
1. Double-check API key in config
2. Ensure no extra spaces
3. Verify key is active in provider dashboard
4. Check for key expiration

---

### Issue: "Rate Limit Exceeded"

**Solutions:**
1. Wait for cooldown (60 seconds)
2. Add more API keys (load balancing)
3. Upgrade to higher quota plan
4. Reduce request frequency

**Example: Multiple API Keys**
```json
{
  "model_list": [
    {
      "model_name": "kimi-code",
      "model": "kimi-code/kimi-coding",
      "api_key": "sk-kimi-key-1"
    },
    {
      "model_name": "kimi-code",
      "model": "kimi-code/kimi-coding",
      "api_key": "sk-kimi-key-2"
    },
    {
      "model_name": "qwen",
      "model": "qwen/qwen-max",
      "api_key": "sk-qwen-key"
    }
  ]
}
```

---

### Issue: "Timeout"

**Solutions:**
1. Check network connectivity
2. Reduce `max_tokens` in config
3. Use closer API endpoint (if available)
4. Add timeout configuration:

```json
{
  "providers": {
    "kimi_code": {
      "api_key": "sk-kimi-key",
      "api_base": "https://api.kimi.com/coding/v1",
      "timeout": 30
    }
  }
}
```

---

### Issue: Fallback Not Working

**Check:**
1. `model_fallbacks` is configured correctly
2. Fallback models exist in `model_list`
3. Error type is retriable (see table above)

**Debug:**
```bash
# Enable debug logging
export PICOCLAW_LOG_LEVEL=debug

# Run and check logs
domeclaw agent -m "test"
cat ~/.domeclaw/domeclaw.log | grep -A5 -B5 fallback
```

---

## 🎯 Best Practices

### ✅ Do's

1. **Always configure at least 1 fallback** for production
2. **Test fallback regularly** (simulate failures)
3. **Monitor fallback frequency** (>5% indicates issues)
4. **Use different providers** (not just different models)
5. **Keep API keys secure** (use environment variables in production)

### ❌ Don'ts

1. **Don't use same provider for primary + fallback** (defeats the purpose)
2. **Don't ignore fallback logs** (indicates provider problems)
3. **Don't configure too many fallbacks** (3 is optimal)
4. **Don't mix very different models** (context window, capabilities)

---

## 🔐 Security Best Practices

### Use Environment Variables (Production)

Instead of hardcoding API keys:

```bash
export PICOCLAW_MODEL_LIST_KIMI_CODE_API_KEY="sk-kimi-key"
export PICOCLAW_MODEL_LIST_QWEN_API_KEY="sk-qwen-key"
```

Then in config:
```json
{
  "model_list": [
    {
      "model_name": "kimi-code",
      "model": "kimi-code/kimi-coding",
      "api_key": "${PICOCLAW_MODEL_LIST_KIMI_CODE_API_KEY}"
    }
  ]
}
```

---

## 📚 Next Steps

- [Multi-Provider Fallback Guide](multi-provider-fallback.md) - Advanced configuration
- [Kimi Code Provider](kimi_code.md) - Kimi-specific features
- [Qwen Provider](qwen.md) - Qwen-specific features
- [Error Classification](../pkg/providers/error_classifier.go) - How errors are classified

---

## 🆘 Getting Help

- **Documentation**: [docs/providers/](../docs/providers/)
- **GitHub Issues**: https://github.com/sipeed/domeclaw/issues
- **Discord**: https://discord.gg/V4sAZ9XWpN
- **WeChat Group**: See README.md

---

**You're all set!** Your DomeClaw instance now has automatic failover from Kimi Code to Qwen. If Kimi Code experiences any issues, Qwen will automatically take over with zero downtime! 🚀
