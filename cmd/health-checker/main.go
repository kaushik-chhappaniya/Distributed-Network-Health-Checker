package main

import (
	logs "kaushik-chhappaniya/distributed-network-health-checker/utils/logger"
)

func main() {
	log, logComplete, err := logs.New("docker-compose.log")
	defer logComplete()
	if err != nil {
		panic("Logger instance starting failed.")
	}
	log.Println("Application started.")
}