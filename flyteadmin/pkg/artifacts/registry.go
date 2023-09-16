package artifacts

import (
	"fmt"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"google.golang.org/grpc"

	"context"
)

// ArtifactRegistry contains a client to talk to an Artifact service and has helper methods
type ArtifactRegistry struct {
	client artifact.ArtifactRegistryClient
}

func (a *ArtifactRegistry) RegisterArtifactProducer(ctx context.Context, id *core.Identifier, ti core.TypedInterface) {
	if a.client == nil {
		logger.Debugf(ctx, "Artifact client not configured, skipping registration for task [%+v]", id)
		return
	}
	ap := &artifact.ArtifactProducer{
		EntityId: id,
		Outputs:  ti.Outputs,
	}
	_, err := a.client.RegisterProducer(ctx, &artifact.RegisterProducerRequest{
		Producers: []*artifact.ArtifactProducer{ap},
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to register artifact producer for task [%+v] with err: %v", id, err)
	}
	logger.Debugf(ctx, "Registered artifact producer [%+v]", id)
}

func (a *ArtifactRegistry) RegisterArtifactConsumer(ctx context.Context, id *core.Identifier, pm core.ParameterMap) {
	if a.client == nil {
		logger.Debugf(ctx, "Artifact client not configured, skipping registration for consumer [%+v]", id)
		return
	}
	ac := &artifact.ArtifactConsumer{
		EntityId: id,
		Inputs:   &pm,
	}
	_, err := a.client.RegisterConsumer(ctx, &artifact.RegisterConsumerRequest{
		Consumers: []*artifact.ArtifactConsumer{ac},
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to register artifact consumer for entity [%+v] with err: %v", id, err)
	}
	logger.Debugf(ctx, "Registered artifact consumer [%+v]", id)
}

func (a *ArtifactRegistry) RegisterTrigger(ctx context.Context, plan *admin.LaunchPlan) error {
	if a.client == nil {
		logger.Debugf(ctx, "Artifact client not configured, skipping trigger [%+v]", plan)
		return fmt.Errorf("artifact client not configured")
	}
	_, err := a.client.CreateTrigger(ctx, &artifact.CreateTriggerRequest{
		TriggerLaunchPlan: plan,
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to register trigger for [%+v] with err: %v", plan.Id, err)
		return err
	}
	logger.Debugf(ctx, "Registered trigger for [%+v]", plan.Id)
	return nil
}

func (a *ArtifactRegistry) GetClient() artifact.ArtifactRegistryClient {
	return a.client
}

func NewArtifactRegistry(ctx context.Context, config *Config, opts ...grpc.DialOption) ArtifactRegistry {
	return ArtifactRegistry{
		client: InitializeArtifactClient(ctx, config, opts...),
	}
}
