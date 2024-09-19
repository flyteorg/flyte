package plugin

import (
	"context"

	"google.golang.org/grpc/codes"
)

// ClusterResourcePlugin defines a subset of the Union Cloud service API that is used by the cluster resource controller
type ClusterResourcePlugin interface {
	BatchUpdateProvisioned(ctx context.Context, input *BatchUpdateProvisionedInput) (BatchUpdateProvisionedOutput, []BatchUpdateProvisionedError, error)
}

type BatchUpdateProvisionedInput struct {
	OrgsName []string
}

type BatchUpdateProvisionedOutput struct{}

type BatchUpdateProvisionedError struct {
	OrgName      string
	ErrorMessage string
	ErrorCode    codes.Code
}

// NoopClusterResourcePlugin is a noops implementation of the ClusterResourcePlugin interface.
type NoopClusterResourcePlugin struct {
}

// BatchUpdateProvisioned does nothing.
func (n *NoopClusterResourcePlugin) BatchUpdateProvisioned(ctx context.Context, input *BatchUpdateProvisionedInput) (BatchUpdateProvisionedOutput, []BatchUpdateProvisionedError, error) {
	return BatchUpdateProvisionedOutput{}, nil, nil
}

func NewNoopClusterResourcePlugin() *NoopClusterResourcePlugin {
	return &NoopClusterResourcePlugin{}
}
