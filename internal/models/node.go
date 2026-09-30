package models

import "time"

// Node represents a network node across discovery, health checks, and database persistence.
type Node struct {
	ID        int       `json:"db_id,omitempty"`
	Name      string    `json:"name"`
	IPAddress string    `json:"ip,omitempty"`
	Status    string    `json:"status"`
	Ports     string    `json:"ports,omitempty"`
	Kind      string    `json:"kind,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// Nodes is a slice of Node items.
type Nodes []Node
