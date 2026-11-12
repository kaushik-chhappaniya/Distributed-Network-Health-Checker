package parser

import (
	_ "embed"
	"fmt"
	"strconv"
	"time"

	"kaushik-chhappaniya/distributed-network-health-checker/internal/models"

	"github.com/sirikothe/gotextfsm"
)

//go:embed templates/ping_result.textfsm
var pingResultTmpl string

// ParsePingOutput parses raw CLI ping output into a structured models.PingMetric.
// Note: gotextfsm.TextFSM mutates internal maps during parsing, so each call
// uses its own local fsm instance to guarantee complete goroutine concurrency safety.
func ParsePingOutput(target string, rawOutput string) (*models.PingMetric, error) {
	// 1. Initialize local TextFSM instance (avoids concurrent map write race conditions)
	fsm := gotextfsm.TextFSM{}
	if err := fsm.ParseString(pingResultTmpl); err != nil {
		return nil, fmt.Errorf("failed to parse ping TextFSM template: %w", err)
	}

	// 2. Parse the raw ping output
	parser := gotextfsm.ParserOutput{}
	if err := parser.ParseTextString(rawOutput, fsm, true); err != nil {
		return &models.PingMetric{
			Target:             target,
			PacketsTransmitted: 0,
			PacketsReceived:    0,
			PacketLossPercent:  100.0,
			Status:             "FAILED",
			CheckedAt:          time.Now(),
		}, fmt.Errorf("textfsm parse error: %w", err)
	}

	if len(parser.Dict) == 0 {
		return &models.PingMetric{
			Target:             target,
			PacketsTransmitted: 0,
			PacketsReceived:    0,
			PacketLossPercent:  100.0,
			Status:             "FAILED",
			CheckedAt:          time.Now(),
		}, fmt.Errorf("no ping statistics matched in output")
	}

	record := parser.Dict[0]

	// Safe helper to extract values without panicking on nil or unexpected types
	getStr := func(k string) string {
		if val, ok := record[k]; ok && val != nil {
			if s, ok := val.(string); ok {
				return s
			}
			return fmt.Sprintf("%v", val)
		}
		return ""
	}

	transmitted, _ := strconv.Atoi(getStr("TRANSMITTED"))
	received, _ := strconv.Atoi(getStr("RECEIVED"))
	loss, _ := strconv.ParseFloat(getStr("LOSS_PERCENT"), 64)
	minRTT, _ := strconv.ParseFloat(getStr("MIN_RTT"), 64)
	avgRTT, _ := strconv.ParseFloat(getStr("AVG_RTT"), 64)
	maxRTT, _ := strconv.ParseFloat(getStr("MAX_RTT"), 64)

	// Determine status based on packet loss
	status := "SUCCESS"
	if loss == 100 || received == 0 {
		status = "FAILED"
	} else if loss > 0 {
		status = "DEGRADED"
	}

	return &models.PingMetric{
		Target:             target,
		PacketsTransmitted: transmitted,
		PacketsReceived:    received,
		PacketLossPercent:  loss,
		MinRTTMs:           minRTT,
		AvgRTTMs:           avgRTT,
		MaxRTTMs:           maxRTT,
		Status:             status,
		CheckedAt:          time.Now(),
	}, nil
}
