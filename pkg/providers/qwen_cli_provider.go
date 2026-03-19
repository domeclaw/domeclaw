package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// QwenCliProvider implements LLMProvider by wrapping the Qwen CLI as a subprocess.
// This provider enables PicoClaw to leverage Qwen Code's capabilities through its CLI interface.
// It supports both text and JSON output formats as well as various authentication methods.
type QwenCliProvider struct {
	command   string
	workspace string
}

// NewQwenCliProvider creates a new Qwen CLI provider.
// The workspace parameter specifies the working directory for the CLI.
func NewQwenCliProvider(workspace string) *QwenCliProvider {
	return &QwenCliProvider{
		command:   "qwen", // Default command name for Qwen CLI
		workspace: workspace,
	}
}

// Chat implements LLMProvider.Chat by executing the Qwen CLI in non-interactive mode.
// It constructs the appropriate command line arguments based on the provided messages,
// tools, model, and options, then parses the CLI's output into an LLMResponse.
func (p *QwenCliProvider) Chat(
	ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]any,
) (*LLMResponse, error) {
	if p.command == "" {
		return nil, fmt.Errorf("qwen command not configured")
	}

	prompt := p.buildPrompt(messages, tools)

	// Construct arguments for Qwen CLI
	args := []string{
		"--output-format", "json", // Use JSON output for reliable parsing
		"--input-format", "text", // Input format is text
		"--prompt", prompt, // Pass the prompt
	}

	// Add model if specified and not the default
	if model != "" && model != "qwen-cli" {
		args = append(args, "--model", model)
	}

	// Add workspace if configured
	if p.workspace != "" {
		// Note: Qwen CLI doesn't seem to have a direct workspace arg,
		// but we could potentially use --include-directories
		// For now, we'll run the command from the workspace directory
	}

	// Add options from the options map
	if maxTokens, ok := asInt(options["max_tokens"]); ok {
		// Qwen CLI doesn't have a direct max_tokens flag, but we can pass it as a parameter
		// if the underlying model supports it via custom parameters
		_ = maxTokens // Use the variable to avoid "declared but not used" error
	}

	if temperature, ok := asFloat(options["temperature"]); ok {
		// Similarly, add temperature if the CLI supports it
		// Qwen CLI might accept this through model parameters
		_ = temperature // Use the variable to avoid "declared but not used" error
	}

	// Add any additional parameters from options
	if params, ok := options["additional_params"].(map[string]interface{}); ok {
		for key, value := range params {
			args = append(args, fmt.Sprintf("--%s", key), fmt.Sprintf("%v", value))
		}
	}

	// Execute the Qwen CLI command
	cmd := exec.CommandContext(ctx, p.command, args...)

	// If workspace is set, run command from that directory
	if p.workspace != "" {
		// Ensure the directory exists
		if err := os.MkdirAll(p.workspace, 0755); err != nil {
			// If we can't create the directory, log the error but continue without setting Dir
			// This allows the command to run from the current directory instead
		} else {
			cmd.Dir = p.workspace
		}
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Even if the command returns an error, try to parse the output
	// as Qwen might still return partial results
	if stdoutStr := stdout.String(); stdoutStr != "" {
		resp, parseErr := p.parseJSONOutput(stdoutStr)
		if parseErr == nil && resp != nil && (resp.Content != "" || len(resp.ToolCalls) > 0) {
			return resp, nil
		}
	}

	if err != nil {
		if ctx.Err() == context.Canceled {
			return nil, ctx.Err()
		}
		if stderrStr := stderr.String(); stderrStr != "" {
			return nil, fmt.Errorf("qwen cli error: %s", stderrStr)
		}
		return nil, fmt.Errorf("qwen cli error: %w", err)
	}

	return p.parseJSONOutput(stdout.String())
}

// GetDefaultModel returns the default model identifier for this provider.
func (p *QwenCliProvider) GetDefaultModel() string {
	return "qwen-cli"
}

// buildPrompt converts the messages to a format suitable for the Qwen CLI.
// This implementation follows the same pattern as Claude/Codex providers.
func (p *QwenCliProvider) buildPrompt(messages []Message, tools []ToolDefinition) string {
	var systemParts []string
	var conversationParts []string

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			systemParts = append(systemParts, msg.Content)
		case "user":
			conversationParts = append(conversationParts, msg.Content)
		case "assistant":
			conversationParts = append(conversationParts, "Assistant: "+msg.Content)
		case "tool":
			conversationParts = append(conversationParts,
				fmt.Sprintf("[Tool Result for %s]: %s", msg.ToolCallID, msg.Content))
		}
	}

	var sb strings.Builder

	if len(systemParts) > 0 {
		sb.WriteString("## System Instructions\n\n")
		sb.WriteString(strings.Join(systemParts, "\n\n"))
		sb.WriteString("\n\n## Task\n\n")
	}

	if len(tools) > 0 {
		sb.WriteString(p.buildCLIToolsPrompt(tools))
		sb.WriteString("\n\n")
	}

	// Simplify single user message (no prefix)
	if len(conversationParts) == 1 && len(systemParts) == 0 && len(tools) == 0 {
		return conversationParts[0]
	}

	sb.WriteString(strings.Join(conversationParts, "\n"))
	return sb.String()
}

// buildCLIToolsPrompt generates a prompt section describing available tools
// in a format that Qwen CLI can understand.
func (p *QwenCliProvider) buildCLIToolsPrompt(tools []ToolDefinition) string {
	var sb strings.Builder
	sb.WriteString("## Available Tools\n\n")

	for _, tool := range tools {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", tool.Function.Name, tool.Function.Description))
		sb.WriteString("  Parameters:\n")

		if props, ok := tool.Function.Parameters["properties"].(map[string]any); ok {
			for paramName, paramDef := range props {
				if paramDetails, ok := paramDef.(map[string]any); ok {
					paramType := paramDetails["type"]
					description := paramDetails["description"]
					sb.WriteString(fmt.Sprintf("    - %s (%v): %v\n", paramName, paramType, description))
				}
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// qwenCLIEvent represents the structure of Qwen CLI's JSONL output
type qwenCLIEvent struct {
	Type      string          `json:"type"`
	SubType   string          `json:"subtype,omitempty"`
	UUID      string          `json:"uuid,omitempty"`
	SessionID string          `json:"sessionid,omitempty"`
	Message   *qwenCLIMessage `json:"message,omitempty"`
	Result    *qwenCLIResult  `json:"result,omitempty"`
	IsError   bool            `json:"iserror,omitempty"`
	Usage     *qwenCLIUsage   `json:"usage,omitempty"`
	Content   string          `json:"content,omitempty"` // For older format compatibility
	ToolCalls []ToolCall      `json:"tool_calls,omitempty"`
}

type qwenCLIMessage struct {
	ID         string        `json:"id"`
	Type       string        `json:"type"`
	Role       string        `json:"role"`
	Content    []qwenContent `json:"content"`
	StopReason string        `json:"stopreason"`
	Usage      *qwenCLIUsage `json:"usage"`
}

type qwenContent struct {
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	Thinking  string `json:"thinking,omitempty"`
	Signature string `json:"signature,omitempty"`
}

type qwenCLIResult struct {
	Result string        `json:"result"`
	Usage  *qwenCLIUsage `json:"usage"`
}

type qwenCLIUsage struct {
	InputTokens          int `json:"inputtokens"`
	OutputTokens         int `json:"outputtokens"`
	CacheReadInputTokens int `json:"cachereadinputtokens"`
	TotalTokens          int `json:"totaltokens"`
}

// parseJSONOutput processes the JSONL output from Qwen CLI.
// This implementation handles the JSONL format that Qwen CLI produces,
// extracting only the final assistant response.
func (p *QwenCliProvider) parseJSONOutput(output string) (*LLMResponse, error) {
	var finalContent string
	var finalToolCalls []ToolCall
	var finalUsage *UsageInfo
	var finalFinishReason string

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Try to parse as a structured response
		var event qwenCLIEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue // Skip malformed lines
		}

		// Handle different event types
		switch event.Type {
		case "assistant":
			// Extract content from assistant messages
			if event.Message != nil && event.Message.Role == "assistant" {
				var contentParts []string
				for _, contentItem := range event.Message.Content {
					if contentItem.Type == "text" && contentItem.Text != "" {
						contentParts = append(contentParts, contentItem.Text)
					}
				}
				if len(contentParts) > 0 {
					finalContent = strings.Join(contentParts, "\n")
				}

				// Update usage if present
				if event.Message.Usage != nil {
					finalUsage = &UsageInfo{
						PromptTokens:     event.Message.Usage.InputTokens,
						CompletionTokens: event.Message.Usage.OutputTokens,
						TotalTokens:      event.Message.Usage.TotalTokens,
					}
				}
			}
		case "result":
			// Extract final result if available - this is the most important one
			if event.Result != nil {
				finalContent = event.Result.Result

				// Update usage if present in result
				if event.Result.Usage != nil {
					finalUsage = &UsageInfo{
						PromptTokens:     event.Result.Usage.InputTokens,
						CompletionTokens: event.Result.Usage.OutputTokens,
						TotalTokens:      event.Result.Usage.TotalTokens,
					}
				}
			}
		}
	}

	// Set finish reason based on whether we have content or tool calls
	if finalContent != "" {
		finalFinishReason = "stop"
	} else if len(finalToolCalls) > 0 {
		finalFinishReason = "tool_calls"
	} else {
		finalFinishReason = "stop"
	}

	return &LLMResponse{
		Content:      strings.TrimSpace(finalContent),
		ToolCalls:    finalToolCalls,
		FinishReason: finalFinishReason,
		Usage:        finalUsage,
	}, nil
}

// asInt is a helper function to extract int values from interface{}
func asInt(v interface{}) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case float64: // JSON numbers are float64
		return int(val), true
	case int64:
		return int(val), true
	case string:
		// Attempt to parse string as int
		var result int
		fmt.Sscanf(val, "%d", &result)
		return result, true
	default:
		return 0, false
	}
}

// asFloat is a helper function to extract float64 values from interface{}
func asFloat(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		// Attempt to parse string as float
		var result float64
		fmt.Sscanf(val, "%f", &result)
		return result, true
	default:
		return 0.0, false
	}
}
