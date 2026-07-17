package main

import (
	discovery "kaushik-chhappaniya/distributed-network-health-checker/internal/discovery"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/ping"
	logs "kaushik-chhappaniya/distributed-network-health-checker/utils/logger"
	"os"
)

const IMAGE_NAME = "alpine"

func main() {
	defer func() {
		if r := recover(); r != nil {
			println("Error:", r)
		}
	}()

	log, logComplete, err := logs.New("App.log", true)
	defer logComplete()
	if err != nil {
		panic("Logger instance starting failed.")
	}
	log.Println("<INFO> Application started.")
	log.Println("<INFO> Starting node discovery.")

	var nodes discovery.Nodes
	err = nodes.Discovery(IMAGE_NAME)
	if err != nil {
		log.Printf("<ERROR> Error discovering nodes: %v", err)
		os.Exit(1)
	}
	log.Println("<INFO> Node discovery completed.")
	log.Printf("Discovered nodes: %+v", nodes)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll("/app/output", 0755); err != nil {
		log.Printf("<ERROR> Failed to create output directory: %v", err)
		os.Exit(1)
	}

	// Pinging the nodes
	err = ping.Ping_Nodes(&nodes)
	if err != nil {
		log.Printf("<ERROR> Error pinging nodes: %v", err)
		os.Exit(1)
	}
	log.Println("<INFO> Ping nodes check completed.")

	err = ping.Ping_External(&nodes)
	if err != nil {
		log.Printf("<ERROR> Error pinging external target: %v", err)
		os.Exit(1)
	}
	log.Println("<INFO> Ping external check completed.")
	log.Println("<INFO> Application completed successfully.")

}
