package handler

import (
	"encoding/json"
	"fmt"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/db"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/discovery"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/mappers"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/models"
	"log"
	"net/http"
)

// Handler for discovery func
type DiscoveryHandler struct {
	logger           *log.Logger
	discoveryService *discovery.Service
	dbService        *db.Service
}

func NewDiscoveryHandler(
	logger *log.Logger,
	service *discovery.Service,
	dbService *db.Service,
) *DiscoveryHandler {
	return &DiscoveryHandler{
		logger:           logger,
		discoveryService: service,
		dbService:        dbService,
	}
}

func (h *DiscoveryHandler) TriggerDiscovery(w http.ResponseWriter, r *http.Request) {
	// Find topology file and parse nodes
	h.logger.Printf("<INFO> Discovering topology file.")
	topologyFile, err := h.discoveryService.DiscoverTopology("./lab")
	if err != nil {
		h.logger.Printf("<ERROR> No topology json file found: %v", err)
		http.Error(w, fmt.Sprintf("topology file not found: %v", err), http.StatusNotFound)
		return
	}
	h.logger.Printf("<INFO> Found topology file: %s", topologyFile)
	discoveryConfig := discovery.DiscoveryData{
		TopologyFile: &topologyFile,
		Nodes:        discovery.ClabNodes{},
	}

	// Discover the nodes here. Based on the image provided or clab topology provided.
	if err := h.discoveryService.Discover(&discoveryConfig); err != nil {
		h.logger.Printf("<ERROR> Error discovering nodes: %v", err)
		http.Error(w, fmt.Sprintf("discovery failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Map the clab nodes to db nodes struct
	var dbNodes models.Nodes
	for _, n := range discoveryConfig.Nodes {
		dbNode, err := mappers.MapClabNodeToDBNode(&n)
		if err != nil {
			h.logger.Printf("<ERROR> Error in mapping nodes: %V", err)
		}
		dbNodes = append(dbNodes, *dbNode)
	}
	h.logger.Printf("<INFO> Clab nodes mapped to db nodes struct.")

	// Insert db nodes into the database
	if err := h.dbService.InsertMultipleNodes(dbNodes); err != nil {
		log.Printf("<WARN> Failed to insert nodes into DB: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"count":  len(dbNodes),
	})
}
