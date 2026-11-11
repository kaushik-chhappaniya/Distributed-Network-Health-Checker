package discovery

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
)

const EXACTFILENAME = "topology-data.json"

var errFileFound = errors.New("file found")

type NodeImage string
type ClabTopologyFile string
type ClabNodes []ClabNode

// Service struct for all parameters to be shared
type Service struct {
	log *log.Logger
}

// Constructor for returning logger instance across all functions
func New(logInstance *log.Logger) *Service {
	return &Service{
		log: logInstance,
	}
}

// Config data required for the discovery of nodes- 1.Clab topology or 2.Alpine nodes
type DiscoveryData struct {
	TopologyFile *ClabTopologyFile
	Image        *NodeImage
	Nodes        ClabNodes
}

// Discover discovers nodes based on the provided image and populates the nodes slice.
func (s *Service) Discover(config *DiscoveryData) error {
	// Start the logger instance for the nodes Discovery
	s.log.Println("Started nodes discovery.")
	if config.Image != nil {
		s.log.Println("<INFO> Config image provided proceeding with alpine node discovery.")
		// Perform alpine nodes discovery
		_, err := s.discoverNodes(config)
		if err != nil {
			s.log.Printf("<ERROR> Error in discovering sros nodes.")
		}
	} else if config.TopologyFile != nil {
		// We discover the clab nodes using the topology file given. Assign the returned nodes to config struct.
		s.log.Println("<INFO> Topology file provided proceeding with alpine node discovery.")
		clabNodes, err := s.discoverClabNodes(*config)
		if err != nil {
			s.log.Printf("<ERROR> Error discovering nodes: %v", err)
			return fmt.Errorf("%v; <ERROR> Error discovering nodes: %v", err, err)
		}
		s.log.Println("<INFO> Printing nodes here:", clabNodes)
		config.Nodes = clabNodes
		if len(config.Nodes) == 0 {
			s.log.Println("<INFO> No nodes discovered.")
			return fmt.Errorf("<INFO> No nodes discovered.")
		}
	}
	s.log.Printf("Discovered nodes: %v\n", config.Nodes)
	return nil
}

// DiscoverTopology finds the topology file recursively in the given location.
func (s *Service) DiscoverTopology(rootDir string) (ClabTopologyFile, error) {
	s.log.Println("<INFO> Finding the json file.")
	var foundPath ClabTopologyFile
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // Pass error up to stop the walk
		}
		// Check for an exact case-sensitive match
		if !d.IsDir() && d.Name() == EXACTFILENAME {
			foundPath = ClabTopologyFile(path)
			return errFileFound // Return sentinel error to halt execution immediately
		}
		return nil
	})
	// Check if the walk stopped because we found the file
	if errors.Is(err, errFileFound) {
		return foundPath, nil
	}

	return "", err
}
