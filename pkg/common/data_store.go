package common

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/storage"
)

func OffloadLiteralMap(ctx context.Context, storageClient *storage.DataStore, literalMap *core.LiteralMap, nestedKeys ...string) (storage.DataReference, error) {
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
	if err := storageClient.WriteProtobuf(ctx, uri, storage.Options{}, literalMap); err != nil {
		return "", errors.NewFlyteAdminErrorf(codes.Internal, "Failed to write protobuf for [%+v] with err: %v", nestedKeys, err)
	}
	return uri, nil
}
