package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sipeed/domeclaw/pkg/bus"
	"github.com/sipeed/domeclaw/pkg/config"
	"github.com/sipeed/domeclaw/pkg/logger"
)

// WebhookChannel handles incoming webhooks via HTTP POST
type WebhookChannel struct {
	*BaseChannel
	config    *config.WebhookConfig
	bus       *bus.MessageBus
	server    *http.Server
	running   bool
	manager   *Manager
	mu        sync.RWMutex
}

// WebhookRequest represents the incoming webhook payload
type WebhookRequest struct {
	Message   string            `json:"message"`
	ChatID    string            `json:"chat_id"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	AuthToken string            `json:"auth_token,omitempty"`
}

// WebhookResponse represents the response to the webhook caller
type WebhookResponse struct {
	Status    string `json:"status"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// NewWebhookChannel creates a new webhook channel
func NewWebhookChannel(cfg *config.WebhookConfig, bus *bus.MessageBus) *WebhookChannel {
	return &WebhookChannel{
		BaseChannel: NewBaseChannel("webhook", cfg, bus, nil),
		config:      cfg,
		bus:         bus,
		running:     false,
		manager:     nil, // Will be set later if needed
	}
}

// SetManager sets the channel manager for outbound sending
func (c *WebhookChannel) SetManager(m *Manager) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.manager = m
	logger.InfoC("webhook", "Channel manager attached to webhook")
}

// Start starts the webhook server
func (c *WebhookChannel) Start(ctx context.Context) error {
	if !c.config.Enabled {
		logger.InfoC("webhook", "Webhook channel disabled, skipping start")
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc(c.config.Path, c.handleWebhook)

	c.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", c.config.Host, c.config.Port),
		Handler: mux,
	}

	c.running = true
 logger.InfoCF("webhook", "Webhook server started",
		map[string]any{
			"address": c.server.Addr,
			"path":    c.config.Path,
		})

	// Start server in goroutine
	go func() {
		if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.ErrorCF("webhook", "Webhook server error", map[string]any{"error": err.Error()})
		}
	}()

	return nil
}

//Stop stops the webhook server
func (c *WebhookChannel) Stop(ctx context.Context) error {
	c.running = false
	if c.server != nil {
		if err := c.server.Shutdown(ctx); err != nil {
			logger.ErrorCF("webhook", "Webhook server shutdown error", map[string]any{"error": err.Error()})
			return err
		}
	}
	logger.InfoC("webhook", "Webhook server stopped")
	return nil
}

// Send sends a message via the webhook (not typically used for inbound)
func (c *WebhookChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	return fmt.Errorf("webhook channel is inbound-only")
}

// IsRunning returns whether the webhook server is running
func (c *WebhookChannel) IsRunning() bool {
	return c.running
}

// IsAllowed checks if the incoming request is allowed (token-based)
func (c *WebhookChannel) IsAllowed(senderID string) bool {
	// For webhooks, we use token validation instead of allowlist
	// The auth token is validated in the handler
	return true
}

// handleWebhook handles incoming webhook requests
func (c *WebhookChannel) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(WebhookResponse{
			Status: "error",
			Error:  "Method not allowed",
		})
		return
	}

	// Check authentication
	authToken := r.Header.Get("Authorization")
	if c.config.Token != "" {
		// Expect "Bearer <token>" format
		expected := "Bearer " + c.config.Token
		if authToken != expected {
			logger.WarnC("webhook", "Invalid or missing authorization token")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(WebhookResponse{
				Status: "error",
				Error:  "Invalid or missing authorization token",
			})
			return
		}
	}

	// Parse request body
	var req WebhookRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		logger.ErrorCF("webhook", "Failed to parse request body", map[string]any{"error": err.Error()})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(WebhookResponse{
			Status: "error",
			Error:  "Invalid JSON payload",
		})
		return
	}

	// Validate required fields
	if strings.TrimSpace(req.Message) == "" {
		logger.WarnC("webhook", "Empty message in webhook request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(WebhookResponse{
			Status: "error",
			Error:  "Message is required",
		})
		return
	}

	if strings.TrimSpace(req.ChatID) == "" {
		logger.WarnC("webhook", "Empty chat_id in webhook request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(WebhookResponse{
			Status: "error",
			Error:  "chat_id is required",
		})
		return
	}

	// Determine target channel from metadata or default to configured channel
	targetChannel := "telegram" // Default to telegram
	if req.Metadata != nil {
		if ch, ok := req.Metadata["target_channel"]; ok && ch != "" {
			targetChannel = ch
		}
	}

	// Validate target channel is enabled
	if !c.isChannelEnabled(targetChannel) {
		logger.ErrorCF("webhook", "Target channel not enabled", map[string]any{"channel": targetChannel})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(WebhookResponse{
			Status: "error",
			Error:  fmt.Sprintf("Target channel '%s' is not enabled", targetChannel),
		})
		return
	}

	// Generate a message ID
	messageID := fmt.Sprintf("webhook-%d", time.Now().UnixNano())

	// Create metadata with target channel info
	metadata := make(map[string]string)
	for k, v := range req.Metadata {
		metadata[k] = v
	}
	metadata["webhook_source"] = "true"

	// Publish to inbound message bus
	// Use the target channel as the channel so it gets routed correctly
	msg := bus.InboundMessage{
		Channel:    targetChannel, // Use the target channel for proper routing
		SenderID:   fmt.Sprintf("webhook|%s", req.ChatID),
		ChatID:     req.ChatID,
		Content:    req.Message,
		Metadata:   metadata,
		SessionKey: "",
	}

	c.bus.PublishInbound(msg)

	logger.InfoCF("webhook", "Webhook message processed",
		map[string]any{
			"chat_id":        req.ChatID,
			"target_channel": targetChannel,
			"message_len":    len(req.Message),
		})

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(WebhookResponse{
		Status:    "success",
		MessageID: messageID,
	})
}

// isChannelEnabled checks if a channel is enabled in the manager
func (c *WebhookChannel) isChannelEnabled(channelName string) bool {
	// This is a simplified check - in practice, you'd need access to the manager
	// For now, we'll allow known channels
	knownChannels := map[string]bool{
		"telegram":  true,
		"discord":   true,
		"whatsapp":  true,
		"feishu":    true,
		"dingtalk":  true,
		"slack":     true,
		"line":      true,
		"qq":        true,
		"wecom":     true,
		"wecom_app": true,
		"onebot":    true,
		"maixcam":   true,
	}
	return knownChannels[strings.ToLower(channelName)]
}
