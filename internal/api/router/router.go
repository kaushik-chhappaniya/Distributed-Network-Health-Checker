package router

import (
	"net/http"

	"kaushik-chhappaniya/distributed-network-health-checker/internal/api/docs"
	handler "kaushik-chhappaniya/distributed-network-health-checker/internal/api/handler"
)

// New initializes and configures the HTTP request multiplexer with all application routes.
func New(
	healthHandler *handler.HealthHandler,
	discoveryHandler *handler.DiscoveryHandler,
	pingHandler *handler.PingHandler,
	dbHandler *handler.DbHandler,
) *http.ServeMux {
	mux := http.NewServeMux()

	// 0. Interactive API Documentation (Swagger UI)
	mux.HandleFunc("GET /docs", docs.ServeSwaggerUI)
	mux.HandleFunc("GET /openapi.json", docs.ServeOpenAPISpec)

	// 1. Default Health Check routes (Go 1.22+ "GET /{$}" strictly matches root "/")
	mux.HandleFunc("GET /{$}", healthHandler.HealthCheck)
	mux.HandleFunc("GET /health", healthHandler.HealthCheck)

	// 2. Inventory / Database routes
	mux.HandleFunc("GET /api/v1/nodes", dbHandler.GetNodes)

	// 3. Discovery routes
	mux.HandleFunc("POST /api/v1/discovery", discoveryHandler.TriggerDiscovery)

	// 4. Health check / Ping routes
	mux.HandleFunc("GET /api/v1/metrics", pingHandler.GetPingMetrics)
	mux.HandleFunc("GET /api/v1/ping-metrics", pingHandler.GetPingMetrics)
	mux.HandleFunc("POST /api/v1/ping", pingHandler.TriggerPing)
	mux.HandleFunc("POST /api/v1/ping-external", pingHandler.TriggerExternalPing)

	return mux
}
