package plugin

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

// ClusterResourcePlugin defines a subset of the Union Cloud service API that is used by the cluster resource controller
//
//go:generate mockery -name ClusterResourcePlugin -output=./mocks -case=underscore
type ClusterResourcePlugin interface {
	BatchUpdateClusterResourceState(ctx context.Context, input *BatchUpdateClusterResourceStateInput) (BatchUpdateClusterResourceStateOutput, []BatchUpdateClusterResourceStateError, error)
	GetProvisionProjectFilter(ctx context.Context) (string, error)
	GetDeprovisionProjectFilter(ctx context.Context) (string, error)
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

// DefaultClusterResourcePlugin is the default implementation of the ClusterResourcePlugin interface.
type DefaultClusterResourcePlugin struct {
}

// BatchUpdateClusterResourceState does nothing.
func (n *DefaultClusterResourcePlugin) BatchUpdateClusterResourceState(ctx context.Context, input *BatchUpdateClusterResourceStateInput) (BatchUpdateClusterResourceStateOutput, []BatchUpdateClusterResourceStateError, error) {
	return BatchUpdateClusterResourceStateOutput{}, nil, nil
}

func (n *DefaultClusterResourcePlugin) GetProvisionProjectFilter(ctx context.Context) (string, error) {
	return fmt.Sprintf("value_not_in(state,%d;%d)", admin.Project_ARCHIVED, admin.Project_SYSTEM_ARCHIVED), nil
}

func (n *DefaultClusterResourcePlugin) GetDeprovisionProjectFilter(ctx context.Context) (string, error) {
	return fmt.Sprintf("value_in(state,%d;%d)", admin.Project_ARCHIVED, admin.Project_SYSTEM_ARCHIVED), nil
}

func NewDefaultClusterResourcePlugin() *DefaultClusterResourcePlugin {
	return &DefaultClusterResourcePlugin{}
}
