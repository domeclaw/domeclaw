package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/channels"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
)

// WebhookChannel handles incoming webhooks via HTTP POST and broadcasts to other channels
type WebhookChannel struct {
	*channels.BaseChannel
	cfg     config.WebhookConfig
	bus     *bus.MessageBus
	server  *http.Server
	running bool
}

// WebhookRequest represents the incoming webhook payload
type WebhookRequest struct {
	Message  string            `json:"message"`
	ChatID   string            `json:"chat_id"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// WebhookResponse represents the response to the webhook caller
type WebhookResponse struct {
	Status    string `json:"status"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// NewWebhookChannel creates a new webhook channel
func NewWebhookChannel(cfg config.WebhookConfig, messageBus *bus.MessageBus) *WebhookChannel {
	return &WebhookChannel{
		BaseChannel: channels.NewBaseChannel("webhook", cfg, messageBus, nil),
		cfg:         cfg,
		bus:         messageBus,
		running:     false,
	}
}

// Start starts the webhook HTTP server
func (c *WebhookChannel) Start(ctx context.Context) error {
	if !c.cfg.Enabled {
		logger.InfoC("webhook", "Webhook channel disabled, skipping start")
		return nil
	}

	mux := http.NewServeMux()
	path := "/webhook"
	if c.cfg.Path != "" {
		path = c.cfg.Path
	}
	mux.HandleFunc(path, c.handleWebhook)

	c.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port),
		Handler: mux,
	}

	c.running = true
	c.SetRunning(true)

	logger.InfoCF("webhook", "Webhook server started",
		map[string]any{
			"address": c.server.Addr,
			"path":    path,
		})

	// Start server in goroutine
	go func() {
		if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.ErrorCF("webhook", "Webhook server error", map[string]any{"error": err.Error()})
		}
	}()

	return nil
}

// Stop stops the webhook server
func (c *WebhookChannel) Stop(ctx context.Context) error {
	c.running = false
	c.SetRunning(false)
	if c.server != nil {
		if err := c.server.Shutdown(ctx); err != nil {
			logger.ErrorCF("webhook", "Webhook server shutdown error", map[string]any{"error": err.Error()})
			return err
		}
	}
	logger.InfoC("webhook", "Webhook server stopped")
	return nil
}

// Send is not implemented for webhook channel (inbound-only)
func (c *WebhookChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	return fmt.Errorf("webhook channel is inbound-only")
}

// IsRunning returns whether the webhook server is running
func (c *WebhookChannel) IsRunning() bool {
	return c.running
}

// IsAllowed webhook uses token auth instead of allowlist
func (c *WebhookChannel) IsAllowed(senderID string) bool {
	return true
}

// handleWebhook processes incoming POST requests
func (c *WebhookChannel) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Check authentication
	if c.cfg.Token != "" {
		authHeader := r.Header.Get("Authorization")
		expected := "Bearer " + c.cfg.Token
		if authHeader != expected {
			logger.WarnC("webhook", "Invalid or missing authorization token")
			respondWithError(w, http.StatusUnauthorized, "Invalid or missing authorization token")
			return
		}
	}

	// Parse request body
	var req WebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.ErrorCF("webhook", "Failed to parse request body", map[string]any{"error": err.Error()})
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate required fields
	if strings.TrimSpace(req.Message) == "" {
		logger.WarnC("webhook", "Empty message in webhook request")
		respondWithError(w, http.StatusBadRequest, "Message is required")
		return
	}

	if strings.TrimSpace(req.ChatID) == "" {
		logger.WarnC("webhook", "Empty chat_id in webhook request")
		respondWithError(w, http.StatusBadRequest, "chat_id is required")
		return
	}

	// Determine target channel from metadata (default: telegram)
	targetChannel := "telegram"
	if req.Metadata != nil {
		if ch, ok := req.Metadata["target_channel"]; ok && ch != "" {
			targetChannel = ch
		}
	}

	// Generate message ID
	messageID := fmt.Sprintf("webhook-%d", time.Now().UnixNano())

	// Prepare metadata
	metadata := make(map[string]string)
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			metadata[k] = v
		}
	}
	metadata["webhook_source"] = "true"
	metadata["original_chat_id"] = req.ChatID

	// Create inbound message
	msg := bus.InboundMessage{
		Channel:   targetChannel,
		SenderID:  fmt.Sprintf("webhook|%s", req.ChatID),
		ChatID:    req.ChatID,
		Content:   req.Message,
		Metadata:  metadata,
		MessageID: messageID,
	}

	// Publish to message bus
	if err := c.bus.PublishInbound(r.Context(), msg); err != nil {
		logger.ErrorCF("webhook", "Failed to publish message", map[string]any{"error": err.Error()})
		respondWithError(w, http.StatusInternalServerError, "Failed to process message")
		return
	}

	logger.InfoCF("webhook", "Webhook message processed",
		map[string]any{
			"chat_id":        req.ChatID,
			"target_channel": targetChannel,
			"message_len":    len(req.Message),
		})

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(WebhookResponse{
		Status:    "success",
		MessageID: messageID,
	})
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, status int, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(WebhookResponse{
		Status: "error",
		Error:  errMsg,
	})
}
