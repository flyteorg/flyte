package plugin

import (
	"context"

	"google.golang.org/grpc/codes"
)

// ClusterResourcePlugin defines a subset of the Union Cloud service API that is used by the cluster resource controller
type ClusterResourcePlugin interface {
	BatchUpdateClusterResourceState(ctx context.Context, input *BatchUpdateClusterResourceStateInput) (BatchUpdateClusterResourceStateOutput, []BatchUpdateClusterResourceStateError, error)
}

type ClusterResourceState int

const (
	ClusterResourceStateProvisioned ClusterResourceState = iota
	ClusterResourceStateDeprovisioned
)

type BatchUpdateClusterResourceStateInput struct {
	OrgsName             []string
	ClusterName          string
	ClusterResourceState ClusterResourceState
}

type BatchUpdateClusterResourceStateOutput struct{}

type BatchUpdateClusterResourceStateError struct {
	OrgName      string
	ErrorMessage string
	ErrorCode    codes.Code
}

// NoopClusterResourcePlugin is a noops implementation of the ClusterResourcePlugin interface.
type NoopClusterResourcePlugin struct {
}

// BatchUpdateClusterResourceState does nothing.
func (n *NoopClusterResourcePlugin) BatchUpdateClusterResourceState(ctx context.Context, input *BatchUpdateClusterResourceStateInput) (BatchUpdateClusterResourceStateOutput, []BatchUpdateClusterResourceStateError, error) {
	return BatchUpdateClusterResourceStateOutput{}, nil, nil
}

func NewNoopClusterResourcePlugin() *NoopClusterResourcePlugin {
	return &NoopClusterResourcePlugin{}
}
