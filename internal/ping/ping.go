package ping

import (
	"fmt"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/discovery"
	logs "kaushik-chhappaniya/distributed-network-health-checker/utils/logger"
	"sync"
)

const PING_FILE_NAME = "/app/output/ping_output.txt"

type PingTarget struct {
	Name   string
	Target string
}

/*
Ping orchestrates the ping utility lifecycle. It initializes the logging infrastructure,
sets up the worker pool, manages the communication channels, and ensures a graceful
shutdown of all concurrent components.
*/
func Ping_Nodes(nodes *discovery.Nodes) error {
	log, logComplete, err := logs.New("Ping.log", false)
	defer logComplete()

	if err != nil {
		return fmt.Errorf("<ERROR> Logger instance starting failed.")
	}

	// log.Println("<INFO> Started ping check.")

	pings := make(chan PingTarget, len(*nodes)) // Buffer is fine to be larger than data
	logChan := make(chan string, 100)

	var wg sync.WaitGroup
	var logWg sync.WaitGroup // For the logger

	// Starting the logger first
	logWg.Add(1)
	go logs.LoggerGoroutine(logChan, &logWg, log, PING_FILE_NAME)

	// Start 5 waitgroup workers
	for w := 1; w <= 5; w++ {
		wg.Add(1)
		// log.Printf("<WORKER> Starting ping worker %d\n", w)
		go worker(pings, logChan, &wg)
	}

	// Feeding ping data - send the node data to the pings channel and then close it to signal no more data
	for _, node := range *nodes {
		pings <- PingTarget{Name: node.Name, Target: node.Name}
	}
	close(pings) // Signal to workers that no more data will be sent

	wg.Wait()      // Wait for workers to finish processing
	close(logChan) // Close log channel to stop the logger loop
	logWg.Wait()   // Wait for logger to finish writing remaining logs
	// log.Printf("<INFO> Ping check completed and output stored at %s\n", PING_FILE_NAME)
	return nil
}
