package providers

import (
	"fmt"
	"strings"

	"github.com/sipeed/picoclaw/pkg/auth"
	"github.com/sipeed/picoclaw/pkg/config"
)

type providerType int

const (
	providerTypeHTTPCompat providerType = iota
	providerTypeClaudeAuth
	providerTypeCodexAuth
	providerTypeClaudeCLI
	providerTypeGitHubCopilot
)

type providerSelection struct {
	providerType providerType
	apiKey       string
	apiBase      string
	proxy        string
	model        string
	workspace    string
	connectMode  string
}

func createClaudeAuthProvider() (LLMProvider, error) {
	cred, err := auth.GetCredential("anthropic")
	if err != nil {
		return nil, fmt.Errorf("loading auth credentials: %w", err)
	}
	if cred == nil {
		return nil, fmt.Errorf("no credentials for anthropic. Run: picoclaw auth login --provider anthropic")
	}
	return NewClaudeProviderWithTokenSource(cred.AccessToken, createClaudeTokenSource()), nil
}

func createCodexAuthProvider() (LLMProvider, error) {
	cred, err := auth.GetCredential("openai")
	if err != nil {
		return nil, fmt.Errorf("loading auth credentials: %w", err)
	}
	if cred == nil {
		return nil, fmt.Errorf("no credentials for openai. Run: picoclaw auth login --provider openai")
	}
	return NewCodexProviderWithTokenSource(cred.AccessToken, cred.AccountID, createCodexTokenSource()), nil
}

func resolveProviderSelection(cfg *config.Config) (providerSelection, error) {
	model := cfg.Agents.Defaults.Model
	providerName := strings.ToLower(cfg.Agents.Defaults.Provider)
	lowerModel := strings.ToLower(model)

	sel := providerSelection{
		providerType: providerTypeHTTPCompat,
		model:        model,
	}

	// First, prefer explicit provider configuration.
	if providerName != "" {
		switch providerName {
		case "groq":
			if cfg.Providers.Groq.APIKey != "" {
				sel.apiKey = cfg.Providers.Groq.APIKey
				sel.apiBase = cfg.Providers.Groq.APIBase
				if sel.apiBase == "" {
					sel.apiBase = "https://api.groq.com/openai/v1"
				}
			}
		case "openai", "gpt":
			if cfg.Providers.OpenAI.APIKey != "" || cfg.Providers.OpenAI.AuthMethod != "" {
				if cfg.Providers.OpenAI.AuthMethod == "oauth" || cfg.Providers.OpenAI.AuthMethod == "token" {
					sel.providerType = providerTypeCodexAuth
					return sel, nil
				}
				sel.apiKey = cfg.Providers.OpenAI.APIKey
				sel.apiBase = cfg.Providers.OpenAI.APIBase
				if sel.apiBase == "" {
					sel.apiBase = "https://api.openai.com/v1"
				}
			}
		case "anthropic", "claude":
			if cfg.Providers.Anthropic.APIKey != "" || cfg.Providers.Anthropic.AuthMethod != "" {
				if cfg.Providers.Anthropic.AuthMethod == "oauth" || cfg.Providers.Anthropic.AuthMethod == "token" {
					sel.providerType = providerTypeClaudeAuth
					return sel, nil
				}
				sel.apiKey = cfg.Providers.Anthropic.APIKey
				sel.apiBase = cfg.Providers.Anthropic.APIBase
				if sel.apiBase == "" {
					sel.apiBase = "https://api.anthropic.com/v1"
				}
			}
		case "openrouter":
			if cfg.Providers.OpenRouter.APIKey != "" {
				sel.apiKey = cfg.Providers.OpenRouter.APIKey
				if cfg.Providers.OpenRouter.APIBase != "" {
					sel.apiBase = cfg.Providers.OpenRouter.APIBase
				} else {
					sel.apiBase = "https://openrouter.ai/api/v1"
				}
			}
		case "zhipu", "glm":
			if cfg.Providers.Zhipu.APIKey != "" {
				sel.apiKey = cfg.Providers.Zhipu.APIKey
				sel.apiBase = cfg.Providers.Zhipu.APIBase
				if sel.apiBase == "" {
					sel.apiBase = "https://open.bigmodel.cn/api/paas/v4"
				}
			}
		case "gemini", "google":
			if cfg.Providers.Gemini.APIKey != "" {
				sel.apiKey = cfg.Providers.Gemini.APIKey
				sel.apiBase = cfg.Providers.Gemini.APIBase
				if sel.apiBase == "" {
					sel.apiBase = "https://generativelanguage.googleapis.com/v1beta"
				}
			}
		case "vllm":
			if cfg.Providers.VLLM.APIBase != "" {
				sel.apiKey = cfg.Providers.VLLM.APIKey
				sel.apiBase = cfg.Providers.VLLM.APIBase
			}
		case "shengsuanyun":
			if cfg.Providers.ShengSuanYun.APIKey != "" {
				sel.apiKey = cfg.Providers.ShengSuanYun.APIKey
				sel.apiBase = cfg.Providers.ShengSuanYun.APIBase
				if sel.apiBase == "" {
					sel.apiBase = "https://router.shengsuanyun.com/api/v1"
				}
			}
		case "claude-cli", "claude-code", "claudecode":
			workspace := cfg.Agents.Defaults.Workspace
			if workspace == "" {
				workspace = "."
			}
			sel.providerType = providerTypeClaudeCLI
			sel.workspace = workspace
			return sel, nil
		case "deepseek":
			if cfg.Providers.DeepSeek.APIKey != "" {
				sel.apiKey = cfg.Providers.DeepSeek.APIKey
				sel.apiBase = cfg.Providers.DeepSeek.APIBase
				if sel.apiBase == "" {
					sel.apiBase = "https://api.deepseek.com/v1"
				}
				if model != "deepseek-chat" && model != "deepseek-reasoner" {
					sel.model = "deepseek-chat"
				}
			}
		case "github_copilot", "copilot":
			sel.providerType = providerTypeGitHubCopilot
			if cfg.Providers.GitHubCopilot.APIBase != "" {
				sel.apiBase = cfg.Providers.GitHubCopilot.APIBase
			} else {
				sel.apiBase = "localhost:4321"
			}
			sel.connectMode = cfg.Providers.GitHubCopilot.ConnectMode
			return sel, nil
		}
	}

	// Fallback: infer provider from model and configured keys.
	if sel.apiKey == "" && sel.apiBase == "" {
		switch {
		case (strings.Contains(lowerModel, "kimi") || strings.Contains(lowerModel, "moonshot") || strings.HasPrefix(model, "moonshot/")) && cfg.Providers.Moonshot.APIKey != "":
			sel.apiKey = cfg.Providers.Moonshot.APIKey
			sel.apiBase = cfg.Providers.Moonshot.APIBase
			sel.proxy = cfg.Providers.Moonshot.Proxy
			if sel.apiBase == "" {
				sel.apiBase = "https://api.moonshot.cn/v1"
			}
		case strings.HasPrefix(model, "openrouter/") ||
			strings.HasPrefix(model, "anthropic/") ||
			strings.HasPrefix(model, "openai/") ||
			strings.HasPrefix(model, "meta-llama/") ||
			strings.HasPrefix(model, "deepseek/") ||
			strings.HasPrefix(model, "google/"):
			sel.apiKey = cfg.Providers.OpenRouter.APIKey
			sel.proxy = cfg.Providers.OpenRouter.Proxy
			if cfg.Providers.OpenRouter.APIBase != "" {
				sel.apiBase = cfg.Providers.OpenRouter.APIBase
			} else {
				sel.apiBase = "https://openrouter.ai/api/v1"
			}
		case (strings.Contains(lowerModel, "claude") || strings.HasPrefix(model, "anthropic/")) &&
			(cfg.Providers.Anthropic.APIKey != "" || cfg.Providers.Anthropic.AuthMethod != ""):
			if cfg.Providers.Anthropic.AuthMethod == "oauth" || cfg.Providers.Anthropic.AuthMethod == "token" {
				sel.providerType = providerTypeClaudeAuth
				return sel, nil
			}
			sel.apiKey = cfg.Providers.Anthropic.APIKey
			sel.apiBase = cfg.Providers.Anthropic.APIBase
			sel.proxy = cfg.Providers.Anthropic.Proxy
			if sel.apiBase == "" {
				sel.apiBase = "https://api.anthropic.com/v1"
			}
		case (strings.Contains(lowerModel, "gpt") || strings.HasPrefix(model, "openai/")) &&
			(cfg.Providers.OpenAI.APIKey != "" || cfg.Providers.OpenAI.AuthMethod != ""):
			if cfg.Providers.OpenAI.AuthMethod == "oauth" || cfg.Providers.OpenAI.AuthMethod == "token" {
				sel.providerType = providerTypeCodexAuth
				return sel, nil
			}
			sel.apiKey = cfg.Providers.OpenAI.APIKey
			sel.apiBase = cfg.Providers.OpenAI.APIBase
			sel.proxy = cfg.Providers.OpenAI.Proxy
			if sel.apiBase == "" {
				sel.apiBase = "https://api.openai.com/v1"
			}
		case (strings.Contains(lowerModel, "gemini") || strings.HasPrefix(model, "google/")) && cfg.Providers.Gemini.APIKey != "":
			sel.apiKey = cfg.Providers.Gemini.APIKey
			sel.apiBase = cfg.Providers.Gemini.APIBase
			sel.proxy = cfg.Providers.Gemini.Proxy
			if sel.apiBase == "" {
				sel.apiBase = "https://generativelanguage.googleapis.com/v1beta"
			}
		case (strings.Contains(lowerModel, "glm") || strings.Contains(lowerModel, "zhipu") || strings.Contains(lowerModel, "zai")) && cfg.Providers.Zhipu.APIKey != "":
			sel.apiKey = cfg.Providers.Zhipu.APIKey
			sel.apiBase = cfg.Providers.Zhipu.APIBase
			sel.proxy = cfg.Providers.Zhipu.Proxy
			if sel.apiBase == "" {
				sel.apiBase = "https://open.bigmodel.cn/api/paas/v4"
			}
		case (strings.Contains(lowerModel, "groq") || strings.HasPrefix(model, "groq/")) && cfg.Providers.Groq.APIKey != "":
			sel.apiKey = cfg.Providers.Groq.APIKey
			sel.apiBase = cfg.Providers.Groq.APIBase
			sel.proxy = cfg.Providers.Groq.Proxy
			if sel.apiBase == "" {
				sel.apiBase = "https://api.groq.com/openai/v1"
			}
		case (strings.Contains(lowerModel, "nvidia") || strings.HasPrefix(model, "nvidia/")) && cfg.Providers.Nvidia.APIKey != "":
			sel.apiKey = cfg.Providers.Nvidia.APIKey
			sel.apiBase = cfg.Providers.Nvidia.APIBase
			sel.proxy = cfg.Providers.Nvidia.Proxy
			if sel.apiBase == "" {
				sel.apiBase = "https://integrate.api.nvidia.com/v1"
			}
		case cfg.Providers.VLLM.APIBase != "":
			sel.apiKey = cfg.Providers.VLLM.APIKey
			sel.apiBase = cfg.Providers.VLLM.APIBase
			sel.proxy = cfg.Providers.VLLM.Proxy
		default:
			if cfg.Providers.OpenRouter.APIKey != "" {
				sel.apiKey = cfg.Providers.OpenRouter.APIKey
				sel.proxy = cfg.Providers.OpenRouter.Proxy
				if cfg.Providers.OpenRouter.APIBase != "" {
					sel.apiBase = cfg.Providers.OpenRouter.APIBase
				} else {
					sel.apiBase = "https://openrouter.ai/api/v1"
				}
			} else {
				return providerSelection{}, fmt.Errorf("no API key configured for model: %s", model)
			}
		}
	}

	if sel.providerType == providerTypeHTTPCompat {
		if sel.apiKey == "" && !strings.HasPrefix(model, "bedrock/") {
			return providerSelection{}, fmt.Errorf("no API key configured for provider (model: %s)", model)
		}
		if sel.apiBase == "" {
			return providerSelection{}, fmt.Errorf("no API base configured for provider (model: %s)", model)
		}
	}

	return sel, nil
}

func CreateProvider(cfg *config.Config) (LLMProvider, error) {
	sel, err := resolveProviderSelection(cfg)
	if err != nil {
		return nil, err
	}

	switch sel.providerType {
	case providerTypeClaudeAuth:
		return createClaudeAuthProvider()
	case providerTypeCodexAuth:
		return createCodexAuthProvider()
	case providerTypeClaudeCLI:
		return NewClaudeCliProvider(sel.workspace), nil
	case providerTypeGitHubCopilot:
		return NewGitHubCopilotProvider(sel.apiBase, sel.connectMode, sel.model)
	default:
		return NewHTTPProvider(sel.apiKey, sel.apiBase, sel.proxy), nil
	}
}
