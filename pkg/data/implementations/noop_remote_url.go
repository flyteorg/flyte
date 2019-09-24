package implementations

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/data/interfaces"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flytestdlib/storage"
	"google.golang.org/grpc/codes"
)

// No-op implementation of a RemoteURLInterface
type NoopRemoteURL struct {
	remoteDataStoreClient storage.DataStore
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

func NewNoopRemoteURL(remoteDataStoreClient storage.DataStore) interfaces.RemoteURLInterface {
	return &NoopRemoteURL{
		remoteDataStoreClient: remoteDataStoreClient,
	}
}
