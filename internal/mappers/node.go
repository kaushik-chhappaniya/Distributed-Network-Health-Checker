package mappers

import (
	"fmt"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/discovery"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/models"
)

func MapClabNodeToDBNode(clabNode *discovery.ClabNode) (*models.Node, error) {
	if clabNode == nil {
		return nil, fmt.Errorf("No clab nodes received.")
	}

	status := "unknown"

	return &models.Node{
		Name:      clabNode.LongName,
		IPAddress: clabNode.MgmtIPv4,
		Kind:      clabNode.Kind,
		Status:    status,
	}, nil
}
