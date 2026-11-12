package ping

import (
	"fmt"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/models"
	pinghelper "kaushik-chhappaniya/distributed-network-health-checker/utils/ping_helper"
	"sync"
)

/*
worker implements the consumer side of the worker pool pattern. It continuously
consumes tasks (nodes) from the pings channel until the channel is closed.
It decouples the management of goroutines from the actual business logic of pinging
*/
func worker(pings <-chan PingTarget, logChan chan<- string, resultsChan chan<- models.PingMetric, wg *sync.WaitGroup) {
	defer wg.Done()
	logChan <- fmt.Sprintf("<WORKER> Inside worker. worker started.")
	// The range loop blocks here (waiting for work) until the channel closes
	for ping := range pings {
		// Ping helper returns the parsed PingMetric model from TextFSM
		metric, err := pinghelper.PingHelper(ping.Name, logChan, ping.Target)
		if err != nil {
			logChan <- fmt.Sprintf("<ERROR> Error parsing ping results from node: %s : %v", ping.Name, err)
		}
		if metric != nil {
			resultsChan <- *metric
		}
	}
}
