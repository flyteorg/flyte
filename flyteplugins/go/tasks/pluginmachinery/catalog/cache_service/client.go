package cache_service

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	catalog "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/pbhash"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	cacheservicepb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cacheservice"
	cacheservicev2 "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cacheservice/v2"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cacheservice/v2/v2connect"
	corepb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

const (
	maxParamHashLength = 8
)

var emptyVariableMap = corepb.VariableMap{Variables: []*corepb.VariableEntry{}}

type outputPathReader interface {
	GetOutputPath() storage.DataReference
}

type Client struct {
	service          v2connect.CacheServiceClient
	store            *storage.DataStore
	maxCacheAge      time.Duration
	reservationCache sync.Map
}

var _ catalog.Client = (*Client)(nil)

func NewClient(httpClient connect.HTTPClient, store *storage.DataStore, baseURL string, maxCacheAge time.Duration) *Client {
	return &Client{
		service:     v2connect.NewCacheServiceClient(httpClient, baseURL),
		store:       store,
		maxCacheAge: maxCacheAge,
	}
}

func NewHTTPClient(store *storage.DataStore, baseURL string, maxCacheAge time.Duration) *Client {
	return NewClient(http.DefaultClient, store, baseURL, maxCacheAge)
}

func NewWithServiceClient(service v2connect.CacheServiceClient, store *storage.DataStore, maxCacheAge time.Duration) *Client {
	return &Client{
		service:     service,
		store:       store,
		maxCacheAge: maxCacheAge,
	}
}

func (c *Client) Get(ctx context.Context, key catalog.Key) (catalog.Entry, error) {
	cacheKey, err := buildCacheKey(ctx, key)
	if err != nil {
		return catalog.NewFailedCatalogEntry(catalog.NewStatus(corepb.CatalogCacheStatus_CACHE_LOOKUP_FAILURE, nil)), err
	}

	resp, err := c.service.Get(ctx, connect.NewRequest(&cacheservicev2.GetCacheRequest{
		BaseRequest: &cacheservicepb.GetCacheRequest{Key: cacheKey},
		Identifier:  newIdentifier(key.Identifier),
	}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return catalog.Entry{}, grpcstatus.Error(codes.NotFound, err.Error())
		}
		return catalog.NewFailedCatalogEntry(catalog.NewStatus(corepb.CatalogCacheStatus_CACHE_LOOKUP_FAILURE, nil)), err
	}

	output := resp.Msg.GetOutput()
	if output == nil || output.GetOutput() == nil || output.GetMetadata() == nil {
		return catalog.Entry{}, grpcstatus.Error(codes.Internal, "received malformed response from cache service")
	}
	if c.maxCacheAge > 0 {
		lastUpdatedAt := output.GetMetadata().GetLastUpdatedAt()
		if lastUpdatedAt == nil {
			return catalog.Entry{}, grpcstatus.Error(codes.Internal, "received cache metadata without last_updated_at")
		}
		if time.Since(lastUpdatedAt.AsTime()) > c.maxCacheAge {
			return catalog.Entry{}, grpcstatus.Error(codes.NotFound, "artifact over age limit")
		}
	}

	outputs, err := readCachedOutput(ctx, c.store, output)
	if err != nil {
		return catalog.Entry{}, err
	}

	source, err := getSourceFromMetadata(output.GetMetadata())
	if err != nil {
		return catalog.Entry{}, err
	}

	outputReader := ioutils.NewInMemoryOutputReader(outputs, nil, nil)
	status := catalog.NewStatus(corepb.CatalogCacheStatus_CACHE_HIT, newCatalogMetadata(output.GetMetadata(), source))
	return catalog.NewCatalogEntry(outputReader, status), nil
}

func (c *Client) GetOrExtendReservation(ctx context.Context, key catalog.Key, ownerID string, heartbeatInterval time.Duration) (*cacheservicepb.Reservation, error) {
	cacheKey, err := buildCacheKey(ctx, key)
	if err != nil {
		return nil, err
	}

	resp, err := c.service.GetOrExtendReservation(ctx, connect.NewRequest(&cacheservicev2.GetOrExtendReservationRequest{
		BaseRequest: &cacheservicepb.GetOrExtendReservationRequest{
			Key:               cacheKey,
			OwnerId:           ownerID,
			HeartbeatInterval: durationpb.New(heartbeatInterval),
		},
		Identifier: newIdentifier(key.Identifier),
	}))
	if err != nil {
		return nil, err
	}

	return resp.Msg.GetReservation(), nil
}

func (c *Client) Put(ctx context.Context, key catalog.Key, reader io.OutputReader, metadata catalog.Metadata) (catalog.Status, error) {
	return c.put(ctx, key, reader, metadata, false)
}

func (c *Client) Update(ctx context.Context, key catalog.Key, reader io.OutputReader, metadata catalog.Metadata) (catalog.Status, error) {
	return c.put(ctx, key, reader, metadata, true)
}

func (c *Client) ReleaseReservation(ctx context.Context, key catalog.Key, ownerID string) error {
	cacheKey, err := buildCacheKey(ctx, key)
	if err != nil {
		return err
	}

	_, err = c.service.ReleaseReservation(ctx, connect.NewRequest(&cacheservicev2.ReleaseReservationRequest{
		BaseRequest: &cacheservicepb.ReleaseReservationRequest{
			Key:     cacheKey,
			OwnerId: ownerID,
		},
		Identifier: newIdentifier(key.Identifier),
	}))
	return err
}

func (c *Client) GetReservationCache(ownerID string) catalog.ReservationCache {
	value, ok := c.reservationCache.Load(ownerID)
	if !ok {
		return catalog.ReservationCache{}
	}

	entry, ok := value.(catalog.ReservationCache)
	if !ok {
		return catalog.ReservationCache{}
	}

	return entry
}

func (c *Client) UpdateReservationCache(ownerID string, entry catalog.ReservationCache) {
	c.reservationCache.Store(ownerID, entry)
}

func (c *Client) put(ctx context.Context, key catalog.Key, reader io.OutputReader, metadata catalog.Metadata, overwrite bool) (catalog.Status, error) {
	cacheKey, err := buildCacheKey(ctx, key)
	if err != nil {
		return catalog.NewPutFailureStatus(&key), err
	}

	exists, err := reader.Exists(ctx)
	if err != nil {
		return catalog.NewPutFailureStatus(&key), err
	}
	if !exists {
		return catalog.NewPutFailureStatus(&key), grpcstatus.Errorf(codes.NotFound, "output file does not exist for %v", key.Identifier)
	}

	outputURI, err := resolveOutputURI(reader)
	if err != nil {
		return catalog.NewPutFailureStatus(&key), err
	}

	_, err = c.service.Put(ctx, connect.NewRequest(&cacheservicev2.PutCacheRequest{
		BaseRequest: &cacheservicepb.PutCacheRequest{
			Key: cacheKey,
			Output: &cacheservicepb.CachedOutput{
				Output:   &cacheservicepb.CachedOutput_OutputUri{OutputUri: outputURI},
				Metadata: newCacheMetadata(key, metadata),
			},
			Overwrite: &cacheservicepb.OverwriteOutput{
				Overwrite: overwrite,
			},
		},
		Identifier: newIdentifier(key.Identifier),
	}))
	if err != nil {
		return catalog.NewPutFailureStatus(&key), err
	}

	return catalog.NewStatus(corepb.CatalogCacheStatus_CACHE_POPULATED, newCatalogMetadata(newCacheMetadata(key, metadata), metadata.TaskExecutionIdentifier)), nil
}

func buildCacheKey(ctx context.Context, key catalog.Key) (string, error) {
	if key.Identifier == nil {
		return "", fmt.Errorf("catalog key identifier is required")
	}
	if key.TypedInterface == nil {
		return "", fmt.Errorf("catalog key typed interface is required")
	}
	if key.InputReader == nil {
		return "", fmt.Errorf("catalog key input reader is required")
	}

	identifierHash, err := catalog.HashIdentifierExceptVersion(ctx, key.Identifier)
	if err != nil {
		return "", err
	}

	signatureHash, err := generateInterfaceSignatureHash(ctx, key.TypedInterface)
	if err != nil {
		return "", err
	}

	inputsHash, err := hashInputs(ctx, key)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s-%s-%s", identifierHash, signatureHash, inputsHash, key.CacheVersion), nil
}

func hashInputs(ctx context.Context, key catalog.Key) (string, error) {
	inputs := &corepb.LiteralMap{}
	if key.TypedInterface.GetInputs() != nil {
		retInputs, err := key.InputReader.Get(ctx)
		if err != nil {
			return "", err
		}
		inputs = retInputs
	}

	return catalog.HashLiteralMap(ctx, inputs, key.CacheIgnoreInputVars)
}

func resolveOutputURI(reader io.OutputReader) (string, error) {
	pathReader, ok := reader.(outputPathReader)
	if !ok {
		return "", fmt.Errorf("output reader %T does not expose an output path", reader)
	}

	return pathReader.GetOutputPath().String(), nil
}

func newIdentifier(id *corepb.Identifier) *cacheservicev2.Identifier {
	if id == nil {
		return nil
	}

	return &cacheservicev2.Identifier{
		Org:     id.GetOrg(),
		Project: id.GetProject(),
		Domain:  id.GetDomain(),
	}
}

func newCacheMetadata(key catalog.Key, metadata catalog.Metadata) *cacheservicepb.Metadata {
	cacheMetadata := &cacheservicepb.Metadata{
		SourceIdentifier: key.Identifier,
		CreatedAt:        metadata.CreatedAt,
	}

	if metadata.TaskExecutionIdentifier != nil {
		cacheMetadata.KeyMap = &cacheservicepb.KeyMapMetadata{
			Values: map[string]string{
				TaskVersionKey:     metadata.TaskExecutionIdentifier.GetTaskId().GetVersion(),
				ExecProjectKey:     metadata.TaskExecutionIdentifier.GetNodeExecutionId().GetExecutionId().GetProject(),
				ExecDomainKey:      metadata.TaskExecutionIdentifier.GetNodeExecutionId().GetExecutionId().GetDomain(),
				ExecNameKey:        metadata.TaskExecutionIdentifier.GetNodeExecutionId().GetExecutionId().GetName(),
				ExecNodeIDKey:      metadata.TaskExecutionIdentifier.GetNodeExecutionId().GetNodeId(),
				ExecTaskAttemptKey: fmt.Sprintf("%d", metadata.TaskExecutionIdentifier.GetRetryAttempt()),
				ExecOrgKey:         metadata.TaskExecutionIdentifier.GetNodeExecutionId().GetExecutionId().GetOrg(),
			},
		}
	} else if metadata.NodeExecutionIdentifier != nil {
		cacheMetadata.KeyMap = &cacheservicepb.KeyMapMetadata{
			Values: map[string]string{
				ExecProjectKey: metadata.NodeExecutionIdentifier.GetExecutionId().GetProject(),
				ExecDomainKey:  metadata.NodeExecutionIdentifier.GetExecutionId().GetDomain(),
				ExecNameKey:    metadata.NodeExecutionIdentifier.GetExecutionId().GetName(),
				ExecNodeIDKey:  metadata.NodeExecutionIdentifier.GetNodeId(),
				ExecOrgKey:     metadata.NodeExecutionIdentifier.GetExecutionId().GetOrg(),
			},
		}
	}

	if cacheMetadata.CreatedAt == nil {
		cacheMetadata.CreatedAt = timestamppb.Now()
	}

	return cacheMetadata
}

func readCachedOutput(ctx context.Context, store *storage.DataStore, output *cacheservicepb.CachedOutput) (*corepb.LiteralMap, error) {
	switch typedOutput := output.GetOutput().(type) {
	case *cacheservicepb.CachedOutput_OutputLiterals:
		return typedOutput.OutputLiterals, nil
	case *cacheservicepb.CachedOutput_OutputUri:
		outputs := &corepb.LiteralMap{}
		if err := store.ReadProtobuf(ctx, storage.DataReference(typedOutput.OutputUri), outputs); err != nil {
			return nil, fmt.Errorf("failed to read output data from %q: %w", typedOutput.OutputUri, err)
		}
		return outputs, nil
	default:
		return nil, grpcstatus.Error(codes.Internal, "received malformed response from cache service")
	}
}

func getSourceFromMetadata(metadata *cacheservicepb.Metadata) (*corepb.TaskExecutionIdentifier, error) {
	if metadata == nil {
		return nil, fmt.Errorf("output does not have metadata")
	}
	if metadata.GetSourceIdentifier() == nil {
		return nil, fmt.Errorf("output does not have source identifier")
	}

	keyMap := metadata.GetKeyMap().GetValues()
	return &corepb.TaskExecutionIdentifier{
		TaskId: &corepb.Identifier{
			ResourceType: metadata.GetSourceIdentifier().GetResourceType(),
			Project:      metadata.GetSourceIdentifier().GetProject(),
			Domain:       metadata.GetSourceIdentifier().GetDomain(),
			Name:         metadata.GetSourceIdentifier().GetName(),
			Version:      getMetadataValue(keyMap, TaskVersionKey, "unknown"),
			Org:          metadata.GetSourceIdentifier().GetOrg(),
		},
		RetryAttempt: parseAttempt(getMetadataValue(keyMap, ExecTaskAttemptKey, "0")),
		NodeExecutionId: &corepb.NodeExecutionIdentifier{
			NodeId: getMetadataValue(keyMap, ExecNodeIDKey, "unknown"),
			ExecutionId: &corepb.WorkflowExecutionIdentifier{
				Project: getMetadataValue(keyMap, ExecProjectKey, metadata.GetSourceIdentifier().GetProject()),
				Domain:  getMetadataValue(keyMap, ExecDomainKey, metadata.GetSourceIdentifier().GetDomain()),
				Name:    getMetadataValue(keyMap, ExecNameKey, "unknown"),
				Org:     getMetadataValue(keyMap, ExecOrgKey, metadata.GetSourceIdentifier().GetOrg()),
			},
		},
	}, nil
}

func newCatalogMetadata(metadata *cacheservicepb.Metadata, source *corepb.TaskExecutionIdentifier) *corepb.CatalogMetadata {
	if metadata == nil {
		return nil
	}

	return &corepb.CatalogMetadata{
		DatasetId: metadata.GetSourceIdentifier(),
		SourceExecution: &corepb.CatalogMetadata_SourceTaskExecution{
			SourceTaskExecution: source,
		},
	}
}

func parseAttempt(value string) uint32 {
	var attempt uint32
	if _, err := fmt.Sscanf(value, "%d", &attempt); err != nil {
		return 0
	}
	return attempt
}

func generateInterfaceSignatureHash(ctx context.Context, resourceInterface *corepb.TypedInterface) (string, error) {
	taskInputs := &emptyVariableMap
	taskOutputs := &emptyVariableMap

	if resourceInterface.GetInputs() != nil && len(resourceInterface.GetInputs().GetVariables()) != 0 {
		taskInputs = resourceInterface.GetInputs()
	}

	if resourceInterface.GetOutputs() != nil && len(resourceInterface.GetOutputs().GetVariables()) != 0 {
		taskOutputs = resourceInterface.GetOutputs()
	}

	inputHash, err := pbhash.ComputeHash(ctx, taskInputs)
	if err != nil {
		return "", err
	}

	outputHash, err := pbhash.ComputeHash(ctx, taskOutputs)
	if err != nil {
		return "", err
	}

	inputHashString := truncateHash(base64.RawURLEncoding.EncodeToString(inputHash))
	outputHashString := truncateHash(base64.RawURLEncoding.EncodeToString(outputHash))
	return fmt.Sprintf("%s-%s", inputHashString, outputHashString), nil
}

func truncateHash(value string) string {
	if len(value) <= maxParamHashLength {
		return value
	}

	return value[:maxParamHashLength]
}
