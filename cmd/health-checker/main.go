package main

import (
	"fmt"
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
		fmt.Printf("<ERROR> Error discovering nodes: %v", err)
		os.Exit(1)
	}
	log.Println("<INFO> Node discovery completed.")
	log.Printf("Discovered nodes: %+v", nodes)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll("/app/output", 0755); err != nil {
		log.Printf("<ERROR> Failed to create output directory: %v", err)
		fmt.Printf("<ERROR> Failed to create output directory: %v", err)
		os.Exit(1)
	}

	// Pinging the nodes
	err = ping.Ping(&nodes)
	if err != nil {
		log.Printf("<ERROR> Error pinging nodes: %v", err)
		fmt.Printf("<ERROR> Error pinging nodes: %v", err)
		os.Exit(1)
	}
	log.Println("<INFO> Ping check completed.")

}
