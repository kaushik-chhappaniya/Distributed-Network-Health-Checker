package db

/*
 Handles connecting to Postgres using the environment variables
  (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME), pooling configuration,
   and verifying the connection:
*/
import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect initializes and returns a thread-safe PostgreSQL connection pool.
func Connect(log *log.Logger) (*pgxpool.Pool, error) {

	// 1. Build connection URL (read DATABASE_URL or construct from individual env vars)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		pass := os.Getenv("DB_PASSWORD")
		name := os.Getenv("DB_NAME")

		if host == "" {
			host = "db"
		}
		if port == "" {
			port = "5432"
		}
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)
	}

	log.Println("<INFO> Using db connection string as:", dbURL)

	// 2. Configure connection pool with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Printf("<ERROR> Failed to create DB pool: %v", err)
		return nil, fmt.Errorf("failed to create DB pool: %w", err)
	}

	// 3. Ping the database to verify the connection
	if err := pool.Ping(ctx); err != nil {
		log.Printf("<ERROR> Failed to ping DB: %v", err)
		pool.Close()
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}
	log.Println("<INFO> Database connection pool established successfully.")

	return pool, nil
}
