// PicoClaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 PicoClaw contributors

package providers

import (
	"context"
	"github.com/sipeed/picoclaw/pkg/providers/openai_compat"
)

type HTTPProvider struct {
	delegate *openai_compat.Provider
}

func NewHTTPProvider(apiKey, apiBase string, proxy ...string) *HTTPProvider {
	proxyURL := ""
	if len(proxy) > 0 {
		proxyURL = proxy[0]
	}
	return &HTTPProvider{
		delegate: openai_compat.NewProvider(apiKey, apiBase, proxyURL),
	}
}

func (p *HTTPProvider) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	compatResp, err := p.delegate.Chat(ctx, toOpenAICompatMessages(messages), toOpenAICompatTools(tools), model, options)
	if err != nil {
		return nil, err
	}
	return fromOpenAICompatResponse(compatResp), nil
}

func (p *HTTPProvider) GetDefaultModel() string {
	return ""
}

func toOpenAICompatMessages(messages []Message) []openai_compat.Message {
	out := make([]openai_compat.Message, 0, len(messages))
	for _, msg := range messages {
		out = append(out, openai_compat.Message{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCalls:  toOpenAICompatToolCalls(msg.ToolCalls),
			ToolCallID: msg.ToolCallID,
		})
	}
	return out
}

func toOpenAICompatTools(tools []ToolDefinition) []openai_compat.ToolDefinition {
	out := make([]openai_compat.ToolDefinition, 0, len(tools))
	for _, t := range tools {
		out = append(out, openai_compat.ToolDefinition{
			Type: t.Type,
			Function: openai_compat.ToolFunctionDefinition{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  t.Function.Parameters,
			},
		})
	}
	return out
}

func toOpenAICompatToolCalls(toolCalls []ToolCall) []openai_compat.ToolCall {
	out := make([]openai_compat.ToolCall, 0, len(toolCalls))
	for _, tc := range toolCalls {
		var fn *openai_compat.FunctionCall
		if tc.Function != nil {
			fn = &openai_compat.FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			}
		}
		out = append(out, openai_compat.ToolCall{
			ID:        tc.ID,
			Type:      tc.Type,
			Function:  fn,
			Name:      tc.Name,
			Arguments: tc.Arguments,
		})
	}
	return out
}

func fromOpenAICompatResponse(resp *openai_compat.LLMResponse) *LLMResponse {
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
		ToolCalls:    fromOpenAICompatToolCalls(resp.ToolCalls),
		FinishReason: resp.FinishReason,
		Usage:        usage,
	}
}

func fromOpenAICompatToolCalls(toolCalls []openai_compat.ToolCall) []ToolCall {
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
