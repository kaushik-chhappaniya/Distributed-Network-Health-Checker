package discovery

import (
	"fmt"

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

// Discovery discovers nodes based on the provided image and returns a
// map of node names to their IP addresses.
func (nodes *Nodes) Discovery(img NodeImage) error {
	// Start the logger instance for the nodes Discovery
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
