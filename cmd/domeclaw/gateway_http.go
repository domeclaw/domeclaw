package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sipeed/domeclaw/pkg/agent"
	"github.com/sipeed/domeclaw/pkg/bus"
	"github.com/sipeed/domeclaw/pkg/config"
	"github.com/sipeed/domeclaw/pkg/logger"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// setupGatewayHTTP creates an HTTP server for the gateway API endpoints
func setupGatewayHTTP(cfg *config.Config, msgBus *bus.MessageBus, agentLoop *agent.AgentLoop) *http.Server {
	mux := http.NewServeMux()

	// Health endpoints (keep existing ones)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"service": "domeclaw-gateway",
		})
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ready",
			"service": "domeclaw-gateway",
		})
	})

	// New: Chat endpoint - POST /chat
	mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Message string `json:"message"`
			ChatID  string `json:"chat_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.WarnCF("gateway", "Invalid JSON in chat request", map[string]any{"error": err.Error()})
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.Message == "" {
			logger.WarnC("gateway", "Message is required in chat request")
			http.Error(w, "Message is required", http.StatusBadRequest)
			return
		}

		// Use default chat_id if not provided
		if req.ChatID == "" {
			req.ChatID = "curl_user"
		}

		// Process message with agent
		response, err := agentLoop.ProcessDirectWithChannel(
			context.Background(),
			req.Message,
			fmt.Sprintf("agent:curl:%s", req.ChatID),
			"curl",
			req.ChatID,
		)

		if err != nil {
			logger.ErrorCF("gateway", "Failed to process chat message", map[string]any{"error": err.Error()})
			http.Error(w, "Processing failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"response": response,
		})

		logger.InfoCF("gateway", "Chat request processed", map[string]any{
			"chat_id": req.ChatID,
			"response_preview": response[:min(len(response), 50)],
		})
	})

	// New: Webhook endpoint - POST /webhook (deprecated in favor of webhook channel)
	mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			Event    string            `json:"event"`
			Message  string            `json:"message"`
			SenderID string            `json:"sender_id"`
			ChatID   string            `json:"chat_id"`
			Metadata map[string]string `json:"metadata,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			logger.WarnCF("gateway", "Invalid JSON in webhook request", map[string]any{"error": err.Error()})
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if payload.Message == "" {
			logger.WarnC("gateway", "Message is required in webhook request")
			http.Error(w, "Message is required", http.StatusBadRequest)
			return
		}

		// Use defaults if not provided
		if payload.SenderID == "" {
			payload.SenderID = "webhook"
		}
		if payload.ChatID == "" {
			payload.ChatID = "webhook_chat"
		}

		// Publish message to bus
		msgBus.PublishInbound(bus.InboundMessage{
			Channel:  "webhook",
			SenderID: payload.SenderID,
			ChatID:   payload.ChatID,
			Content:  payload.Message,
			Metadata: payload.Metadata,
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "message queued",
		})

		logger.InfoCF("gateway", "Webhook message queued", map[string]any{
			"sender_id": payload.SenderID,
			"chat_id": payload.ChatID,
		})
	})

	addr := fmt.Sprintf("%s:%d", cfg.Gateway.Host, cfg.Gateway.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server
}
