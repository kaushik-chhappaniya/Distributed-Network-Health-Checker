package db

import (
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

/*
Dbpool holds the database pool and provides receiver methods for database operations.
*/
type Dbpool struct {
	pool *pgxpool.Pool
}

/*
New Func is a constructor returning obj of *pgxpool.Pool and logger instance for dependancy injection
*/
type Service struct {
	pool   *pgxpool.Pool
	logger *log.Logger
}

func New(pool *pgxpool.Pool, logger *log.Logger) *Service {
	return &Service{
		pool:   pool,
		logger: logger,
	}
}
