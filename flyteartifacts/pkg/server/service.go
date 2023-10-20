package server

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type CoreService struct {
	Storage   StorageInterface
	BlobStore BlobStoreInterface
	// TriggerHandler TriggerHandlerInterface
	// SearchHandler  SearchHandlerInterface
}

func (c *CoreService) CreateArtifact(ctx context.Context, request *artifact.CreateArtifactRequest) (*artifact.CreateArtifactResponse, error) {
	if request == nil {
		return nil, nil
	}

	artifactObj, err := models.CreateArtifactModelFromRequest(ctx, request.ArtifactKey, request.Spec, request.Version, request.Partitions, request.Tag, request.Spec.Principal)
	if err != nil {
		logger.Errorf(ctx, "Failed to validate Create request: %v", err)
		return nil, err
	}

	// Offload the metadata object before storing and add the offload location instead.
	if artifactObj.Spec.UserMetadata != nil {
		offloadLocation, err := c.BlobStore.OffloadArtifactCard(ctx,
			artifactObj.ArtifactId.ArtifactKey.Name, artifactObj.ArtifactId.Version, artifactObj.Spec.UserMetadata)
		if err != nil {
			logger.Errorf(ctx, "Failed to offload metadata: %v", err)
			return nil, err
		}
		artifactObj.OffloadedMetadata = offloadLocation.String()
	}

	created, err := c.Storage.CreateArtifact(ctx, artifactObj)
	if err != nil {
		logger.Errorf(ctx, "Failed to create artifact: %v", err)
		return nil, err
	}

	return &artifact.CreateArtifactResponse{Artifact: &created.Artifact}, nil
}

func (c *CoreService) GetArtifact(ctx context.Context, request *artifact.GetArtifactRequest) (*artifact.GetArtifactResponse, error) {
	if request == nil || request.Query == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	getResult, err := c.Storage.GetArtifact(ctx, *request.Query)
	if err != nil {
		logger.Errorf(ctx, "Failed to get artifact: %v", err)
		return nil, err
	}
	if request.Details && len(getResult.OffloadedMetadata) > 0 {
		card, err := c.BlobStore.RetrieveArtifactCard(ctx, storage.DataReference(getResult.OffloadedMetadata))
		if err != nil {
			logger.Errorf(ctx, "Failed to retrieve artifact card: %v", err)
			return nil, err
		}
		getResult.Artifact.GetSpec().UserMetadata = card
	}

	return &artifact.GetArtifactResponse{Artifact: &getResult.Artifact}, nil
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

// HandleCloudEvent is the stand-in for simple open-source handling of the event stream, rather than using
// a real
func (c *CoreService) HandleCloudEvent(ctx context.Context, request *artifact.CloudEventRequest) (*artifact.CloudEventResponse, error) {
	return &artifact.CloudEventResponse{}, nil
}

func NewCoreService(storage StorageInterface, blobStore BlobStoreInterface, _ promutils.Scope) CoreService {
	return CoreService{
		Storage:   storage,
		BlobStore: blobStore,
	}
}
