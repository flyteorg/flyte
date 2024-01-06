package common

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
	errrs "github.com/pkg/errors"
	"google.golang.org/api/googleapi"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/async"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func OffloadData(ctx context.Context, storageClient *storage.DataStore, msg proto.Message, nestedKeys ...string) (storage.DataReference, error) {
	return OffloadDataWithRetryDelayAndAttempts(ctx, storageClient, msg, async.RetryDelay, 5, nestedKeys...)
}

func OffloadDataWithRetryDelayAndAttempts(ctx context.Context, storageClient *storage.DataStore, msg proto.Message, retryDelay time.Duration, attempts int, nestedKeys ...string) (storage.DataReference, error) {
	if msg == nil {
		msg = &core.OutputData{}
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
		err = storageClient.WriteProtobuf(ctx, uri, storage.Options{}, msg)
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
