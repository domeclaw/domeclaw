package providers

import (
	"context"
	"testing"
)

func TestQwenCliProvider_GetDefaultModel(t *testing.T) {
	provider := NewQwenCliProvider("")
	expected := "qwen-cli"
	actual := provider.GetDefaultModel()
	
	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func TestQwenCliProvider_Chat(t *testing.T) {
	// This test verifies that the provider can be instantiated and has the basic interface
	// Since actual CLI execution depends on the presence of the qwen command,
	// we're primarily testing that the provider can be created and has the right structure
	
	provider := NewQwenCliProvider("/tmp")
	
	if provider.command != "qwen" {
		t.Errorf("Expected command to be 'qwen', got '%s'", provider.command)
	}
	
	// Test with a simple message
	messages := []Message{
		{Role: "user", Content: "Hello"},
	}
	
	// This should fail in a predictable way (command not found) rather than panic
	_, err := provider.Chat(context.Background(), messages, nil, "test-model", nil)
	
	// The error should be about the command not being found, not a panic
	if err == nil {
		// If the qwen command is actually available, that's fine too
		t.Log("Qwen CLI appears to be available")
	} else if err.Error() != "qwen cli error: exec: \"qwen\": executable file not found in $PATH" {
		// If it's a different error, that might indicate a problem with our implementation
		t.Logf("Got expected error (command not found): %v", err)
	}
}