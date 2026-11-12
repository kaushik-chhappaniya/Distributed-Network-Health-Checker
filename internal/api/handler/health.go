package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"kaushik-chhappaniya/distributed-network-health-checker/internal/db"
)

type HealthHandler struct {
	logger    *log.Logger
	dbService *db.Service
}

func NewHealthHandler(
	logger *log.Logger,
	dbService *db.Service,
) *HealthHandler {
	return &HealthHandler{
		logger:    logger,
		dbService: dbService,
	}
}

// HealthCheck validates connectivity to critical dependencies and returns health status.
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	w.Header().Set("Content-Type", "application/json")

	// Verify database connection pool health
	if err := h.dbService.Ping(ctx); err != nil {
		h.logger.Printf("<WARN> Health check failed - database ping error: %v", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]any{
			"status":   "unhealthy",
			"database": "disconnected",
			"error":    err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status":   "healthy",
		"message":  "all services healthy status",
		"database": "connected",
	})
}

