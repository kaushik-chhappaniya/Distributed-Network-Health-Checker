package ping

import (
	"fmt"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/discovery"
	logs "kaushik-chhappaniya/distributed-network-health-checker/utils/logger"
	"sync"
)

const PING_EXT_FILE_NAME = "/app/output/ping_external_output.txt"
const PING_TARGET = "google.com" // External target for pinging

/*
Ping orchestrates the ping utility lifecycle. It initializes the logging infrastructure,
sets up the worker pool, manages the communication channels, and ensures a graceful
shutdown of all concurrent components.
*/
func Ping_External(nodes *discovery.Nodes) error {
	log, logComplete, err := logs.New("Ping_External.log", false)
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
	go logs.LoggerGoroutine(logChan, &logWg, log, PING_EXT_FILE_NAME)

	// Start 5 waitgroup workers
	for w := 1; w <= 5; w++ {
		wg.Add(1)
		// log.Printf("<WORKER> Starting ping worker %d\n", w)
		go worker(pings, logChan, &wg)
	}

	// Feeding ping data - send the node data to the pings channel and then close it to signal no more data
	for _, node := range *nodes {
		pings <- PingTarget{Name: node.Name, Target: PING_TARGET}
	}
	close(pings) // Signal to workers that no more data will be sent

	wg.Wait()      // Wait for workers to finish processing
	close(logChan) // Close log channel to stop the logger loop
	logWg.Wait()   // Wait for logger to finish writing remaining logs
	// log.Printf("<INFO> Ping check completed and output stored at %s\n", PING_FILE_NAME)
	return nil
}

/*
runLogProcessor acts as the centralized I/O manager. It consumes messages from the logChan
and routes them: status logs go to the primary logger, while raw ping output is written
to the output file. This function is the "Exclusive Writer," preventing race conditions
by ensuring no other goroutine touches the output file.
*/
// func loggerGoroutine(logChan <-chan string, logWg *sync.WaitGroup, log *log.Logger) {

// 	defer logWg.Done()
// 	log.Println("<WORKER> Inside loggerGoroutine. worker started.")
// 	// Open the ping output file for writing
// 	// if _, err := file.Write(fmt.Appendf(nil, "%s : %s \n\n", node.Name, output)); err != nil {
// 	// if _, err := fmt.Fprintf(file, fmt.Sprintf("%s : %s \n\n", node.Name, output)); err != nil {
// 	file, err := os.OpenFile(PING_FILE_NAME, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	if err != nil {
// 		log.Println("<ERROR> Could not open ping output file")
// 		return
// 	}
// 	defer file.Close()

// 	for msg := range logChan {
// 		if strings.HasPrefix(msg, "<ERROR>") || strings.HasPrefix(msg, "<INFO>") || strings.HasPrefix(msg, "<WORKER>") {
// 			log.Printf("%s\n", msg)
// 			// fmt.Fprintf(file, "%s\n", msg)
// 		} else {
// 			fmt.Fprintf(file, msg)
// 		}
// 	}
// }

/*
worker implements the consumer side of the worker pool pattern. It continuously
consumes tasks (nodes) from the pings channel until the channel is closed.
It decouples the management of goroutines from the actual business logic of pinging
*/
// func worker(pings <-chan discovery.Node, logChan chan<- string, wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	logChan <- fmt.Sprintf("<WORKER> Inside worker. worker started. %v", pings)
// 	// The range loop blocks here (waiting for work) until the channel closes
// 	for ping := range pings {
// 		ping_helper(ping, logChan)
// 		// logChan <- fmt.Sprintf("<INFO> Finished pinging node: %s", ping.Name)
// 	}
// }

/*
ping_helper executes the system ping command for a given node. It does not
handle concurrency or file I/O itself; it relies on the logChan to report
execution status, errors, and raw output back to the central logger.
*/
// func ping_helper(node discovery.Node, logChan chan<- string) {
// 	PING_CMD := "ping -c 5 -q -s 20 -w 10 %s"
// 	// log.Printf("<INFO> Pinging node: %s with IP: %s", node.Name, node.IP)
// 	logChan <- fmt.Sprintf("<INFO> Pinging node: %s with IP: %s", node.Name, node.IP)

// 	// Create context with 10 seconds timeout for the ping command
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel() // Cancel the context to free resources if the command finishes before timeout

// 	// Execute the ping command using exec.CommandContext to allow for timeout
// 	cmd := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf(PING_CMD, node.Name))
// 	output, err := cmd.Output()

// 	// output, err := exec.Command("sh", "-c", cmd).Output()
// 	if err != nil {
// 		logChan <- fmt.Sprintf("<ERROR> Error pinging node %s: %v", node.Name, err)
// 		err = fmt.Errorf("%v; <ERROR> Error pinging node %s: %v", err, node.Name, err)
// 	}

// 	// TODO: Output parsing using template
// 	if output != nil && string(output) != "" {
// 		logChan <- fmt.Sprintf(">> %s : %s \n\n", node.Name, output)
// 	}
// 	// if _, err := fmt.Fprintf(file, "%s : %s \n\n", node.Name, output); err != nil {
// 	// 	// log.Printf("<ERROR> Error writing to %s: %v", PING_FILE_NAME, err)
// 	// 	logChan <- fmt.Sprintf("<ERROR> Error writing to %s: %v", PING_FILE_NAME, err)
// 	// 	err = fmt.Errorf("%v; <ERROR> Error writing to %s: %v", err, PING_FILE_NAME, err)
// 	// }
// }
