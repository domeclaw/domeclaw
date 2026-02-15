package providers

import (
	"context"
	"fmt"

	"github.com/sipeed/picoclaw/pkg/auth"
	anthropicprovider "github.com/sipeed/picoclaw/pkg/providers/anthropic"
)

type ClaudeProvider struct {
	delegate *anthropicprovider.Provider
}

func NewClaudeProvider(token string) *ClaudeProvider {
	return &ClaudeProvider{
		delegate: anthropicprovider.NewProvider(token),
	}
}

func NewClaudeProviderWithTokenSource(token string, tokenSource func() (string, error)) *ClaudeProvider {
	return &ClaudeProvider{
		delegate: anthropicprovider.NewProviderWithTokenSource(token, tokenSource),
	}
}

func newClaudeProviderWithDelegate(delegate *anthropicprovider.Provider) *ClaudeProvider {
	return &ClaudeProvider{delegate: delegate}
}

func (p *ClaudeProvider) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	resp, err := p.delegate.Chat(
		ctx,
		toAnthropicProviderMessages(messages),
		toAnthropicProviderTools(tools),
		model,
		options,
	)
	if err != nil {
		return nil, err
	}
	return fromAnthropicProviderResponse(resp), nil
}

func (p *ClaudeProvider) GetDefaultModel() string {
	return p.delegate.GetDefaultModel()
}

func createClaudeTokenSource() func() (string, error) {
	return func() (string, error) {
		cred, err := auth.GetCredential("anthropic")
		if err != nil {
			return "", fmt.Errorf("loading auth credentials: %w", err)
		}
		if cred == nil {
			return "", fmt.Errorf("no credentials for anthropic. Run: picoclaw auth login --provider anthropic")
		}
		return cred.AccessToken, nil
	}
}

func toAnthropicProviderMessages(messages []Message) []anthropicprovider.Message {
	out := make([]anthropicprovider.Message, 0, len(messages))
	for _, msg := range messages {
		out = append(out, anthropicprovider.Message{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCalls:  toAnthropicProviderToolCalls(msg.ToolCalls),
			ToolCallID: msg.ToolCallID,
		})
	}
	return out
}

func toAnthropicProviderTools(tools []ToolDefinition) []anthropicprovider.ToolDefinition {
	out := make([]anthropicprovider.ToolDefinition, 0, len(tools))
	for _, t := range tools {
		out = append(out, anthropicprovider.ToolDefinition{
			Type: t.Type,
			Function: anthropicprovider.ToolFunctionDefinition{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  t.Function.Parameters,
			},
		})
	}
	return out
}

func toAnthropicProviderToolCalls(toolCalls []ToolCall) []anthropicprovider.ToolCall {
	out := make([]anthropicprovider.ToolCall, 0, len(toolCalls))
	for _, tc := range toolCalls {
		var fn *anthropicprovider.FunctionCall
		if tc.Function != nil {
			fn = &anthropicprovider.FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			}
		}
		out = append(out, anthropicprovider.ToolCall{
			ID:        tc.ID,
			Type:      tc.Type,
			Function:  fn,
			Name:      tc.Name,
			Arguments: tc.Arguments,
		})
	}
	return out
}

func fromAnthropicProviderResponse(resp *anthropicprovider.LLMResponse) *LLMResponse {
	if resp == nil {
		return &LLMResponse{}
	}

	var usage *UsageInfo
	if resp.Usage != nil {
		usage = &UsageInfo{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	return &LLMResponse{
		Content:      resp.Content,
		ToolCalls:    fromAnthropicProviderToolCalls(resp.ToolCalls),
		FinishReason: resp.FinishReason,
		Usage:        usage,
	}
}

func fromAnthropicProviderToolCalls(toolCalls []anthropicprovider.ToolCall) []ToolCall {
	out := make([]ToolCall, 0, len(toolCalls))
	for _, tc := range toolCalls {
		var fn *FunctionCall
		if tc.Function != nil {
			fn = &FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			}
		}
		out = append(out, ToolCall{
			ID:        tc.ID,
			Type:      tc.Type,
			Function:  fn,
			Name:      tc.Name,
			Arguments: tc.Arguments,
		})
	}
	return out
}
