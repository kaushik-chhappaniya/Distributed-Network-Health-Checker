package handler

import (
	"encoding/json"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/db"
	"log"
	"net/http"
)

type DbHandler struct {
	logger    *log.Logger
	dbService *db.Service
}

func NewDbHandler(
	logger *log.Logger,
	service *db.Service,
) *DbHandler {
	return &DbHandler{
		logger:    logger,
		dbService: service,
	}
}

func (h *DbHandler) GetNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.dbService.GetAllNodes(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(nodes); err != nil {
		h.logger.Printf("<ERROR> Failed to encode nodes to JSON: %v", err)
	}
}
