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
func CreateModels(pool *pgxpool.Pool, log *log.Logger) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx, CREATE_TABLE)
	if err != nil {
		return "", fmt.Errorf("failed to create tables: %w", err)
	}
	log.Println("<INFO> Ensured tables exist.")

	var greeting string
	err = pool.QueryRow(ctx, "SELECT 'PostgreSQL connection verified'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return "", err
	}

	log.Println("<INFO> Database ready:", greeting)
	return greeting, nil
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
func (d *Dbpool) InsertNode(log *log.Logger, node *models.Node) (string, error) {
	return InsertNode(d.pool, log, node)
}

/*
InsertMultipleNodes inserts multiple nodes into the postgreSQL database.
*/
func (d *Dbpool) InsertMultipleNodes(log *log.Logger, nodes models.Nodes) error {
	// return InsertMultipleNodes(d.pool, log, nodes)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	batch := &pgx.Batch{}
	log.Printf("<INFO> Sending data to process batch.")
	for _, n := range nodes {
		query := INSERT_NODE
		batch.Queue(query, n.Name, n.IPAddress, n.Status) // Uses queue to lineup all the database
	}

	// Send the batch to the connection pool
	br := d.pool.SendBatch(ctx, batch)
	defer br.Close()

	log.Printf("<INFO> Batch execution is started.")
	// You must call Close or loop through the results to check for execution errors
	for i := range nodes {
		_, err := br.Exec()
		if err != nil {
			log.Printf("<ERROR> Batch statement %d failed: %s", i, err.Error())
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
	log.Printf("<INFO> Successfully saved %d nodes to DB", len(nodes))
	return nil
}
