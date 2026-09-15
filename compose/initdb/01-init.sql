-- =============================================================================
-- Distributed Network Health Checker - PostgreSQL Initialization
-- Database: mydatabase
-- =============================================================================

-- Table: nodes
-- Stores network devices/nodes to monitor (e.g. Nokia SR Linux, Alpine test nodes)
CREATE TABLE IF NOT EXISTS nodes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    ip_address VARCHAR(45),
    kind VARCHAR(50) DEFAULT 'nokia_srlinux',
    status VARCHAR(20) DEFAULT 'UNKNOWN',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Table: ping_metrics
-- Stores health check ping test results and telemetry metrics
CREATE TABLE IF NOT EXISTS ping_metrics (
    id SERIAL PRIMARY KEY,
    node_id INTEGER REFERENCES nodes(id) ON DELETE CASCADE,
    target VARCHAR(100) NOT NULL,
    packets_transmitted INTEGER DEFAULT 0,
    packets_received INTEGER DEFAULT 0,
    packet_loss_percent NUMERIC(5,2) DEFAULT 0.0,
    min_rtt_ms NUMERIC(8,3) DEFAULT -1,
    avg_rtt_ms NUMERIC(8,3) DEFAULT -1,
    max_rtt_ms NUMERIC(8,3) DEFAULT -1,
    status VARCHAR(20) NOT NULL, -- 'SUCCESS', 'FAILED', 'DEGRADED'
    raw_output TEXT,
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed initial nodes from example topology
INSERT INTO nodes (name, ip_address, kind, status) VALUES
    ('clab-srl01-srl1', '172.20.20.2', 'nokia_srlinux', 'ACTIVE'),
    ('clab-srl01-srl2', '172.20.20.3', 'nokia_srlinux', 'ACTIVE')
ON CONFLICT (name) DO NOTHING;

