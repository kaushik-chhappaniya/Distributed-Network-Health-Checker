package db

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

	GET_NODES string = `SELECT id, name, COALESCE(ip_address, ''), COALESCE(kind, ''), COALESCE(status, ''), created_at, updated_at FROM nodes;`

	PING_METRICS string = `CREATE TABLE IF NOT EXISTS ping_metrics (
    id SERIAL PRIMARY KEY,
    target VARCHAR(100) NOT NULL,
    packets_transmitted INT NOT NULL,
    packets_received INT NOT NULL,
    packet_loss_percent NUMERIC(5, 2) NOT NULL,
    min_rtt_ms NUMERIC(8, 3),
    avg_rtt_ms NUMERIC(8, 3),
    max_rtt_ms NUMERIC(8, 3),
    status VARCHAR(20) NOT NULL,
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);`

	INSERT_PING_METRIC string = `INSERT INTO ping_metrics (
		target, packets_transmitted, packets_received, packet_loss_percent,
		min_rtt_ms, avg_rtt_ms, max_rtt_ms, status, checked_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`

	GET_PING_METRICS string = `SELECT id, target, packets_transmitted, packets_received, packet_loss_percent,
		COALESCE(min_rtt_ms, 0), COALESCE(avg_rtt_ms, 0), COALESCE(max_rtt_ms, 0),
		status, checked_at
	FROM ping_metrics
	ORDER BY checked_at DESC;`
)
