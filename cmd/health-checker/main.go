package main

import (
	"os"

	db "kaushik-chhappaniya/distributed-network-health-checker/internal/db"
	discovery "kaushik-chhappaniya/distributed-network-health-checker/internal/discovery"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/models"
	ping "kaushik-chhappaniya/distributed-network-health-checker/internal/ping"
	logs "kaushik-chhappaniya/distributed-network-health-checker/utils/logger"
)

const IMAGE_NAME = "alpine"

func main() {
	defer func() {
		if r := recover(); r != nil {
			println("Error:", r)
		}
	}()

	// Start the logger instance for the App
	log, logComplete, err := logs.New("App.log", true)
	defer logComplete()
	if err != nil {
		panic("Logger instance starting failed.")
	}
	log.Println("<INFO> Application started.")
	log.Println("<INFO> Starting node discovery.")

	// Connect the DB
	pool, err := db.Connect()
	if err != nil {
		log.Printf("Error while starting database: %v", err.Error())
		panic("Error while starting databse")
	}

	defer pool.Close()

	// Start the node discovery. Using var to declare variable but not allocate memory to it
	// Ensure DB schema exists
	if _, err := db.CreateModels(pool, log); err != nil {
		log.Printf("<WARN> Failed to verify schema: %v", err)
	}

	// ------------------- 1. Node Discovery -------------------
	var nodes models.Nodes
	if err := discovery.Discover(IMAGE_NAME, &nodes); err != nil {
		log.Printf("<ERROR> Error discovering nodes: %v", err)
		os.Exit(1)
	}

	log.Println("<INFO> Node discovery completed.")
	log.Printf("Discovered nodes: %+v", nodes)

	// Adding nodes in Postgres database using pool
	// dbNodes := db.New(nodes, pool)
	// for _, n := range nodes {
	// 	dbNode := &db.Node{
	// 		Name: n.Name,
	// 	}
	// 	if _, err := db.InsertNode(pool, log, dbNode); err != nil {
	// 		log.Printf("<WARN> Failed to insert node %s into DB: %v", n.Name, err)
	// 	}
	// Save discovered nodes into PostgreSQL

	// ----------------------- 2. Pushing Nodes data into the PostgreSQL database -----------------
	dbService := db.New(pool)
	if err := dbService.InsertMultipleNodes(log, nodes); err != nil {
		log.Printf("<WARN> Failed to insert nodes into DB: %v", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll("/app/output", 0755); err != nil {
		log.Printf("<ERROR> Failed to create output directory: %v", err)
		os.Exit(1)
	}

	// ------------------- 3a. Pinging the Internal Nodes -------------------
	// Creating a struct with pool and worker count
	pinger := ping.NewPinger(pool)
	// Pinging nodes
	err = pinger.PingNodes(&nodes)
	if err != nil {
		log.Printf("<ERROR> Error pinging nodes: %v", err)
		os.Exit(1)
	}
	log.Println("<INFO> Ping nodes check completed.")

	// ------------------- 3b. Pinging the External Nodes -------------------
	// Ping external targets from the nodes
	err = pinger.PingExternal(&nodes)
	if err != nil {
		log.Printf("<ERROR> Error pinging external target: %v", err)
		os.Exit(1)
	}
	log.Println("<INFO> Ping external check completed.")

	log.Println("<INFO> Application completed successfully.")
}
