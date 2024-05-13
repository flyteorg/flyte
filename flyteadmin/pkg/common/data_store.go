package common

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
	errrs "github.com/pkg/errors"
	"github.com/samber/lo"
	"google.golang.org/api/googleapi"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/async"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func OffloadLiteralMap(ctx context.Context, storageClient *storage.DataStore, literalMap *core.LiteralMap, nestedKeys ...string) (storage.DataReference, error) {
	nestedKeys = lo.Filter(nestedKeys, func(key string, _ int) bool {
		return key != ""
	})
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
	return OffloadMessageWithRetryDelayAndAttempts(ctx, storageClient, literalMap, true, retryDelay, attempts, nestedKeyReference...)
}

func OffloadMessageWithRetryDelayAndAttempts(ctx context.Context, storageClient *storage.DataStore, message proto.Message, overwrite bool, retryDelay time.Duration, attempts int, nestedKeys ...string) (storage.DataReference, error) {
	if message == nil {
		return "", errors.NewFlyteAdminErrorf(codes.Internal, "Message to offload is nil")
	}

	uri, err := storageClient.ConstructReference(ctx, storageClient.GetBaseContainerFQN(ctx), nestedKeys...)
	if err != nil {
		return "", errors.NewFlyteAdminErrorf(codes.Internal, "Failed to construct data reference for [%+v] with err: %v", nestedKeys, err)
	}

	if !overwrite {
		metadata, err := storageClient.Head(ctx, uri)
		if err != nil {
			return "", errors.NewFlyteAdminErrorf(codes.Internal, "Failed to get metadata of data reference [%s] with err: %v", uri, err)
		}
		if metadata.Exists() {
			logger.Debugf(ctx, "Data reference [%s] already exists, skipping write", uri)
			return uri, nil
		}
	}

	err = async.RetryOnSpecificErrors(attempts, retryDelay, func() error {
		err = storageClient.WriteProtobuf(ctx, uri, storage.Options{}, message)
		return err
	}, isRetryableError)

	if err != nil {
		return "", errors.NewFlyteAdminErrorf(codes.Internal, "Failed to write protobuf for [%+v] with err: %v", nestedKeys, err)
	}

	return uri, nil
}

func OffloadConfigurationDocument(ctx context.Context, storageClient *storage.DataStore, document *admin.ConfigurationDocument, nestedKeys ...string) (storage.DataReference, error) {
	return OffloadConfigurationDocumentWithRetryDelayAndAttempts(ctx, storageClient, document, async.RetryDelay, 5, nestedKeys...)
}

func OffloadConfigurationDocumentWithRetryDelayAndAttempts(ctx context.Context, storageClient *storage.DataStore, document *admin.ConfigurationDocument, retryDelay time.Duration, attempts int, nestedKeys ...string) (storage.DataReference, error) {
	if document == nil {
		document = &admin.ConfigurationDocument{}
	}
	nestedKeyReference := []string{
		shared.Configuration,
	}
	nestedKeyReference = append(nestedKeyReference, nestedKeys...)
	return OffloadMessageWithRetryDelayAndAttempts(ctx, storageClient, document, false, retryDelay, attempts, nestedKeyReference...)
}

func isRetryableError(err error) bool {
	if e, ok := errrs.Cause(err).(*googleapi.Error); ok && e.Code == 409 {
		return true
	}
	return false
}
