package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"kaushik-chhappaniya/distributed-network-health-checker/internal/ping"
)

type PingHandler struct {
	logger      *log.Logger
	pingService *ping.Service
}

type ExternalPingRequest struct {
	Target string `json:"target"`
	URL    string `json:"url,omitempty"`
}

func NewPingHandler(
	logger *log.Logger,
	service *ping.Service,
) *PingHandler {
	return &PingHandler{
		logger:      logger,
		pingService: service,
	}
}

// GetPingMetrics retrieves all recorded ping metrics from the database.
func (h *PingHandler) GetPingMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.pingService.GetPingMetrics(r.Context())
	if err != nil {
		h.logger.Printf("<ERROR> Failed to retrieve ping metrics: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		h.logger.Printf("<ERROR> Failed to encode ping metrics to JSON: %v", err)
	}
}

// Trigger Ping to internal nodes.
func (h *PingHandler) TriggerPing(w http.ResponseWriter, r *http.Request) {
	// Derive a context with a 30-second deadline
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	// Frees timer resources as soon as TriggerPing exits
	defer cancel()

	// Pass ctx to the service
	err := h.pingService.PingNodes(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Internal nodes ping check completed",
	})
}

// Trigger external Ping - accepts target/url in request body
func (h *PingHandler) TriggerExternalPing(w http.ResponseWriter, r *http.Request) {
	var req ExternalPingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("<WARN> Invalid JSON body for external ping: %v", err)
		http.Error(w, "invalid request body: expected JSON with 'target' or 'url'", http.StatusBadRequest)
		return
	}

	target := req.Target
	if target == "" {
		target = req.URL
	}
	if target == "" {
		http.Error(w, "missing required field: 'target' or 'url' (e.g. {\"target\": \"google.com\"})", http.StatusBadRequest)
		return
	}

	// Derive a context with a 30-second deadline
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Pass ctx and target to the service
	err := h.pingService.PingExternal(ctx, target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": fmt.Sprintf("External ping check to %s completed", target),
		"target":  target,
	})
}
