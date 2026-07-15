package discovery

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	logs "kaushik-chhappaniya/distributed-network-health-checker/utils/logger"
)

type NodeImage string
type Nodes []Node
type Node struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Status string `json:"status"`
	IP     string `json:"ip" default:""`
	Ports  string `json:"ports"`
}

// Discovery discovers nodes based on the provided image and returns a map of node names to their IP addresses.
func (nodes *Nodes) Discovery(img NodeImage) error {
	log, logComplete, err := logs.New("Discovery.log", false)
	defer logComplete()

	if err != nil {
		panic("<ERROR> Logger instance starting failed.")
	}

	log.Println("Started nodes discovery.")
	err = nodes.discoverNodes(log, img)
	if err != nil {
		log.Printf("<ERROR> Error discovering nodes: %v", err)
		return fmt.Errorf("%v; <ERROR> Error discovering nodes: %v", err, err)
	}
	if len(*nodes) == 0 {
		log.Println("<INFO> No nodes discovered.")
		return fmt.Errorf("<INFO> No nodes discovered.")
	}
	log.Printf("Discovered nodes: %v\n", *nodes)
	return nil
}

// discoverNodes discovers nodes based on the provided image and returns a map of node names to their IP addresses.
func (nodes *Nodes) discoverNodes(log *log.Logger, img NodeImage) error {

	// docker ps --format '{"id":"{{.Name}}: {{range .Names}}status: {{.Status}}, Port:{{.Port}}{{end}}"

	// docker ps --format '{"id": "{{.ID}}", "name": "{{.Names}}", "status": "{{.Status}}", "ports": "{{.Ports}}"}'

	// docker inspect --format '{"id": "{{.Id}}", "name": "{{.Name}}", "status": "{{.State.Status}}", "ip": "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", "ports": "{{json .NetworkSettings.Ports}}"}' $(docker ps -q)
	const CONTAINERS_CMD = `docker ps --format '{"id": "{{.ID}}", "name": "{{.Names}}", "status": "{{.Status}}", "ports": "{{.Ports}}"}'`
	log.Printf("Executing command: %s\n", CONTAINERS_CMD)
	output, err := exec.Command("sh", "-c", CONTAINERS_CMD).Output()
	if err != nil {
		log.Printf("<ERROR> Error executing command: %v", err)
		return err
	}
	log.Printf("Command output: %s\n", output)

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		var n Node
		if err := json.Unmarshal(scanner.Bytes(), &n); err != nil {
			log.Printf("<ERROR> Error unmarshalling JSON: %v", err)
			err = fmt.Errorf("%v; <ERROR> Error unmarshalling JSON: %v", err, err)
			continue
		}
		if n.Name != "" && strings.Contains(n.Name, string(img)) {
			*nodes = append(*nodes, n)
		}
	}
	log.Printf("Discovered nodes: %v\n", *nodes)
	return nil
}
