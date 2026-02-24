package providers

import (
	"testing"

	"github.com/sipeed/domeclaw/pkg/config"
)

func TestKimiCodeProviderSelection(t *testing.T) {
	tests := []struct {
		name             string
		providerName     string
		model            string
		apiKey           string
		apiBase          string
		wantProvider     string
		wantAPIBase      string
		wantProviderType providerType
	}{
		{
			name:             "kimi_code provider with explicit config",
			providerName:     "kimi_code",
			model:            "kimi-coding",
			apiKey:           "sk-test-key",
			apiBase:          "https://api.kimi.com/coding/v1",
			wantProvider:     "kimi_code",
			wantAPIBase:      "https://api.kimi.com/coding/v1",
			wantProviderType: providerTypeHTTPCompat,
		},
		{
			name:             "kimi-code provider with explicit config",
			providerName:     "kimi-code",
			model:            "kimi-coding",
			apiKey:           "sk-test-key",
			apiBase:          "",
			wantProvider:     "kimi-code",
			wantAPIBase:      "https://api.kimi.com/coding/v1",
			wantProviderType: providerTypeHTTPCompat,
		},
		{
			name:             "kimi_code model prefix",
			providerName:     "",
			model:            "kimi-code/kimi-coding",
			apiKey:           "sk-test-key",
			apiBase:          "",
			wantProvider:     "kimi_code",
			wantAPIBase:      "https://api.kimi.com/coding/v1",
			wantProviderType: providerTypeHTTPCompat,
		},
		{
			name:             "kimi_code with custom api base",
			providerName:     "kimi_code",
			model:            "kimi-coding",
			apiKey:           "sk-test-key",
			apiBase:          "https://custom.kimi.api/v1",
			wantProvider:     "kimi_code",
			wantAPIBase:      "https://custom.kimi.api/v1",
			wantProviderType: providerTypeHTTPCompat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Agents: config.AgentsConfig{
					Defaults: config.AgentDefaults{
						Provider: tt.providerName,
						Model:    tt.model,
					},
				},
				Providers: config.ProvidersConfig{
					KimiCode: config.ProviderConfig{
						APIKey:  tt.apiKey,
						APIBase: tt.apiBase,
					},
				},
			}

			sel, err := resolveProviderSelection(cfg)
			if err != nil {
				t.Fatalf("resolveProviderSelection() error = %v", err)
			}

			if sel.apiKey != tt.apiKey {
				t.Errorf("apiKey = %v, want %v", sel.apiKey, tt.apiKey)
			}

			if sel.apiBase != tt.wantAPIBase {
				t.Errorf("apiBase = %v, want %v", sel.apiBase, tt.wantAPIBase)
			}

			if sel.providerType != tt.wantProviderType {
				t.Errorf("providerType = %v, want %v", sel.providerType, tt.wantProviderType)
			}
		})
	}
}

func TestKimiCodeProviderFallback(t *testing.T) {
	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: config.AgentDefaults{
				Model: "kimi-code/kimi-coding",
			},
		},
		Providers: config.ProvidersConfig{
			KimiCode: config.ProviderConfig{
				APIKey: "sk-test-key",
			},
			Moonshot: config.ProviderConfig{
				APIKey: "sk-moonshot-key",
			},
		},
	}

	sel, err := resolveProviderSelection(cfg)
	if err != nil {
		t.Fatalf("resolveProviderSelection() error = %v", err)
	}

	if sel.apiKey != "sk-test-key" {
		t.Errorf("apiKey = %v, want sk-test-key", sel.apiKey)
	}

	if sel.apiBase != "https://api.kimi.com/coding/v1" {
		t.Errorf("apiBase = %v, want https://api.kimi.com/coding/v1", sel.apiBase)
	}
}

func TestKimiCodeVsMoonshot(t *testing.T) {
	tests := []struct {
		name        string
		model       string
		kimiKey     string
		moonshotKey string
		wantAPIBase string
	}{
		{
			name:        "kimi-code uses kimi_code provider",
			model:       "kimi-code/kimi-coding",
			kimiKey:     "sk-kimi-key",
			moonshotKey: "sk-moonshot-key",
			wantAPIBase: "https://api.kimi.com/coding/v1",
		},
		{
			name:        "kimi without code prefix uses moonshot",
			model:       "kimi-k2",
			kimiKey:     "sk-kimi-key",
			moonshotKey: "sk-moonshot-key",
			wantAPIBase: "https://api.moonshot.cn/v1",
		},
		{
			name:        "moonshot prefix uses moonshot provider",
			model:       "moonshot/moonshot-v1",
			kimiKey:     "sk-kimi-key",
			moonshotKey: "sk-moonshot-key",
			wantAPIBase: "https://api.moonshot.cn/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Agents: config.AgentsConfig{
					Defaults: config.AgentDefaults{
						Model: tt.model,
					},
				},
				Providers: config.ProvidersConfig{
					KimiCode: config.ProviderConfig{
						APIKey: tt.kimiKey,
					},
					Moonshot: config.ProviderConfig{
						APIKey: tt.moonshotKey,
					},
				},
			}

			sel, err := resolveProviderSelection(cfg)
			if err != nil {
				t.Fatalf("resolveProviderSelection() error = %v", err)
			}

			if sel.apiBase != tt.wantAPIBase {
				t.Errorf("apiBase = %v, want %v", sel.apiBase, tt.wantAPIBase)
			}
		})
	}
}
