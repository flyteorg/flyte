package datacatalog

import (
	"context"
	"crypto/x509"
	"time"

	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/pkg/errors"

	"github.com/lyft/flytestdlib/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/uuid"
)

const (
	taskVersionKey = "task-version"
	wfExecNameKey  = "execution-name"
)

var (
	_ catalog.Client = &CatalogClient{}
)

// This is the client that caches task executions to DataCatalog service.
type CatalogClient struct {
	client datacatalog.DataCatalogClient
}

// Helper method to retrieve a dataset that is associated with the task
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

// Helper method to retrieve an artifact by the tag
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

	return response.Artifact, nil
}

// Get the cached task execution from Catalog.
// These are the steps taken:
// - Verify there is a Dataset created for the Task
// - Lookup the Artifact that is tagged with the hash of the input values
// - The artifactData contains the literal values that serve as the task outputs
func (m *CatalogClient) Get(ctx context.Context, key catalog.Key) (io.OutputReader, error) {
	dataset, err := m.GetDataset(ctx, key)
	if err != nil {
		logger.Debugf(ctx, "DataCatalog failed to get dataset for ID %s, err: %+v", key.Identifier.String(), err)
		return nil, errors.Wrapf(err, "DataCatalog failed to get dataset for ID %s", key.Identifier.String())
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
		logger.Errorf(ctx, "DataCatalog failed to generate tag for inputs %+v, err: %+v", inputs, err)
		return nil, err
	}

	artifact, err := m.GetArtifactByTag(ctx, tag, dataset)
	if err != nil {
		logger.Debugf(ctx, "DataCatalog failed to get artifact by tag %+v, err: %+v", tag, err)
		return nil, err
	}
	logger.Debugf(ctx, "Artifact found %v from tag %v", artifact, tag)

	outputs, err := GenerateTaskOutputsFromArtifact(key.Identifier, key.TypedInterface, artifact)
	if err != nil {
		logger.Errorf(ctx, "DataCatalog failed to get outputs from artifact %+v, err: %+v", artifact.Id, err)
		return nil, err
	}

	logger.Infof(ctx, "Retrieved %v outputs from artifact %v, tag: %v", len(outputs.Literals), artifact.Id, tag)
	return ioutils.NewInMemoryOutputReader(outputs, nil), nil
}

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

func (m *CatalogClient) CreateArtifact(ctx context.Context, datasetID *datacatalog.DatasetID, outputs *core.LiteralMap, md *datacatalog.Metadata) (*datacatalog.Artifact, error) {
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
		Metadata: md,
	}

	createArtifactRequest := &datacatalog.CreateArtifactRequest{Artifact: cachedArtifact}
	_, err := m.client.CreateArtifact(ctx, createArtifactRequest)
	if err != nil {
		logger.Errorf(ctx, "Failed to create Artifact %+v, err: %v", cachedArtifact, err)
		return cachedArtifact, err
	}
	logger.Debugf(ctx, "Created artifact: %v, with %v outputs from execution %v", cachedArtifact.Id, len(artifactDataList))
	return cachedArtifact, nil
}

// Catalog the task execution as a cached Artifact. We associate an Artifact as the cached data by tagging the Artifact
// with the hash of the input values.
//
// The steps taken to cache an execution:
// - Ensure a Dataset exists for the Artifact. The Dataset represents the proj/domain/name/version of the task
// - Create an Artifact with the execution data that belongs to the dataset
// - Tag the Artifact with a hash generated by the input values
func (m *CatalogClient) Put(ctx context.Context, key catalog.Key, reader io.OutputReader, metadata catalog.Metadata) error {

	// Try creating the dataset in case it doesn't exist
	md := &datacatalog.Metadata{
		KeyMap: map[string]string{
			taskVersionKey: key.Identifier.Version,
		},
	}
	if metadata.WorkflowExecutionIdentifier != nil {
		md.KeyMap[wfExecNameKey] = metadata.WorkflowExecutionIdentifier.Name
	}

	datasetID, err := m.CreateDataset(ctx, key, md)
	if err != nil {
		return err
	}

	inputs := &core.LiteralMap{}
	outputs := &core.LiteralMap{}
	if key.TypedInterface.Inputs != nil && len(key.TypedInterface.Inputs.Variables) != 0 {
		retInputs, err := key.InputReader.Get(ctx)
		if err != nil {
			logger.Errorf(ctx, "DataCatalog failed to read inputs err: %s", err)
			return err
		}
		logger.Debugf(ctx, "DataCatalog read inputs")
		inputs = retInputs
	}

	if key.TypedInterface.Outputs != nil && len(key.TypedInterface.Outputs.Variables) != 0 {
		retOutputs, retErr, err := reader.Read(ctx)
		if err != nil {
			logger.Errorf(ctx, "DataCatalog failed to read outputs err: %s", err)
			return err
		}
		if retErr != nil {
			logger.Errorf(ctx, "DataCatalog failed to read outputs, err :%s", retErr.Message)
			return errors.Errorf("Failed to read outputs. EC: %s, Msg: %s", retErr.Code, retErr.Message)
		}
		logger.Debugf(ctx, "DataCatalog read outputs")
		outputs = retOutputs
	}

	// Create the artifact for the execution that belongs in the task
	cachedArtifact, err := m.CreateArtifact(ctx, datasetID, outputs, md)
	if err != nil {
		return errors.Wrapf(err, "failed to create dataset for ID %s", key.Identifier.String())
	}

	// Tag the artifact since it is the cached artifact
	tagName, err := GenerateArtifactTagName(ctx, inputs)
	if err != nil {
		logger.Errorf(ctx, "Failed to generate tag for artifact %+v, err: %+v", cachedArtifact.Id, err)
		return err
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
			return err
		}
	}

	return nil
}

// Create a new Datacatalog client for task execution caching
func NewDataCatalog(ctx context.Context, endpoint string, insecureConnection bool) (*CatalogClient, error) {
	var opts []grpc.DialOption

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

	retryInterceptor := grpc.WithUnaryInterceptor(grpcRetry.UnaryClientInterceptor(grpcOptions...))

	opts = append(opts, retryInterceptor)
	clientConn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return nil, err
	}

	client := datacatalog.NewDataCatalogClient(clientConn)

	return &CatalogClient{
		client: client,
	}, nil
}
