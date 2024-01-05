package server

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/lib"

	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type CoreService struct {
	Storage   StorageInterface
	BlobStore BlobStoreInterface
}

// This string is a tracker basically that will be installed in the metadata of the literal. See the ArtifactKey constant for more information.
func (c *CoreService) getTrackingString(request *artifact.CreateArtifactRequest) string {
	ak := request.ArtifactKey
	t := fmt.Sprintf("%s/%s/%s@%s", ak.Project, ak.Domain, ak.Name, request.Version)

	return t
}

func (c *CoreService) CreateArtifact(ctx context.Context, request *artifact.CreateArtifactRequest) (*artifact.CreateArtifactResponse, error) {

	// todo: move one layer higher to server.go
	if request == nil || request.GetArtifactKey() == nil {
		logger.Errorf(ctx, "Ignoring nil or partially nil request")
		return nil, nil
	}

	if request.GetSpec().GetValue().Metadata == nil {
		request.GetSpec().GetValue().Metadata = make(map[string]string, 1)
	}
	trackingStr := c.getTrackingString(request)
	request.GetSpec().GetValue().Metadata[lib.ArtifactKey] = trackingStr

	artifactObj, err := models.CreateArtifactModelFromRequest(ctx, request.ArtifactKey, request.Spec, request.Version, request.Partitions, request.Tag, request.Source)
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
	// Create the new trigger object.
	// Mark all older versions of the trigger as inactive.

	// trigger handler create trigger(storage layer)
	serviceTrigger, err := models.CreateTriggerModelFromRequest(ctx, request)
	if err != nil {
		logger.Errorf(ctx, "Failed to create a valid Trigger from create request: %v with err %v", request, err)
		return nil, err
	}

	createdTrigger, err := c.Storage.CreateTrigger(ctx, serviceTrigger)
	if err != nil {
		logger.Errorf(ctx, "Failed to create trigger: %v", err)
	}
	logger.Infof(ctx, "Created trigger: %+v", createdTrigger)

	return &artifact.CreateTriggerResponse{}, nil
}

func (c *CoreService) DeleteTrigger(ctx context.Context, request *artifact.DeleteTriggerRequest) (*artifact.DeleteTriggerResponse, error) {
	// Todo: gatepr - This needs to be implemented before merging.
	return &artifact.DeleteTriggerResponse{}, nil
}

func (c *CoreService) AddTag(ctx context.Context, request *artifact.AddTagRequest) (*artifact.AddTagResponse, error) {
	// Holding off on implementing for a while.
	return &artifact.AddTagResponse{}, nil
}

func (c *CoreService) RegisterProducer(ctx context.Context, request *artifact.RegisterProducerRequest) (*artifact.RegisterResponse, error) {
	// These are lineage endpoints slated for future work
	return &artifact.RegisterResponse{}, nil
}

func (c *CoreService) RegisterConsumer(ctx context.Context, request *artifact.RegisterConsumerRequest) (*artifact.RegisterResponse, error) {
	// These are lineage endpoints slated for future work
	return &artifact.RegisterResponse{}, nil
}

func (c *CoreService) SearchArtifacts(ctx context.Context, request *artifact.SearchArtifactsRequest) (*artifact.SearchArtifactsResponse, error) {
	logger.Infof(ctx, "SearchArtifactsRequest: %+v", request)
	found, continuationToken, err := c.Storage.SearchArtifacts(ctx, *request)
	if err != nil {
		logger.Errorf(ctx, "Failed search  [%+v]: %v", request, err)
		return nil, err
	}
	if len(found) == 0 {
		return &artifact.SearchArtifactsResponse{}, fmt.Errorf("no artifacts found")
	}
	var as []*artifact.Artifact
	for _, m := range found {
		as = append(as, &m.Artifact)
	}
	return &artifact.SearchArtifactsResponse{
		Artifacts: as,
		Token:     continuationToken,
	}, nil
}

func (c *CoreService) SetExecutionInputs(ctx context.Context, req *artifact.ExecutionInputsRequest) (*artifact.ExecutionInputsResponse, error) {
	err := c.Storage.SetExecutionInputs(ctx, req)

	return &artifact.ExecutionInputsResponse{}, err
}

func (c *CoreService) FindByWorkflowExec(ctx context.Context, request *artifact.FindByWorkflowExecRequest) (*artifact.SearchArtifactsResponse, error) {
	logger.Infof(ctx, "FindByWorkflowExecRequest: %+v", request)

	res, err := c.Storage.FindByWorkflowExec(ctx, request)
	if err != nil {
		logger.Errorf(ctx, "Failed to find artifacts by workflow execution: %v", err)
		return nil, err
	}
	var as []*artifact.Artifact
	for _, m := range res {
		as = append(as, &m.Artifact)
	}
	resp := artifact.SearchArtifactsResponse{
		Artifacts: as,
	}

	return &resp, nil
}

func NewCoreService(storage StorageInterface, blobStore BlobStoreInterface, _ promutils.Scope) CoreService {
	return CoreService{
		Storage:   storage,
		BlobStore: blobStore,
	}
}
