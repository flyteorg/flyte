package common

import (
	"context"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	errrs "github.com/pkg/errors"
	"github.com/samber/lo"
	"google.golang.org/api/googleapi"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/async"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

//go:generate mockery -name=DatastoreClient -output=../mocks -case=underscore

// DatastoreClient is a light wrapper over the stdlib DataStore with only a subset of methods exposed.
type DatastoreClient interface {
	OffloadWorkflowClosure(ctx context.Context, closure *admin.WorkflowClosure, identifier *core.Identifier) (
		storage.DataReference, error)
	OffloadCompiledWorkflowClosure(ctx context.Context, closure *core.CompiledWorkflowClosure, identifier *core.Identifier, prefix string) (
		storage.DataReference, error)
	OffloadExecutionLiteralMap(
		ctx context.Context, literalMap *core.LiteralMap, identifier *core.WorkflowExecutionIdentifier, prefix string) (
		storage.DataReference, error)
	OffloadNodeExecutionLiteralMap(
		ctx context.Context, literalMap *core.LiteralMap, identifier *core.NodeExecutionIdentifier, prefix string) (
		storage.DataReference, error)
	OffloadTaskExecutionLiteralMap(
		ctx context.Context, literalMap *core.LiteralMap, identifier *core.TaskExecutionIdentifier, prefix string) (
		storage.DataReference, error)
	ReadProtobuf(ctx context.Context, reference storage.DataReference, msg proto.Message) error
	Head(ctx context.Context, reference storage.DataReference) (storage.Metadata, error)
}

// Implements DatastoreClient
type datastoreClient struct {
	storagePrefix []string
	datastore     *storage.DataStore
}

// Formats the blob store path prefix for offloading a literal map for a flyte inventory item (task, workflow, launch plan).
func getIdentifierStorageKeys(identifier *core.Identifier) []string {
	return []string{identifier.Org, identifier.Project, identifier.Domain, identifier.Name, identifier.Name}
}

// Formats the blob store path prefix for offloading a literal map for a specific workflow execution.
func getWorkflowExecutionIdentifierStorageKeys(identifier *core.WorkflowExecutionIdentifier) []string {
	return []string{identifier.Org, identifier.Project, identifier.Domain, identifier.Name}
}

// Formats the blob store path prefix for offloading a literal map for a specific node execution.
func getNodeExecutionIdentifierStorageKeys(identifier *core.NodeExecutionIdentifier) []string {
	workflowExecutionStorageKeys := getWorkflowExecutionIdentifierStorageKeys(identifier.ExecutionId)
	return append(workflowExecutionStorageKeys, identifier.NodeId)
}

func (c datastoreClient) offloadLiteralMap(ctx context.Context, literalMap *core.LiteralMap, nestedKeys ...string) (storage.DataReference, error) {
	if literalMap == nil {
		literalMap = &core.LiteralMap{}
	}

	return c.offloadMessage(ctx, literalMap, nestedKeys...)
}

func (c datastoreClient) offloadMessage(ctx context.Context, msg proto.Message, nestedKeys ...string) (storage.DataReference, error) {
	nestedKeys = lo.Filter(nestedKeys, func(key string, _ int) bool {
		return lo.IsNotEmpty(key)
	})
	nestedKeys = append(c.storagePrefix, nestedKeys...)
	return c.offloadWithRetryDelayAndAttempts(ctx, msg, async.RetryDelay, 5, nestedKeys...)
}

func (c datastoreClient) offloadWithRetryDelayAndAttempts(ctx context.Context, msg proto.Message, retryDelay time.Duration, attempts int, nestedKeys ...string) (storage.DataReference, error) {
	nestedKeyReference := []string{
		shared.Metadata,
	}
	nestedKeyReference = append(nestedKeyReference, nestedKeys...)
	uri, err := c.datastore.ConstructReference(ctx, c.datastore.GetBaseContainerFQN(ctx), nestedKeyReference...)
	if err != nil {
		return "", errors.NewFlyteAdminErrorf(codes.Internal, "Failed to construct data reference for [%+v] with err: %v", nestedKeys, err)
	}

	err = async.RetryOnSpecificErrors(attempts, retryDelay, func() error {
		err = c.datastore.WriteProtobuf(ctx, uri, storage.Options{}, msg)
		return err
	}, isRetryableError)

	if err != nil {
		return "", errors.NewFlyteAdminErrorf(codes.Internal, "Failed to write protobuf for [%+v] with err: %v", nestedKeys, err)
	}

	return uri, nil
}

func isRetryableError(err error) bool {
	if e, ok := errrs.Cause(err).(*googleapi.Error); ok && e.Code == 409 {
		return true
	}
	return false
}

func (c datastoreClient) OffloadWorkflowClosure(ctx context.Context, closure *admin.WorkflowClosure, identifier *core.Identifier) (
	storage.DataReference, error) {
	workflowIdentifierStorageKeys := getIdentifierStorageKeys(identifier)
	return c.offloadMessage(ctx, closure, workflowIdentifierStorageKeys...)
}

func (c datastoreClient) OffloadCompiledWorkflowClosure(
	ctx context.Context, closure *core.CompiledWorkflowClosure, identifier *core.Identifier, prefix string) (storage.DataReference, error) {
	workflowIdentifierStorageKeys := getIdentifierStorageKeys(identifier)
	nestedKeys := append(workflowIdentifierStorageKeys, prefix)
	return c.offloadMessage(ctx, closure, nestedKeys...)
}

func (c datastoreClient) OffloadExecutionLiteralMap(
	ctx context.Context, literalMap *core.LiteralMap, identifier *core.WorkflowExecutionIdentifier, prefix string) (
	storage.DataReference, error) {
	workflowExecutionStorageKeys := getWorkflowExecutionIdentifierStorageKeys(identifier)
	nestedKeys := append(workflowExecutionStorageKeys, prefix)
	return c.offloadLiteralMap(ctx, literalMap, nestedKeys...)

}

func (c datastoreClient) OffloadNodeExecutionLiteralMap(
	ctx context.Context, literalMap *core.LiteralMap, identifier *core.NodeExecutionIdentifier, prefix string) (
	storage.DataReference, error) {
	nodeExecutionStorageKeys := getNodeExecutionIdentifierStorageKeys(identifier)
	nestedKeys := append(nodeExecutionStorageKeys, prefix)
	return c.offloadLiteralMap(ctx, literalMap, nestedKeys...)
}

func (c datastoreClient) OffloadTaskExecutionLiteralMap(
	ctx context.Context, literalMap *core.LiteralMap, identifier *core.TaskExecutionIdentifier, prefix string) (
	storage.DataReference, error) {
	nodeExecutionStorageKeys := getNodeExecutionIdentifierStorageKeys(identifier.NodeExecutionId)
	taskStorageKeys := getIdentifierStorageKeys(identifier.TaskId)
	nestedKeys := append(nodeExecutionStorageKeys, taskStorageKeys...)
	retryAttempt := strconv.FormatUint(uint64(identifier.RetryAttempt), 10)
	nestedKeys = append(nestedKeys, retryAttempt)
	nestedKeys = append(nestedKeys, prefix)
	return c.offloadLiteralMap(ctx, literalMap, nestedKeys...)
}

func (c datastoreClient) ReadProtobuf(ctx context.Context, reference storage.DataReference, msg proto.Message) error {
	return c.datastore.ReadProtobuf(ctx, reference, msg)
}

func (c datastoreClient) Head(ctx context.Context, reference storage.DataReference) (storage.Metadata, error) {
	return c.datastore.Head(ctx, reference)
}

func NewClient(storagePrefix []string, datastore *storage.DataStore) DatastoreClient {
	return &datastoreClient{
		storagePrefix: storagePrefix,
		datastore:     datastore,
	}
}
