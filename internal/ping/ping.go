package ping

import (
	"context"
	logs "kaushik-chhappaniya/distributed-network-health-checker/utils/logger"
	"log"
	"sync"

	"kaushik-chhappaniya/distributed-network-health-checker/internal/models"
)

const PING_FILE_NAME = "/app/output/ping_output.txt"

type NodeRepository interface {
	GetAllNodes(ctx context.Context) (models.Nodes, error)
	UpdatePingResults(ctx context.Context, results []models.PingMetric) error
	GetPingMetrics(ctx context.Context) ([]models.PingMetric, error)
}

type Service struct {
	repo        NodeRepository
	workerCount int
	log         *log.Logger
}

type PingTarget struct {
	Name         string
	Target       string
	ExecFromNode bool
}

/*
Constructor returning a struct with database connection pool
*/
func New(repo NodeRepository, workerCount int, log *log.Logger) *Service {
	return &Service{
		repo:        repo,
		workerCount: workerCount,
		log:         log,
	}
}

/*
Ping orchestrates the ping utility lifecycle. It initializes the logging infrastructure,
sets up the worker pool, manages the communication channels, and ensures a graceful
shutdown of all concurrent components.
*/
func (s *Service) PingNodes(ctx context.Context) error {

	s.log.Println("<INFO> Inside Ping nodes.")
	nodes, err := s.repo.GetAllNodes(ctx)
	if err != nil {
		s.log.Println("<ERROR> Error in getting all nodes from the database.")
		return err
	}

	pings := make(chan PingTarget, len(nodes)) // Buffer is fine to be larger than data
	logChan := make(chan string, 100)
	resultsChan := make(chan models.PingMetric, len(nodes))

	var wg sync.WaitGroup
	var logWg sync.WaitGroup // For the logger

	// Starting the logger first
	logWg.Add(1)
	go logs.LoggerGoroutine(logChan, &logWg, s.log, PING_FILE_NAME)

	// Start configured worker pool
	for w := 1; w <= s.workerCount; w++ {
		wg.Add(1)
		s.log.Printf("<WORKER> Starting ping worker %d\n", w)
		go worker(pings, logChan, resultsChan, &wg)
	}

	// Feeding ping data - send the node data to the pings channel and then close it to signal no more data
	s.log.Println("Feeding ping data - send the node data to the pings channel.")
	for _, node := range nodes {
		pings <- PingTarget{Name: node.Name, Target: node.IPAddress, ExecFromNode: false}
	}
	close(pings) // Signal to workers that no more data will be sent

	wg.Wait()          // Wait for workers to finish processing
	close(resultsChan) // Close results channel so we can drain it
	close(logChan)     // Close log channel to stop the logger loop
	logWg.Wait()       // Wait for logger to finish writing remaining logs

	// Collect parsed metrics from workers
	var metrics []models.PingMetric
	for m := range resultsChan {
		metrics = append(metrics, m)
	}

	// Persist metrics into PostgreSQL
	if err := s.repo.UpdatePingResults(ctx, metrics); err != nil {
		s.log.Printf("<ERROR> Failed to save ping metrics to DB: %v", err)
		return err
	}

	s.log.Printf("<INFO> Ping check completed and %d metrics saved to DB; output stored at %s\n", len(metrics), PING_FILE_NAME)
	return nil
}

// GetPingMetrics fetches all recorded ping telemetry metrics from the repository.
func (s *Service) GetPingMetrics(ctx context.Context) ([]models.PingMetric, error) {
	return s.repo.GetPingMetrics(ctx)
}
