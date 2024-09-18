package plugin

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

// ClusterResourcePlugin defines a subset of the Union Cloud service API that is used by the cluster resource controller
type ClusterResourcePlugin interface {
	BatchUpdateProvisioned(ctx context.Context, input *BatchUpdateProvisionedInput) (BatchUpdateProvisionedOutput, []BatchUpdateProvisionedError, error)
	GetProvisionNamespaceProjectFilter(ctx context.Context) (string, error)
	GetDeleteNamespaceProjectFilter(ctx context.Context) (string, error)
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

// DefaultClusterResourcePlugin is the default implementation of the ClusterResourcePlugin interface.
type DefaultClusterResourcePlugin struct {
}

// BatchUpdateProvisioned does nothing.
func (n *DefaultClusterResourcePlugin) BatchUpdateProvisioned(ctx context.Context, input *BatchUpdateProvisionedInput) (BatchUpdateProvisionedOutput, []BatchUpdateProvisionedError, error) {
	return BatchUpdateProvisionedOutput{}, nil, nil
}

func (n *DefaultClusterResourcePlugin) GetProvisionNamespaceProjectFilter(ctx context.Context) (string, error) {
	return fmt.Sprintf("value_in(state,%d;%d)", admin.Project_ACTIVE, admin.Project_SYSTEM_GENERATED), nil
}

func (n *DefaultClusterResourcePlugin) GetDeleteNamespaceProjectFilter(ctx context.Context) (string, error) {
	return fmt.Sprintf("eq(state,%d)", admin.Project_SYSTEM_ARCHIVED), nil
}

func NewDefaultClusterResourcePlugin() *DefaultClusterResourcePlugin {
	return &DefaultClusterResourcePlugin{}
}
