package pinghelper

import (
	"context"
	"fmt"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/models"
	"kaushik-chhappaniya/distributed-network-health-checker/utils/parser"
	"os/exec"
	"time"
)

/*
ping_helper executes the system ping command for a given node. It does not
handle concurrency or file I/O itself; it relies on the logChan to report
execution status, errors, and raw output back to the central logger.
*/
func PingHelper(nodeName string, logChan chan<- string, ping_target string) (*models.PingMetric, error) {
	PING_CMD := "ping -c 5 -q -s 20 -w 10 %s"
	// log.Printf("<INFO> Pinging node: %s with IP: %s", node.Name, node.IP)
	logChan <- fmt.Sprintf("<INFO> Pinging node: %s with IP: %s", nodeName, ping_target)

	// Create context with 10 seconds timeout for the ping command
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Cancel the context to free resources if the command finishes before timeout

	// Execute the ping command using exec.CommandContext to allow for timeout
	cmd := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf(PING_CMD, ping_target))
	output, err := cmd.Output()

	// output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		logChan <- fmt.Sprintf("<ERROR> Error pinging node %s: %v", nodeName, err)
		err = fmt.Errorf("%v; <ERROR> Error pinging node %s: %v", err, nodeName, err)
	}

	// Log raw output to the logger channel
	if output != nil && string(output) != "" {
		logChan <- fmt.Sprintf(">> %s : %s \n\n", nodeName, output)
	}

	// Parse raw ping output into structured PingMetric using TextFSM
	metric, parseErr := parser.ParsePingOutput(ping_target, string(output))
	if parseErr != nil {
		logChan <- fmt.Sprintf("<ERROR> Error parsing ping output for node %s: %v", nodeName, parseErr)
		return nil, parseErr
	}

	return metric, nil
}
