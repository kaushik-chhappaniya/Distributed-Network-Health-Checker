package discovery

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const JSON_FILE_NAME = "Distributed-Network-Health-Checker/compose/clab-srl01/topology-data.json"

type ClabTopology struct {
	Name  string              `json:"name"`
	Type  string              `json:"type"`
	Nodes map[string]ClabNode `json:"nodes"`
}

type ClabNode struct {
	LongName string `json:"longname"`
	FQDN     string `json:"fqdn"`
	Group    string `json:"group"`
	MgmtIPv4 string `json:"mgmt-ipv4-address"`
	Kind     string `json:"kind"`
}

/*
Discover nodes form reading the topology file of clab.
*/
func (s *Service) discoverClabNodes(config DiscoveryData) (ClabNodes, error) {
	// find the json file. parse it based on nodes and return nodes.
	fileData, err := os.ReadFile(string(*config.TopologyFile))
	if err != nil {
		return nil, fmt.Errorf("<ERROR> Error in reading the JSON file: %w", err)
	}

	// we can unmarshal directly as per our struct and only required fields
	//  if we've a complete static struct defined.
	var topology ClabTopology
	if err := json.Unmarshal(fileData, &topology); err != nil {
		return nil, fmt.Errorf("<ERROR> Error in unmarshaling topology file: %w", err)
	}

	if len(topology.Nodes) == 0 {
		return nil, fmt.Errorf("no nodes found in topology")
	}
	var nodes ClabNodes
	// Range over the unmarshaled nodes data to get the nodes and append them to the clab nodes and return
	for name, node := range topology.Nodes {
		s.log.Println("<NODE> name: ", name, " Node:", node)
		nodes = append(nodes, node)
	}

	log.Printf("<INFO> JSON Topology data captured.")
	return nodes, nil
}
