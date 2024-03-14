package artifacts

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	admin2 "github.com/flyteorg/flyte/flyteidl/clients/go/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifacts"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// IDE "Go Generate File". This will create a mocks/AdminServiceClient.go file
//go:generate mockery -dir ../../../flyteidl/gen/pb-go/flyteidl/artifacts -name ArtifactRegistryClient -output ./mocks

// ArtifactRegistry contains a Client to talk to an Artifact service and has helper methods
type ArtifactRegistry struct {
	Client artifacts.ArtifactRegistryClient
}

func (a *ArtifactRegistry) RegisterArtifactProducer(ctx context.Context, id *core.Identifier, ti core.TypedInterface) {
	if a == nil || a.Client == nil {
		logger.Debugf(ctx, "Artifact Client not configured, skipping registration for task [%+v]", id)
		return
	}

	ap := &artifacts.ArtifactProducer{
		EntityId: id,
		Outputs:  ti.Outputs,
	}
	_, err := a.Client.RegisterProducer(ctx, &artifacts.RegisterProducerRequest{
		Producers: []*artifacts.ArtifactProducer{ap},
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to register artifact producer for task [%+v] with err: %v", id, err)
	}
	logger.Debugf(ctx, "Registered artifact producer [%+v]", id)
}

func (a *ArtifactRegistry) RegisterArtifactConsumer(ctx context.Context, id *core.Identifier, pm core.ParameterMap) {
	if a == nil || a.Client == nil {
		logger.Debugf(ctx, "Artifact Client not configured, skipping registration for consumer [%+v]", id)
		return
	}
	ac := &artifacts.ArtifactConsumer{
		EntityId: id,
		Inputs:   &pm,
	}
	_, err := a.Client.RegisterConsumer(ctx, &artifacts.RegisterConsumerRequest{
		Consumers: []*artifacts.ArtifactConsumer{ac},
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to register artifact consumer for entity [%+v] with err: %v", id, err)
	}
	logger.Debugf(ctx, "Registered artifact consumer [%+v]", id)
}

func (a *ArtifactRegistry) RegisterTrigger(ctx context.Context, plan *admin.LaunchPlan) error {
	if a == nil || a.Client == nil {
		logger.Debugf(ctx, "Artifact Client not configured, skipping trigger [%+v]", plan)
		return fmt.Errorf("artifact Client not configured")
	}
	_, err := a.Client.CreateTrigger(ctx, &artifacts.CreateTriggerRequest{
		TriggerLaunchPlan: plan,
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to register trigger for [%+v] with err: %v", plan.Id, err)
		return err
	}
	logger.Debugf(ctx, "Registered trigger for [%+v]", plan.Id)
	return nil
}

func (a *ArtifactRegistry) ActivateTrigger(ctx context.Context, identifier *core.Identifier) error {
	if a == nil || a.Client == nil {
		logger.Debugf(ctx, "Artifact Client not configured, skipping activate [%+v]", identifier)
		return fmt.Errorf("artifact Client not configured")
	}
	_, err := a.Client.ActivateTrigger(ctx, &artifacts.ActivateTriggerRequest{
		TriggerId: identifier,
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to activate trigger [%+v] err: %v", identifier, err)
		return err
	}
	logger.Debugf(ctx, "Activated trigger [%+v]", identifier)
	return nil
}

func (a *ArtifactRegistry) DeactivateTrigger(ctx context.Context, identifier *core.Identifier) error {
	if a == nil || a.Client == nil {
		logger.Debugf(ctx, "Artifact Client not configured, skipping deactivate [%+v]", identifier)
		return fmt.Errorf("artifact Client not configured")
	}
	_, err := a.Client.DeactivateTrigger(ctx, &artifacts.DeactivateTriggerRequest{
		TriggerId: identifier,
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to deactivate trigger [%+v] err: %v", identifier, err)
		return err
	}
	logger.Debugf(ctx, "Deactivated trigger [%+v]", identifier)
	return nil
}

func (a *ArtifactRegistry) GetClient() artifacts.ArtifactRegistryClient {
	if a == nil {
		return nil
	}
	return a.Client
}

// NewArtifactRegistry todo: update this to return error, and proper cfg handling.
// if nil, should either call the default config or return an error
func NewArtifactRegistry(ctx context.Context, connCfg *admin2.Config, _ ...grpc.DialOption) *ArtifactRegistry {

	if connCfg == nil {
		return &ArtifactRegistry{
			Client: nil,
		}
	}

	clients, err := admin2.NewClientsetBuilder().WithConfig(connCfg).Build(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to create Artifact Client")
		// too many calls to this function to update, just panic for now.
		panic(err)
	}

	return &ArtifactRegistry{
		Client: clients.ArtifactServiceClient(),
	}
}
