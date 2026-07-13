package main

import (
	"flag"
	"strings"

	compose "kaushik-chhappaniya/distributed-network-health-checker/internal/compose"
	logs "kaushik-chhappaniya/distributed-network-health-checker/utils/logger"
)

func main() {
	log, logComplete, err := logs.New("docker-compose.log")
	defer logComplete()
	if err != nil {
		panic("Logger instance starting failed.")
	}
	log.Println("Docker Compose Application started.")

	nodes := flag.Int("nodes", 1, "Number of nodes.")
	outputFile := flag.String("output", "docker-compose.yml", "Docker compose file name.")
	flag.Parse()

	if *nodes < 1 {
		log.Println("No nodes present. Defaulting to 1 node.")
	}
	if outputFile == nil || strings.TrimSpace(*outputFile) == "" {
		log.Println("No output file name provided. Defaulting to docker-compose.yml.")
	}

	cfg := compose.Config{
		Nodes:   *nodes,
		Output:  *outputFile,
		Network: "health-net",
		Image:   "debian:bookworm-slim",
	}
	compose.Generate(cfg, log)

}
