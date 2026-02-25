package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sipeed/domeclaw/pkg/bus"
	"github.com/sipeed/domeclaw/pkg/config"
	"github.com/sipeed/domeclaw/pkg/logger"
)

// WebhookConfig holds configuration for the webhook channel.
type WebhookConfig struct {
	Enabled bool   `json:"enabled"`
	Token   string `json:"token"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
}

// DefaultWebhookConfig returns the default webhook configuration.
func DefaultWebhookConfig() WebhookConfig {
	return WebhookConfig{
		Enabled: false,
		Token:   "",
		Host:    "localhost",
		Port:    18795,
	}
}

// WebhookChannel implements the Channel interface for receiving messages via HTTP webhook.
type WebhookChannel struct {
	*BaseChannel
	config     config.WebhookConfig
	httpServer *http.Server
	serverMu   sync.Mutex
	running    bool
	runningMu  sync.RWMutex
}

// NewWebhookChannel creates a new webhook channel instance.
func NewWebhookChannel(cfg config.WebhookConfig, messageBus *bus.MessageBus) (*WebhookChannel, error) {
	if cfg.Token == "" {
		return nil, fmt.Errorf("webhook token is required")
	}

	// Convert FlexibleStringSlice to []string
	allowList := []string(cfg.AllowFrom)

	base := NewBaseChannel("webhook", cfg, messageBus, allowList)

	return &WebhookChannel{
		BaseChannel: base,
		config:      cfg,
	}, nil
}

// Start launches the HTTP webhook server.
func (c *WebhookChannel) Start(ctx context.Context) error {
	logger.InfoC("webhook", "Starting webhook channel")

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health/webhook", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"channel": "webhook",
		})
	})

	// Receive webhook endpoint
	mux.HandleFunc("/webhook", c.handleWebhook)

	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	c.serverMu.Lock()
	c.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	c.serverMu.Unlock()

	go func() {
		logger.InfoCF("webhook", "Webhook server listening", map[string]any{
			"addr": addr,
		})
		if err := c.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.ErrorCF("webhook", "Webhook server error", map[string]any{
				"error": err.Error(),
			})
		}
	}()

	c.runningMu.Lock()
	c.running = true
	c.runningMu.Unlock()

	logger.InfoC("webhook", "Webhook channel started")
	return nil
}

// handleWebhook handles incoming webhook requests.
func (c *WebhookChannel) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate with token
	token := r.Header.Get("Authorization")
	if token == "" {
		token = r.URL.Query().Get("token")
	}

	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	if !c.verifyToken(token) {
		logger.WarnC("webhook", "Unauthorized webhook request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.ErrorCF("webhook", "Failed to read request body", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var payload struct {
		Message  string            `json:"message"`
		SenderID string            `json:"sender_id"`
		ChatID   string            `json:"chat_id"`
		Metadata map[string]string `json:"metadata,omitempty"`
		Target   struct {
			Channel string `json:"channel"`
			ChatID  string `json:"chat_id"`
		} `json:"target,omitempty"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		logger.ErrorCF("webhook", "Failed to parse webhook payload", map[string]any{
			"error": err.Error(),
		})
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if payload.Message == "" {
		logger.ErrorC("webhook", "Message is required")
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	// Use default values if not provided
	if payload.SenderID == "" {
		payload.SenderID = "webhook_user"
	}
	if payload.ChatID == "" {
		payload.ChatID = "webhook_chat"
	}

	// Publish message to bus
	c.HandleMessage(payload.SenderID, payload.ChatID, payload.Message, nil, payload.Metadata)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "message received",
	})
}

// verifyToken validates the authentication token.
func (c *WebhookChannel) verifyToken(token string) bool {
	if c.config.Token == "" {
		return true // No token required
	}
	return token == "Bearer "+c.config.Token
}

// Stop gracefully shuts down the HTTP server.
func (c *WebhookChannel) Stop(ctx context.Context) error {
	logger.InfoC("webhook", "Stopping webhook channel")

	c.serverMu.Lock()
	if c.httpServer != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := c.httpServer.Shutdown(shutdownCtx); err != nil {
			logger.ErrorCF("webhook", "Webhook server shutdown error", map[string]any{
				"error": err.Error(),
			})
		}
	}
	c.serverMu.Unlock()

	c.runningMu.Lock()
	c.running = false
	c.runningMu.Unlock()

	logger.InfoC("webhook", "Webhook channel stopped")
	return nil
}

// Send is not implemented for webhook channel (it's inbound only).
func (c *WebhookChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	// Webhook channel doesn't send messages, it only receives them
	logger.DebugCF("webhook", "Ignoring outbound message (webhook is inbound only)", map[string]any{
		"channel": msg.Channel,
		"chat_id": msg.ChatID,
	})
	return nil
}

// IsRunning returns whether the channel is currently running.
func (c *WebhookChannel) IsRunning() bool {
	c.runningMu.RLock()
	defer c.runningMu.RUnlock()
	return c.running
}
