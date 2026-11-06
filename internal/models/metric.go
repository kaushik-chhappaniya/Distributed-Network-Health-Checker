package models

import "time"

// PingMetric represents telemetry metrics collected from a network probe.
type PingMetric struct {
	ID                 int       `json:"id,omitempty"`
	NodeID             int       `json:"node_id,omitempty"`
	Target             string    `json:"target"`
	PacketsTransmitted int       `json:"packets_transmitted,omitempty"`
	PacketsReceived    int       `json:"packets_received,omitempty"`
	PacketLossPercent  float64   `json:"packet_loss_percent,omitempty"`
	MinRTTMs           float64   `json:"min_rtt_ms,omitempty"`
	AvgRTTMs           float64   `json:"avg_rtt_ms,omitempty"`
	MaxRTTMs           float64   `json:"max_rtt_ms,omitempty"`
	Status             string    `json:"status"`
	RawOutput          string    `json:"raw_output,omitempty"`
	CheckedAt          time.Time `json:"checked_at,omitempty"`
}
