package discovery

import (
	"fmt"

	"kaushik-chhappaniya/distributed-network-health-checker/internal/models"
	logs "kaushik-chhappaniya/distributed-network-health-checker/utils/logger"
)

type NodeImage string
type Node = models.Node
type Nodes = models.Nodes

// Discover discovers nodes based on the provided image and populates the nodes slice.
func Discover(img NodeImage, nodes *models.Nodes) error {
	// Start the logger instance for the nodes Discovery
	log, logComplete, err := logs.New("Discovery.log", false)
	defer logComplete()

	if err != nil {
		return fmt.Errorf("<ERROR> Logger instance starting failed: %w", err)
	}

	log.Println("Started nodes discovery.")
	err = discoverNodes(nodes, log, img)
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
