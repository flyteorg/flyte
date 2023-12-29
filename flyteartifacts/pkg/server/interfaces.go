package server

import (
	"context"

	"github.com/golang/protobuf/ptypes/any"

	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	stdLibStorage "github.com/flyteorg/flyte/flytestdlib/storage"
)

type StorageInterface interface {
	CreateArtifact(context.Context, models.Artifact) (models.Artifact, error)

	GetArtifact(context.Context, core.ArtifactQuery) (models.Artifact, error)

	CreateTrigger(context.Context, models.Trigger) (models.Trigger, error)

	GetLatestTrigger(ctx context.Context, project, domain, name string) (models.Trigger, error)

	GetTriggersByArtifactKey(ctx context.Context, key core.ArtifactKey) ([]models.Trigger, error)

	// DeleteTrigger(context.Context, models.Trigger) error

	SearchArtifacts(context.Context, artifact.SearchArtifactsRequest) ([]models.Artifact, string, error)

	SetExecutionInputs(ctx context.Context, req *artifact.ExecutionInputsRequest) error

	FindByWorkflowExec(ctx context.Context, request *artifact.FindByWorkflowExecRequest) ([]models.Artifact, error)
}

type BlobStoreInterface interface {
	OffloadArtifactCard(ctx context.Context, name string, version string, userMetadata *any.Any) (stdLibStorage.DataReference, error)

	RetrieveArtifactCard(context.Context, stdLibStorage.DataReference) (*any.Any, error)
}

type TriggerHandlerInterface interface {
	EvaluateNewArtifact(context.Context, *artifact.Artifact) ([]core.WorkflowExecutionIdentifier, error)
}
