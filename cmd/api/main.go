package main

import (
	"log"
	"net/http"

	"kaushik-chhappaniya/distributed-network-health-checker/internal/api/handler"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/api/router"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/app"
)

func main() {
	// Initialize app instance that initializes all db connections, loggers, and services
	app, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	defer app.Close()

	// Initialize HTTP handlers with their dependencies
	healthHandler := handler.NewHealthHandler(
		app.Logger,
		app.DBService,
	)

	discoveryHandler := handler.NewDiscoveryHandler(
		app.Logger,
		app.DiscoveryService,
		app.DBService,
	)

	pingHandler := handler.NewPingHandler(
		app.Logger,
		app.PingService,
	)

	dbHandler := handler.NewDbHandler(
		app.Logger,
		app.DBService,
	)

	// Build the HTTP router with registered routes
	mux := router.New(healthHandler, discoveryHandler, pingHandler, dbHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	app.Logger.Println("API server running on :8080")

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		app.Logger.Fatal(err)
	}

	log.Println("<INFO> Application completed successfully.")
}

