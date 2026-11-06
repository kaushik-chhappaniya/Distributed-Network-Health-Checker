package main

import (
	"log"
	"os"

	db "kaushik-chhappaniya/distributed-network-health-checker/internal/db"
	discovery "kaushik-chhappaniya/distributed-network-health-checker/internal/discovery"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/mappers"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/models"
	logs "kaushik-chhappaniya/distributed-network-health-checker/utils/logger"
)

const IMAGE_NAME = "alpine"

func main() {
	defer func() {
		if r := recover(); r != nil {
			println("Error:", r)
		}
	}()

	// ------------------- 1. Create All Dependancy Injection Services -------------------
	// Start the logger instance for the App
	appLog, logComplete, err := logs.New("App.log", true)
	defer logComplete()
	if err != nil {
		panic("Logger instance starting failed.")
	}

	// Start the logger instance for discovery service
	discoveryLog, discvoeryLogComplete, discoveryErr := logs.New("Discovery.log", true)
	defer discvoeryLogComplete()
	if discoveryErr != nil {
		panic("Logger instance starting failed.")
	}
	discoveryService := discovery.New(discoveryLog)

	// Start the logger instance for db service
	dbLog, dbLogComplete, dbErr := logs.New("Db.log", true)
	defer dbLogComplete()
	if dbErr != nil {
		panic("Logger instance starting failed.")
	}
	// Connect the DB
	pool, err := db.Connect(appLog)
	if err != nil {
		appLog.Printf("Error while starting database: %v", err.Error())
		panic("Error while starting databse")
	}
	defer pool.Close()
	dbService := db.New(pool, dbLog)
	// Create db models
	helloGreeting, err := dbService.CreateModels()
	appLog.Println("<INFO> Database table created successfully.", helloGreeting)
	// Ensure DB schema exists
	if _, err := dbService.CreateModels(); err != nil {
		appLog.Printf("<WARN> Failed to verify schema: %v", err)
	}

	// ------------------- 1. Node Discovery -------------------
	// we create a nodes variable to add in the struct that we'll use to create discovery service
	var nodes discovery.ClabNodes
	// ------------------- 1.1 Topology Discovery -------------------
	appLog.Printf("Discover topology file.")
	topologyFile, err := discoveryService.DiscoverTopology("./lab")
	if err != nil {
		appLog.Println("<ERROR> No topology json file found.")
		panic("No Topology file found.")
	}
	// TODO: Add automating script that will automatically move to clab nodes
	//  if img not present in ENV or not provided by user. Moving with the Alpine Nodes discovery.
	appLog.Println("<INFO> Found topology file: ", topologyFile)
	discoveryConfig := discovery.DiscoveryData{
		TopologyFile: &topologyFile,
		Nodes:        nodes,
	}

	// ------------------- 1.2 Nodes Discovery -------------------
	// Discover the nodes here. Based on the image provided or clab topology provided we will
	if err := discoveryService.Discover(&discoveryConfig); err != nil {
		appLog.Printf("<ERROR> Error discovering nodes: %v", err)
		os.Exit(1)
	}

	appLog.Println("<INFO> Node discovered from topology files.")
	appLog.Printf("Discovered nodes: %+v", discoveryConfig.Nodes)

	// ----------------------- 2. Pushing Nodes data into the PostgreSQL database -----------------

	// ----------------------- 2.1 Map the containerLab nodes with the database nodes struct -----------------
	var dbNodes models.Nodes
	// Here we iterate over the received nodes and map them with the database struct.
	// TODO: Add scaling architechture here. workers can be added but assigning is cheap task. No big performance change.
	for _, n := range discoveryConfig.Nodes {
		dbNode, err := mappers.MapClabNodeToDBNode(&n)
		if err != nil {
			appLog.Printf("<ERROR> Error in mapping nodes: %V", err)
		}
		dbNodes = append(dbNodes, *dbNode)
	}
	appLog.Printf("<INFO> Clab nodes mapped to db nodes struct.")
	appLog.Println("dbNodes: ", dbNodes)
	// After mapping nodes to db struct. Push nodes into DB.
	if err := dbService.InsertMultipleNodes(dbNodes); err != nil {
		log.Printf("<WARN> Failed to insert nodes into DB: %v", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll("/app/output", 0755); err != nil {
		log.Printf("<ERROR> Failed to create output directory: %v", err)
		os.Exit(1)
	}

	// // ------------------- 3a. Pinging the Internal Nodes -------------------
	// // Creating a struct with pool and worker count
	// pinger := ping.NewPinger(pool)
	// // Pinging nodes
	// err = pinger.PingNodes(&nodes)
	// if err != nil {
	// 	log.Printf("<ERROR> Error pinging nodes: %v", err)
	// 	os.Exit(1)
	// }
	// log.Println("<INFO> Ping nodes check completed.")

	// // ------------------- 3b. Pinging the External Nodes -------------------
	// // Ping external targets from the nodes
	// err = pinger.PingExternal(&nodes)
	// if err != nil {
	// 	log.Printf("<ERROR> Error pinging external target: %v", err)
	// 	os.Exit(1)
	// }
	// log.Println("<INFO> Ping external check completed.")

	log.Println("<INFO> Application completed successfully.")
}
