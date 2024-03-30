package common

import (
	"context"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"time"

	errrs "github.com/pkg/errors"
	"google.golang.org/api/googleapi"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/async"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func OffloadLiteralMap(ctx context.Context, storageClient *storage.DataStore, literalMap *core.LiteralMap, nestedKeys ...string) (storage.DataReference, error) {
	return OffloadLiteralMapWithRetryDelayAndAttempts(ctx, storageClient, literalMap, async.RetryDelay, 5, nestedKeys...)
}

func OffloadLiteralMapWithRetryDelayAndAttempts(ctx context.Context, storageClient *storage.DataStore, literalMap *core.LiteralMap, retryDelay time.Duration, attempts int, nestedKeys ...string) (storage.DataReference, error) {
	if literalMap == nil {
		literalMap = &core.LiteralMap{}
	}
	nestedKeyReference := []string{
		shared.Metadata,
	}
	nestedKeyReference = append(nestedKeyReference, nestedKeys...)
	uri, err := storageClient.ConstructReference(ctx, storageClient.GetBaseContainerFQN(ctx), nestedKeyReference...)
	if err != nil {
		return "", errors.NewFlyteAdminErrorf(codes.Internal, "Failed to construct data reference for [%+v] with err: %v", nestedKeys, err)
	}

	err = async.RetryOnSpecificErrors(attempts, retryDelay, func() error {
		err = storageClient.WriteProtobuf(ctx, uri, storage.Options{}, literalMap)
		return err
	}, isRetryableError)

	if err != nil {
		return "", errors.NewFlyteAdminErrorf(codes.Internal, "Failed to write protobuf for [%+v] with err: %v", nestedKeys, err)
	}

	return uri, nil
}

func OffloadOverrideAttributesDocument(ctx context.Context, storageClient *storage.DataStore, document *admin.Document, nestedKeys ...string) (storage.DataReference, error) {
	return OffloadOverrideAttributesDocumentWithRetryDelayAndAttempts(ctx, storageClient, document, async.RetryDelay, 5, nestedKeys...)
}

func OffloadOverrideAttributesDocumentWithRetryDelayAndAttempts(ctx context.Context, storageClient *storage.DataStore, document *admin.Document, retryDelay time.Duration, attempts int, nestedKeys ...string) (storage.DataReference, error) {
	if document == nil {
		document = &admin.Document{}
	}
	nestedKeyReference := []string{
		shared.OverrideAttributesDocument,
	}
	nestedKeyReference = append(nestedKeyReference, nestedKeys...)
	uri, err := storageClient.ConstructReference(ctx, storageClient.GetBaseContainerFQN(ctx), nestedKeyReference...)
	if err != nil {
		return "", errors.NewFlyteAdminErrorf(codes.Internal, "Failed to construct data reference for [%+v] with err: %v", nestedKeys, err)
	}

	err = async.RetryOnSpecificErrors(attempts, retryDelay, func() error {
		err = storageClient.WriteProtobuf(ctx, uri, storage.Options{}, document)
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
