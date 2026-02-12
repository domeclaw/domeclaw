package heartbeat

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/pkg/logger"
)

// ToolResult represents a structured result from tool execution.
// This is a minimal local definition to avoid circular dependencies.
type ToolResult struct {
	ForLLM  string `json:"for_llm"`
	ForUser string `json:"for_user,omitempty"`
	Silent  bool   `json:"silent"`
	IsError bool   `json:"is_error"`
	Async   bool   `json:"async"`
	Err     error  `json:"-"`
}

// HeartbeatHandler is the function type for handling heartbeat with tool support.
// It returns a ToolResult that can indicate async operations.
type HeartbeatHandler func(prompt string) *ToolResult

type HeartbeatService struct {
	workspace   string
	onHeartbeat func(string) (string, error)
	// onHeartbeatWithTools is the new handler that supports ToolResult returns
	onHeartbeatWithTools HeartbeatHandler
	interval             time.Duration
	enabled              bool
	mu                   sync.RWMutex
	started              bool
	stopChan             chan struct{}
}

func NewHeartbeatService(workspace string, onHeartbeat func(string) (string, error), intervalS int, enabled bool) *HeartbeatService {
	return &HeartbeatService{
		workspace:   workspace,
		onHeartbeat: onHeartbeat,
		interval:    time.Duration(intervalS) * time.Second,
		enabled:     enabled,
		stopChan:    make(chan struct{}),
	}
}

// SetOnHeartbeatWithTools sets the tool-supporting heartbeat handler.
// This handler returns a ToolResult that can indicate async operations.
// When set, this handler takes precedence over the legacy onHeartbeat callback.
func (hs *HeartbeatService) SetOnHeartbeatWithTools(handler HeartbeatHandler) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.onHeartbeatWithTools = handler
}

func (hs *HeartbeatService) Start() error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	if hs.started {
		return nil
	}

	if !hs.enabled {
		return fmt.Errorf("heartbeat service is disabled")
	}

	hs.started = true
	go hs.runLoop()

	return nil
}

func (hs *HeartbeatService) Stop() {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	if !hs.started {
		return
	}

	hs.started = false
	close(hs.stopChan)
}

func (hs *HeartbeatService) running() bool {
	select {
	case <-hs.stopChan:
		return false
	default:
		return true
	}
}

func (hs *HeartbeatService) runLoop() {
	ticker := time.NewTicker(hs.interval)
	defer ticker.Stop()

	for {
		select {
		case <-hs.stopChan:
			return
		case <-ticker.C:
			hs.checkHeartbeat()
		}
	}
}

func (hs *HeartbeatService) checkHeartbeat() {
	hs.mu.RLock()
	if !hs.enabled || !hs.running() {
		hs.mu.RUnlock()
		return
	}
	hs.mu.RUnlock()

	prompt := hs.buildPrompt()

	// Prefer the new tool-supporting handler
	if hs.onHeartbeatWithTools != nil {
		hs.executeHeartbeatWithTools(prompt)
	} else if hs.onHeartbeat != nil {
		_, err := hs.onHeartbeat(prompt)
		if err != nil {
			hs.log(fmt.Sprintf("Heartbeat error: %v", err))
		}
	}
}

// ExecuteHeartbeatWithTools executes a heartbeat using the tool-supporting handler.
// This method processes ToolResult returns and handles async tasks appropriately.
// If the result is async, it logs that the task started in background.
// If the result is an error, it logs the error message.
// This method is designed to be called from checkHeartbeat or directly by external code.
func (hs *HeartbeatService) ExecuteHeartbeatWithTools(prompt string) {
	hs.executeHeartbeatWithTools(prompt)
}

// executeHeartbeatWithTools is the internal implementation of tool-supporting heartbeat.
func (hs *HeartbeatService) executeHeartbeatWithTools(prompt string) {
	result := hs.onHeartbeatWithTools(prompt)

	if result == nil {
		hs.log("Heartbeat handler returned nil result")
		return
	}

	// Handle different result types
	if result.IsError {
		hs.log(fmt.Sprintf("Heartbeat error: %s", result.ForLLM))
		return
	}

	if result.Async {
		// Async task started - log and return immediately
		hs.log(fmt.Sprintf("Async task started: %s", result.ForLLM))
		logger.InfoCF("heartbeat", "Async heartbeat task started",
			map[string]interface{}{
				"message": result.ForLLM,
			})
		return
	}

	// Normal completion - log result
	hs.log(fmt.Sprintf("Heartbeat completed: %s", result.ForLLM))
}

func (hs *HeartbeatService) buildPrompt() string {
	notesDir := filepath.Join(hs.workspace, "memory")
	notesFile := filepath.Join(notesDir, "HEARTBEAT.md")

	var notes string
	if data, err := os.ReadFile(notesFile); err == nil {
		notes = string(data)
	}

	now := time.Now().Format("2006-01-02 15:04")

	prompt := fmt.Sprintf(`# Heartbeat Check

Current time: %s

Check if there are any tasks I should be aware of or actions I should take.
Review the memory file for any important updates or changes.
Be proactive in identifying potential issues or improvements.

%s
`, now, notes)

	return prompt
}

func (hs *HeartbeatService) log(message string) {
	logFile := filepath.Join(hs.workspace, "memory", "heartbeat.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(f, "[%s] %s\n", timestamp, message)
}
