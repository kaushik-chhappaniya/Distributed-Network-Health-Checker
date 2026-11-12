package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"kaushik-chhappaniya/distributed-network-health-checker/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
CreateModels creates the database schema if it doesn't already exist and verifies connectivity.
*/
func (s *Service) CreateModels() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.pool.Exec(ctx, CREATE_TABLE)
	if err != nil {
		return "", fmt.Errorf("failed to create tables: %w", err)
	}
	s.logger.Println("<INFO> Ensured tables exist.")

	_, err = s.pool.Exec(ctx, PING_METRICS)
	if err != nil {
		return "", fmt.Errorf("failed to create ping_metrics table: %w", err)
	}
	s.logger.Println("<INFO> Ensured ping_metrics table exists.")

	var greeting string
	err = s.pool.QueryRow(ctx, "SELECT 'PostgreSQL connection verified'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return "", err
	}

	s.logger.Println("<INFO> Database ready:", greeting)
	return greeting, nil
}

// Ping checks the database connection health.
func (s *Service) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// InsertNode inserts or updates a single node in the database.
func InsertNode(pool *pgxpool.Pool, log *log.Logger, node *models.Node) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	op, err := pool.Exec(ctx, INSERT_NODE, node.Name, node.IPAddress, node.Status)
	if err != nil {
		log.Printf("<ERROR> Failed to insert node %s: %v", node.Name, err)
		return "", err
	}
	log.Printf("<INFO> Saved node %s to DB: %s", node.Name, op.String())
	return op.String(), nil
}

// func InsertMultipleNodes(pool *pgxpool.Pool, log *log.Logger, nodes models.Nodes) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	batch := &pgx.Batch{}

// 	for _, n := range nodes {
// 		query := INSERT_NODE
// 		batch.Queue(query, n.Name, n.IPAddress, n.Status) //
// 	}

// 	// Send the batch to the connection pool
// 	br := pool.SendBatch(ctx, batch)
// 	defer br.Close()

// 	// You must call Close or loop through the results to check for execution errors
// 	for i := 0; i < len(nodes); i++ {
// 		_, err := br.Exec()
// 		if err != nil {
// 			return fmt.Errorf("batch statement %d failed: %w", i, err)
// 		}
// 	}
// 	// _, err := pool.Exec(ctx)
// 	// for _, node := range nodes {
// 	// 	_, err := pool.Exec(ctx, INSERT_NODE, node.Name, node.IPAddress, node.Status)
// 	// 	if err != nil {
// 	// 		log.Printf("<ERROR> Failed to insert node %s: %v", node.Name, err)
// 	// 		return err
// 	// 	}
// 	// }
// 	log.Printf("<INFO> Successfully saved %d nodes to DB", len(nodes))
// 	return nil
// }

/*
InsertNode inserts the node into the postgreSQL database.
This method called on receiver Dbpool. Abstracting actual definition of InsertNode
*/
func (s *Service) InsertNode(node *models.Node) (string, error) {
	return InsertNode(s.pool, s.logger, node)
}

/*
InsertMultipleNodes inserts multiple nodes into the postgreSQL database.
*/
func (s *Service) InsertMultipleNodes(nodes models.Nodes) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	batch := &pgx.Batch{}
	s.logger.Printf("<INFO> Sending data to process batch.")
	for _, n := range nodes {
		query := INSERT_NODE
		batch.Queue(query, n.Name, n.IPAddress, n.Status) // Uses queue to lineup all the database
	}

	// Send the batch to the connection pool
	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	s.logger.Printf("<INFO> Batch execution is started.")
	// You must call Close or loop through the results to check for execution errors
	for i := range nodes {
		_, err := br.Exec()
		if err != nil {
			s.logger.Printf("<ERROR> Batch statement %d failed: %s", i, err.Error())
			return fmt.Errorf("<ERROR> Batch statement %d failed: %w", i, err)
		}
	}
	// _, err := pool.Exec(ctx)
	// for _, node := range nodes {
	// 	_, err := pool.Exec(ctx, INSERT_NODE, node.Name, node.IPAddress, node.Status)
	// 	if err != nil {
	// 		log.Printf("<ERROR> Failed to insert node %s: %v", node.Name, err)
	// 		return err
	// 	}
	// }
	s.logger.Printf("<INFO> Successfully saved %d nodes to DB", len(nodes))
	return nil
}

func (s *Service) GetAllNodes(
	ctx context.Context,
) (models.Nodes, error) {
	// PostgreSQL query to fetch all nodes
	s.logger.Printf("<INFO> Querying all nodes.")
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	// Query all rows from the database
	rows, err := s.pool.Query(ctx, GET_NODES)
	if err != nil {
		s.logger.Printf("<ERROR> Query failed: %v\n", err)
		return nil, fmt.Errorf("failed to query nodes: %w", err)
	}
	// CollectRows implicitly closes rows, but deferring ensures closure if helper panic/errors occur early
	defer rows.Close()

	// Scan all rows into a slice of structs automatically
	var nodes models.Nodes
	for rows.Next() {
		var n models.Node
		err := rows.Scan(
			&n.ID,
			&n.Name,
			&n.IPAddress,
			&n.Kind,
			&n.Status,
			&n.CreatedAt,
			&n.UpdatedAt,
		)
		if err != nil {
			s.logger.Printf("<ERROR> Failed to scan node row: %v", err)
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}
		nodes = append(nodes, n)
	}
	// nodes, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Node])
	// if err != nil {
	// 	s.logger.Printf("<ERROR> Failed to collect rows: %v\n", err)
	// }
	if err := rows.Err(); err != nil {
		s.logger.Printf("<ERROR> Rows iteration error: %v", err)
		return nil, fmt.Errorf("error reading node rows: %w", err)
	}
	return nodes, nil
}

func (s *Service) UpdatePingResults(
	ctx context.Context,
	results []models.PingMetric,
) error {
	if len(results) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, m := range results {
		batch.Queue(
			INSERT_PING_METRIC,
			m.Target,
			m.PacketsTransmitted,
			m.PacketsReceived,
			m.PacketLossPercent,
			m.MinRTTMs,
			m.AvgRTTMs,
			m.MaxRTTMs,
			m.Status,
			m.CheckedAt,
		)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := range results {
		_, err := br.Exec()
		if err != nil {
			s.logger.Printf("<ERROR> Failed to insert ping metric statement %d: %v", i, err)
			return fmt.Errorf("failed to insert ping metric statement %d: %w", i, err)
		}
	}

	s.logger.Printf("<INFO> Successfully saved %d ping metrics to DB", len(results))
	return nil
}

// GetPingMetrics queries all stored ping telemetry metrics ordered by latest first.
func (s *Service) GetPingMetrics(ctx context.Context) ([]models.PingMetric, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := s.pool.Query(ctx, GET_PING_METRICS)
	if err != nil {
		s.logger.Printf("<ERROR> Query failed for ping metrics: %v", err)
		return nil, fmt.Errorf("failed to query ping metrics: %w", err)
	}
	defer rows.Close()

	metrics := make([]models.PingMetric, 0)
	for rows.Next() {
		var m models.PingMetric
		err := rows.Scan(
			&m.ID,
			&m.Target,
			&m.PacketsTransmitted,
			&m.PacketsReceived,
			&m.PacketLossPercent,
			&m.MinRTTMs,
			&m.AvgRTTMs,
			&m.MaxRTTMs,
			&m.Status,
			&m.CheckedAt,
		)
		if err != nil {
			s.logger.Printf("<ERROR> Failed to scan ping metric row: %v", err)
			return nil, fmt.Errorf("failed to scan ping metric: %w", err)
		}
		metrics = append(metrics, m)
	}

	if err := rows.Err(); err != nil {
		s.logger.Printf("<ERROR> Rows iteration error on ping metrics: %v", err)
		return nil, fmt.Errorf("error reading ping metric rows: %w", err)
	}

	return metrics, nil
}
