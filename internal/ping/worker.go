package ping

import (
	"fmt"
	pinghelper "kaushik-chhappaniya/distributed-network-health-checker/utils/ping_helper"
	"sync"
)

/*
worker implements the consumer side of the worker pool pattern. It continuously
consumes tasks (nodes) from the pings channel until the channel is closed.
It decouples the management of goroutines from the actual business logic of pinging
*/
func worker(pings <-chan PingTarget, logChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	logChan <- fmt.Sprintf("<WORKER> Inside worker. worker started. %v", pings)
	// The range loop blocks here (waiting for work) until the channel closes
	for ping := range pings {
		// Ping helper doesnt need file-name, log-file as it is handled by the loggerGoroutine. It just needs the node name and target to ping.
		pinghelper.PingHelper(ping.Name, logChan, ping.Target)
		// logChan <- fmt.Sprintf("<INFO> Finished pinging node: %s", ping.Name)
	}
}

// func (p *Pinger) worker(pings <-chan PingTarget, logChan chan<- string, wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	for ping := range pings {
// 		// Run system ping
// 		pinghelper.Ping_helper(ping.Name, logChan, ping.Target)
// 		// Directly access p.pool to insert metrics into DB!
// 		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 		_, err := p.pool.Exec(ctx,
// 			`INSERT INTO ping_metrics (target, status) VALUES ($1, $2)`,
// 			ping.Target, "SUCCESS",
// 		)
// 		cancel()
// 		if err != nil {
// 			logChan <- fmt.Sprintf("<ERROR> Failed to write metric to DB for %s: %v", ping.Target, err)
// 		}
// 	}
