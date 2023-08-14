package datacatalog

import (
	"context"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"
	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/uuid"
)

var (
	_ catalog.Client = &CatalogClient{}
)

// CatalogClient is the client that caches task executions to DataCatalog service.
type CatalogClient struct {
	client      datacatalog.DataCatalogClient
	maxCacheAge time.Duration
}

// GetDataset retrieves a dataset that is associated with the task represented by the provided catalog.Key.
func (m *CatalogClient) GetDataset(ctx context.Context, key catalog.Key) (*datacatalog.Dataset, error) {
	datasetID, err := GenerateDatasetIDForTask(ctx, key)
	if err != nil {
		return nil, err
	}
	logger.Debugf(ctx, "Get Dataset %v", datasetID)

	dsQuery := &datacatalog.GetDatasetRequest{
		Dataset: datasetID,
	}

	datasetResponse, err := m.client.GetDataset(ctx, dsQuery)
	if err != nil {
		return nil, err
	}

	return datasetResponse.Dataset, nil
}

// GetArtifactByTag retrieves an artifact using the provided tag and dataset.
func (m *CatalogClient) GetArtifactByTag(ctx context.Context, tagName string, dataset *datacatalog.Dataset) (*datacatalog.Artifact, error) {
	logger.Debugf(ctx, "Get Artifact by tag %v", tagName)
	artifactQuery := &datacatalog.GetArtifactRequest{
		Dataset: dataset.Id,
		QueryHandle: &datacatalog.GetArtifactRequest_TagName{
			TagName: tagName,
		},
	}
	response, err := m.client.GetArtifact(ctx, artifactQuery)
	if err != nil {
		return nil, err
	}

	// check artifact's age if the configuration specifies a max age
	if m.maxCacheAge > time.Duration(0) {
		artifact := response.Artifact
		createdAt, err := ptypes.Timestamp(artifact.CreatedAt)
		if err != nil {
			logger.Errorf(ctx, "DataCatalog Artifact has invalid createdAt %+v, err: %+v", artifact.CreatedAt, err)
			return nil, err
		}

		if time.Since(createdAt) > m.maxCacheAge {
			logger.Warningf(ctx, "Expired Cached Artifact %v created on %v, older than max age %v",
				artifact.Id, createdAt.String(), m.maxCacheAge)
			return nil, status.Error(codes.NotFound, "Artifact over age limit")
		}
	}

	return response.Artifact, nil
}

// Get the cached task execution from Catalog.
// These are the steps taken:
// - Verify there is a Dataset created for the Task
// - Lookup the Artifact that is tagged with the hash of the input values
// - The artifactData contains the literal values that serve as the task outputs
func (m *CatalogClient) Get(ctx context.Context, key catalog.Key) (catalog.Entry, error) {
	dataset, err := m.GetDataset(ctx, key)
	if err != nil {
		logger.Debugf(ctx, "DataCatalog failed to get dataset for ID %s, err: %+v", key.Identifier.String(), err)
		return catalog.Entry{}, errors.Wrapf(err, "DataCatalog failed to get dataset for ID %s", key.Identifier.String())
	}

	inputs := &core.LiteralMap{}
	if key.TypedInterface.Inputs != nil {
		retInputs, err := key.InputReader.Get(ctx)
		if err != nil {
			return catalog.Entry{}, errors.Wrap(err, "failed to read inputs when trying to query catalog")
		}
		inputs = retInputs
	}

	tag, err := GenerateArtifactTagName(ctx, inputs)
	if err != nil {
		logger.Errorf(ctx, "DataCatalog failed to generate tag for inputs %+v, err: %+v", inputs, err)
		return catalog.Entry{}, err
	}

	artifact, err := m.GetArtifactByTag(ctx, tag, dataset)
	if err != nil {
		logger.Debugf(ctx, "DataCatalog failed to get artifact by tag %+v, err: %+v", tag, err)
		return catalog.Entry{}, err
	}
	logger.Debugf(ctx, "Artifact found %v from tag %v", artifact, tag)

	var relevantTag *datacatalog.Tag
	if len(artifact.GetTags()) > 0 {
		// TODO should we look through all the tags to find the relevant one?
		relevantTag = artifact.GetTags()[0]
	}

	source, err := GetSourceFromMetadata(dataset.GetMetadata(), artifact.GetMetadata(), key.Identifier)
	if err != nil {
		return catalog.Entry{}, fmt.Errorf("failed to get source from metadata. Error: %w", err)
	}

	md := EventCatalogMetadata(dataset.GetId(), relevantTag, source)

	outputs, err := GenerateTaskOutputsFromArtifact(key.Identifier, key.TypedInterface, artifact)
	if err != nil {
		logger.Errorf(ctx, "DataCatalog failed to get outputs from artifact %+v, err: %+v", artifact.Id, err)
		return catalog.NewCatalogEntry(ioutils.NewInMemoryOutputReader(outputs, nil, nil), catalog.NewStatus(core.CatalogCacheStatus_CACHE_MISS, md)), err
	}

	logger.Infof(ctx, "Retrieved %v outputs from artifact %v, tag: %v", len(outputs.Literals), artifact.Id, tag)
	return catalog.NewCatalogEntry(ioutils.NewInMemoryOutputReader(outputs, nil, nil), catalog.NewStatus(core.CatalogCacheStatus_CACHE_HIT, md)), nil
}

// CreateDataset creates a Dataset in datacatalog including the associated metadata.
func (m *CatalogClient) CreateDataset(ctx context.Context, key catalog.Key, metadata *datacatalog.Metadata) (*datacatalog.DatasetID, error) {
	datasetID, err := GenerateDatasetIDForTask(ctx, key)
	if err != nil {
		logger.Errorf(ctx, "DataCatalog failed to generate dataset for ID: %s, err: %s", key.Identifier, err)
		return nil, err
	}

	newDataset := &datacatalog.Dataset{
		Id:       datasetID,
		Metadata: metadata,
	}

	_, err = m.client.CreateDataset(ctx, &datacatalog.CreateDatasetRequest{Dataset: newDataset})
	if err != nil {
		logger.Debugf(ctx, "Create dataset %v return err %v", datasetID, err)
		if status.Code(err) == codes.AlreadyExists {
			logger.Debugf(ctx, "Create Dataset for ID %s already exists", key.Identifier)
		} else {
			logger.Errorf(ctx, "Unable to create dataset %s, err: %s", datasetID, err)
			return nil, err
		}
	}

	return datasetID, nil
}

// prepareInputsAndOutputs reads the inputs and outputs of a task and returns them as core.LiteralMaps to be consumed by datacatalog.
func (m *CatalogClient) prepareInputsAndOutputs(ctx context.Context, key catalog.Key, reader io.OutputReader) (inputs *core.LiteralMap, outputs *core.LiteralMap, err error) {
	inputs = &core.LiteralMap{}
	outputs = &core.LiteralMap{}
	if key.TypedInterface.Inputs != nil && len(key.TypedInterface.Inputs.Variables) != 0 {
		retInputs, err := key.InputReader.Get(ctx)
		if err != nil {
			logger.Errorf(ctx, "DataCatalog failed to read inputs err: %s", err)
			return nil, nil, err
		}
		logger.Debugf(ctx, "DataCatalog read inputs")
		inputs = retInputs
	}

	if key.TypedInterface.Outputs != nil && len(key.TypedInterface.Outputs.Variables) != 0 {
		retOutputs, retErr, err := reader.Read(ctx)
		if err != nil {
			logger.Errorf(ctx, "DataCatalog failed to read outputs err: %s", err)
			return nil, nil, err
		}
		if retErr != nil {
			logger.Errorf(ctx, "DataCatalog failed to read outputs, err :%s", retErr.Message)
			return nil, nil, errors.Errorf("Failed to read outputs. EC: %s, Msg: %s", retErr.Code, retErr.Message)
		}
		logger.Debugf(ctx, "DataCatalog read outputs")
		outputs = retOutputs
	}

	return inputs, outputs, nil
}

// CreateArtifact creates an Artifact in datacatalog including its associated ArtifactData and tags it with a hash of
// the provided input values for retrieval.
func (m *CatalogClient) CreateArtifact(ctx context.Context, key catalog.Key, datasetID *datacatalog.DatasetID, inputs *core.LiteralMap, outputs *core.LiteralMap, metadata catalog.Metadata) (catalog.Status, error) {
	logger.Debugf(ctx, "Creating artifact for key %+v, dataset %+v and execution %+v", key, datasetID, metadata)

	// Create the artifact for the execution that belongs in the task
	artifactDataList := make([]*datacatalog.ArtifactData, 0, len(outputs.Literals))
	for name, value := range outputs.Literals {
		artifactData := &datacatalog.ArtifactData{
			Name:  name,
			Value: value,
		}
		artifactDataList = append(artifactDataList, artifactData)
	}

	cachedArtifact := &datacatalog.Artifact{
		Id:       string(uuid.NewUUID()),
		Dataset:  datasetID,
		Data:     artifactDataList,
		Metadata: GetArtifactMetadataForSource(metadata.TaskExecutionIdentifier),
	}

	createArtifactRequest := &datacatalog.CreateArtifactRequest{Artifact: cachedArtifact}
	_, err := m.client.CreateArtifact(ctx, createArtifactRequest)
	if err != nil {
		logger.Errorf(ctx, "Failed to create Artifact %+v, err: %v", cachedArtifact, err)
		return catalog.Status{}, err
	}
	logger.Debugf(ctx, "Created artifact: %v, with %v outputs from execution %+v", cachedArtifact.Id, len(artifactDataList), metadata)

	// Tag the artifact since it is the cached artifact
	tagName, err := GenerateArtifactTagName(ctx, inputs)
	if err != nil {
		logger.Errorf(ctx, "Failed to generate tag for artifact %+v, err: %+v", cachedArtifact.Id, err)
		return catalog.Status{}, err
	}
	logger.Infof(ctx, "Cached exec tag: %v, task: %v", tagName, key.Identifier)

	// TODO: We should create the artifact + tag in a transaction when the service supports that
	tag := &datacatalog.Tag{
		Name:       tagName,
		Dataset:    datasetID,
		ArtifactId: cachedArtifact.Id,
	}
	_, err = m.client.AddTag(ctx, &datacatalog.AddTagRequest{Tag: tag})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			logger.Warnf(ctx, "Tag %v already exists for Artifact %v (idempotent)", tagName, cachedArtifact.Id)
		} else {
			logger.Errorf(ctx, "Failed to add tag %+v for artifact %+v, err: %+v", tagName, cachedArtifact.Id, err)
			return catalog.Status{}, err
		}
	}

	logger.Debugf(ctx, "Successfully created artifact %+v for key %+v, dataset %+v and execution %+v", cachedArtifact, key, datasetID, metadata)
	return catalog.NewStatus(core.CatalogCacheStatus_CACHE_POPULATED, EventCatalogMetadata(datasetID, tag, nil)), nil
}

// UpdateArtifact overwrites the ArtifactData of an existing artifact with the provided data in datacatalog.
func (m *CatalogClient) UpdateArtifact(ctx context.Context, key catalog.Key, datasetID *datacatalog.DatasetID, inputs *core.LiteralMap, outputs *core.LiteralMap, metadata catalog.Metadata) (catalog.Status, error) {
	logger.Debugf(ctx, "Updating artifact for key %+v, dataset %+v and execution %+v", key, datasetID, metadata)

	artifactDataList := make([]*datacatalog.ArtifactData, 0, len(outputs.Literals))
	for name, value := range outputs.Literals {
		artifactData := &datacatalog.ArtifactData{
			Name:  name,
			Value: value,
		}
		artifactDataList = append(artifactDataList, artifactData)
	}

	tagName, err := GenerateArtifactTagName(ctx, inputs)
	if err != nil {
		logger.Errorf(ctx, "Failed to generate artifact tag name for key %+v, dataset %+v and execution %+v, err: %+v", key, datasetID, metadata, err)
		return catalog.Status{}, err
	}

	updateArtifactRequest := &datacatalog.UpdateArtifactRequest{
		Dataset:     datasetID,
		QueryHandle: &datacatalog.UpdateArtifactRequest_TagName{TagName: tagName},
		Data:        artifactDataList,
	}
	resp, err := m.client.UpdateArtifact(ctx, updateArtifactRequest)
	if err != nil {
		logger.Errorf(ctx, "Failed to update artifact for key %+v, dataset %+v and execution %+v, err: %v", key, datasetID, metadata, err)
		return catalog.Status{}, err
	}

	tag := &datacatalog.Tag{
		Name:       tagName,
		Dataset:    datasetID,
		ArtifactId: resp.GetArtifactId(),
	}

	source, err := GetSourceFromMetadata(GetDatasetMetadataForSource(metadata.TaskExecutionIdentifier), GetArtifactMetadataForSource(metadata.TaskExecutionIdentifier), key.Identifier)
	if err != nil {
		return catalog.Status{}, fmt.Errorf("failed to get source from metadata. Error: %w", err)
	}

	logger.Debugf(ctx, "Successfully updated artifact with ID %v and %d outputs for key %+v, dataset %+v and execution %+v", tag.ArtifactId, len(artifactDataList), key, datasetID, metadata)
	return catalog.NewStatus(core.CatalogCacheStatus_CACHE_POPULATED, EventCatalogMetadata(datasetID, tag, source)), nil
}

// Put stores the result of a task execution as a cached Artifact and associates it with the data by tagging it with
// the hash of the input values.
// The CatalogClient will ensure a dataset exists for the Artifact to be created. A Dataset represents the
// project/domain/name/version of the task executed.
// Lastly, CatalogClient will create an Artifact tagged with the input value hash and store the provided execution data.
func (m *CatalogClient) Put(ctx context.Context, key catalog.Key, reader io.OutputReader, metadata catalog.Metadata) (catalog.Status, error) {
	// Ensure dataset exists, idempotent operations. Populate Metadata for later recovery
	datasetID, err := m.CreateDataset(ctx, key, GetDatasetMetadataForSource(metadata.TaskExecutionIdentifier))
	if err != nil {
		return catalog.Status{}, err
	}

	inputs, outputs, err := m.prepareInputsAndOutputs(ctx, key, reader)
	if err != nil {
		return catalog.Status{}, err
	}

	return m.CreateArtifact(ctx, key, datasetID, inputs, outputs, metadata)
}

// Update stores the result of a task execution as a cached Artifact, overwriting any already stored data from a previous
// execution.
// The CatalogClient will ensure the referenced dataset exists and will silently create a new Artifact if the referenced
// key does not exist in datacatalog yet.
// After the operation succeeds, an artifact with the given key and data will be stored in catalog and a tag with the
// has of the input values will exist.
func (m *CatalogClient) Update(ctx context.Context, key catalog.Key, reader io.OutputReader, metadata catalog.Metadata) (catalog.Status, error) {
	// Ensure dataset exists, idempotent operations. Populate Metadata for later recovery
	datasetID, err := m.CreateDataset(ctx, key, GetDatasetMetadataForSource(metadata.TaskExecutionIdentifier))
	if err != nil {
		return catalog.Status{}, err
	}

	inputs, outputs, err := m.prepareInputsAndOutputs(ctx, key, reader)
	if err != nil {
		return catalog.Status{}, err
	}

	catalogStatus, err := m.UpdateArtifact(ctx, key, datasetID, inputs, outputs, metadata)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// No existing artifact found (e.g. initial execution of task with overwrite flag already set),
			// silently ignore error and create artifact instead to make overwriting an idempotent operation.
			logger.Debugf(ctx, "Artifact %+v for dataset %+v does not exist while updating, creating instead", key, datasetID)
			return m.CreateArtifact(ctx, key, datasetID, inputs, outputs, metadata)
		}

		logger.Errorf(ctx, "Failed to update artifact %+v for dataset %+v: %v", key, datasetID, err)
		return catalog.Status{}, err
	}

	logger.Debugf(ctx, "Successfully updated artifact %+v for dataset %+v", key, datasetID)
	return catalogStatus, nil
}

// GetOrExtendReservation attempts to get a reservation for the cachable task. If you have
// previously acquired a reservation it will be extended. If another entity holds the reservation
// that is returned.
func (m *CatalogClient) GetOrExtendReservation(ctx context.Context, key catalog.Key, ownerID string, heartbeatInterval time.Duration) (*datacatalog.Reservation, error) {
	datasetID, err := GenerateDatasetIDForTask(ctx, key)
	if err != nil {
		return nil, err
	}

	inputs := &core.LiteralMap{}
	if key.TypedInterface.Inputs != nil {
		retInputs, err := key.InputReader.Get(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read inputs when trying to query catalog")
		}
		inputs = retInputs
	}

	tag, err := GenerateArtifactTagName(ctx, inputs)
	if err != nil {
		return nil, err
	}

	reservationQuery := &datacatalog.GetOrExtendReservationRequest{
		ReservationId: &datacatalog.ReservationID{
			DatasetId: datasetID,
			TagName:   tag,
		},
		OwnerId:           ownerID,
		HeartbeatInterval: ptypes.DurationProto(heartbeatInterval),
	}

	response, err := m.client.GetOrExtendReservation(ctx, reservationQuery)
	if err != nil {
		return nil, err
	}

	return response.Reservation, nil
}

// ReleaseReservation attempts to release a reservation for a cachable task. If the reservation
// does not exist (e.x. it never existed or has been acquired by another owner) then this call
// still succeeds.
func (m *CatalogClient) ReleaseReservation(ctx context.Context, key catalog.Key, ownerID string) error {
	datasetID, err := GenerateDatasetIDForTask(ctx, key)
	if err != nil {
		return err
	}

	inputs := &core.LiteralMap{}
	if key.TypedInterface.Inputs != nil {
		retInputs, err := key.InputReader.Get(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to read inputs when trying to query catalog")
		}
		inputs = retInputs
	}

	tag, err := GenerateArtifactTagName(ctx, inputs)
	if err != nil {
		return err
	}

	reservationQuery := &datacatalog.ReleaseReservationRequest{
		ReservationId: &datacatalog.ReservationID{
			DatasetId: datasetID,
			TagName:   tag,
		},
		OwnerId: ownerID,
	}

	_, err = m.client.ReleaseReservation(ctx, reservationQuery)
	if err != nil {
		return err
	}

	return nil
}

// NewDataCatalog creates a new Datacatalog client for task execution caching
func NewDataCatalog(ctx context.Context, endpoint string, insecureConnection bool, maxCacheAge time.Duration, useAdminAuth bool, defaultServiceConfig string, authOpt ...grpc.DialOption) (*CatalogClient, error) {
	var opts []grpc.DialOption
	if useAdminAuth && authOpt != nil {
		opts = append(opts, authOpt...)
	}

	grpcOptions := []grpcRetry.CallOption{
		grpcRetry.WithBackoff(grpcRetry.BackoffLinear(100 * time.Millisecond)),
		grpcRetry.WithCodes(codes.DeadlineExceeded, codes.Unavailable, codes.Canceled),
		grpcRetry.WithMax(5),
	}

	if insecureConnection {
		logger.Debug(ctx, "Establishing insecure connection to DataCatalog")
		opts = append(opts, grpc.WithInsecure())
	} else {
		logger.Debug(ctx, "Establishing secure connection to DataCatalog")
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}

		creds := credentials.NewClientTLSFromCert(pool, "")
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	if defaultServiceConfig != "" {
		opts = append(opts, grpc.WithDefaultServiceConfig(defaultServiceConfig))
	}

	retryInterceptor := grpcRetry.UnaryClientInterceptor(grpcOptions...)

	opts = append(opts, grpc.WithChainUnaryInterceptor(grpcPrometheus.UnaryClientInterceptor,
		retryInterceptor))
	clientConn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return nil, err
	}

	client := datacatalog.NewDataCatalogClient(clientConn)

	return &CatalogClient{
		client:      client,
		maxCacheAge: maxCacheAge,
	}, nil
}
