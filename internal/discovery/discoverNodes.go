package discovery

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// discoverNodes discovers nodes based on the provided image and returns
//
//	a map of node names to their IP addresses.
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
