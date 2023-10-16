package server

import (
	"context"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

type CoreService struct {
	Storage StorageInterface
	// TriggerHandler TriggerHandlerInterface
	// SearchHandler  SearchHandlerInterface
}

func (c *CoreService) CreateArtifact(ctx context.Context, request *artifact.CreateArtifactRequest) (*artifact.CreateArtifactResponse, error) {
	if request == nil {
		return nil, nil
	}
	model, err := CreateArtifactModelFrom(request.ArtifactKey, request.Spec, request.Version, request.Partitions, request.Tag, request.Spec.Principal)
	if err != nil {
		logger.Errorf(ctx, "Failed to create artifact model from request: %v", err)
		return nil, err
	}

	created, err := c.Storage.CreateArtifact(ctx, &model)
	if err != nil {
		logger.Errorf(ctx, "Failed to create artifact: %v", err)
		return nil, err
	}
	idl := FromModelToIdl(created)

	return &artifact.CreateArtifactResponse{Artifact: &idl}, nil
}

func (c *CoreService) GetArtifact(ctx context.Context, request *artifact.GetArtifactRequest) (*artifact.GetArtifactResponse, error) {
	return &artifact.GetArtifactResponse{}, nil
}

func (c *CoreService) CreateTrigger(ctx context.Context, request *artifact.CreateTriggerRequest) (*artifact.CreateTriggerResponse, error) {
	return &artifact.CreateTriggerResponse{}, nil
}

func (c *CoreService) DeleteTrigger(ctx context.Context, request *artifact.DeleteTriggerRequest) (*artifact.DeleteTriggerResponse, error) {
	return &artifact.DeleteTriggerResponse{}, nil
}

func (c *CoreService) AddTag(ctx context.Context, request *artifact.AddTagRequest) (*artifact.AddTagResponse, error) {
	return &artifact.AddTagResponse{}, nil
}

func (c *CoreService) RegisterProducer(ctx context.Context, request *artifact.RegisterProducerRequest) (*artifact.RegisterResponse, error) {
	return &artifact.RegisterResponse{}, nil
}

func (c *CoreService) RegisterConsumer(ctx context.Context, request *artifact.RegisterConsumerRequest) (*artifact.RegisterResponse, error) {
	return &artifact.RegisterResponse{}, nil
}

func (c *CoreService) SearchArtifacts(ctx context.Context, request *artifact.SearchArtifactsRequest) (*artifact.SearchArtifactsResponse, error) {
	return &artifact.SearchArtifactsResponse{}, nil
}
