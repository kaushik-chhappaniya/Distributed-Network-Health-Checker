package db

import "github.com/jackc/pgx/v5/pgxpool"

const (
	CREATE_TABLE string = `CREATE TABLE IF NOT EXISTS nodes (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) UNIQUE NOT NULL,
		ip_address VARCHAR(45),
		kind VARCHAR(50) DEFAULT 'nokia_srlinux',
		status VARCHAR(20) DEFAULT 'UNKNOWN',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	INSERT_NODE string = `INSERT INTO nodes (name, ip_address, status)
	VALUES ($1, $2, $3)
	ON CONFLICT (name) DO UPDATE
	SET ip_address = EXCLUDED.ip_address, status = EXCLUDED.status;`

	INSERT_NODES_MULTIPLE string = `INSERT INTO nodes (name, ip_address, status)
	SELECT * FROM UNNEST($1::text[], $2::text[], $3::text[])
	ON CONFLICT (name) DO UPDATE
	SET ip_address = EXCLUDED.ip_address, 
    status = EXCLUDED.status,
    updated_at = CURRENT_TIMESTAMP;`
)

/*
Dbpool holds the database pool and provides receiver methods for database operations.
*/
type Dbpool struct {
	pool *pgxpool.Pool
}

/*
New Func is a constructor returning obj of *pgxpool.Pool
*/
func New(pool *pgxpool.Pool) *Dbpool {
	return &Dbpool{
		pool: pool,
	}
}
