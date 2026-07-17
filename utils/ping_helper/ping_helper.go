package pinghelper

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

/*
ping_helper executes the system ping command for a given node. It does not
handle concurrency or file I/O itself; it relies on the logChan to report
execution status, errors, and raw output back to the central logger.
*/
func Ping_helper(nodeName string, logChan chan<- string, ping_target string) {
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

	// TODO: Output parsing using template
	if output != nil && string(output) != "" {
		logChan <- fmt.Sprintf(">> %s : %s \n\n", nodeName, output)
	}
	// if _, err := fmt.Fprintf(file, "%s : %s \n\n", node.Name, output); err != nil {
	// 	// log.Printf("<ERROR> Error writing to %s: %v", PING_FILE_NAME, err)
	// 	logChan <- fmt.Sprintf("<ERROR> Error writing to %s: %v", PING_FILE_NAME, err)
	// 	err = fmt.Errorf("%v; <ERROR> Error writing to %s: %v", err, PING_FILE_NAME, err)
	// }
}
