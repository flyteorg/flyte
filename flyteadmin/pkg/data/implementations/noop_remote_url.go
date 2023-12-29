package implementations

import (
	"context"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"

	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/data/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

// No-op implementation of a RemoteURLInterface
type NoopRemoteURL struct {
	remoteDataStoreClient common.DatastoreClient
}

func (n *NoopRemoteURL) Get(ctx context.Context, uri string) (admin.UrlBlob, error) {
	metadata, err := n.remoteDataStoreClient.Head(ctx, storage.DataReference(uri))
	if err != nil {
		return admin.UrlBlob{}, errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to get metadata for uri: %s with err: %v", uri, err)
	}
	return admin.UrlBlob{
		Url:   uri,
		Bytes: metadata.Size(),
	}, nil
}

func NewNoopRemoteURL(remoteDataStoreClient common.DatastoreClient) interfaces.RemoteURLInterface {
	return &NoopRemoteURL{
		remoteDataStoreClient: remoteDataStoreClient,
	}
}
