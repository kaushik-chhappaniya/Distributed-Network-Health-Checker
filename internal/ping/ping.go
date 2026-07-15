package ping

import (
	"fmt"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/discovery"
	"os"
	"os/exec"

	logs "kaushik-chhappaniya/distributed-network-health-checker/utils/logger"
)

const PING_FILE_NAME = "/app/output/ping_output.txt"

func Ping(nodes *discovery.Nodes) error {
	log, logComplete, err := logs.New("Ping.log", false)
	defer logComplete()

	if err != nil {
		panic("<ERROR> Logger instance starting failed.")
	}

	log.Println("<INFO> Started ping check.")

	PING_CMD := "ping -c 5 -q -s 20 -w 10 %s"
	for _, node := range *nodes {
		log.Printf("<INFO> Pinging node: %s with IP: %s", node.Name, node.IP)
		cmd := fmt.Sprintf(PING_CMD, node.Name)
		output, err := exec.Command("sh", "-c", cmd).Output()
		if err != nil {
			log.Printf("<ERROR> Error pinging node %s: %v", node.Name, err)
			err = fmt.Errorf("%v; <ERROR> Error pinging node %s: %v", err, node.Name, err)
		}
		// TODO: Output parsing using template
		file, err := os.OpenFile(PING_FILE_NAME, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("<ERROR> Error opening %s: %v", PING_FILE_NAME, err)
			err = fmt.Errorf("%v; <ERROR> Error opening %s: %v", err, PING_FILE_NAME, err)
		}
		defer file.Close()
		// if _, err := file.Write(fmt.Appendf(nil, "%s : %s \n\n", node.Name, output)); err != nil {
		// if _, err := fmt.Fprintf(file, fmt.Sprintf("%s : %s \n\n", node.Name, output)); err != nil {
		if _, err := fmt.Fprintf(file, "%s : %s \n\n", node.Name, output); err != nil {
			log.Printf("<ERROR> Error writing to %s: %v", PING_FILE_NAME, err)
			err = fmt.Errorf("%v; <ERROR> Error writing to %s: %v", err, PING_FILE_NAME, err)
		}
	}
	log.Printf("<INFO> Ping check completed and output stored at %s\n", PING_FILE_NAME)
	return nil
}
