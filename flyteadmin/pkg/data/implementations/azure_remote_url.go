package implementations

import (
	"context"
	"github.com/flyteorg/flyte/flyteadmin/pkg/data/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	"github.com/flyteorg/stow"
	"google.golang.org/grpc/codes"
	"time"
)

type AzureRemoteURL struct {
	remoteDataStoreClient storage.DataStore
	presignDuration       time.Duration
}

func (n *AzureRemoteURL) Get(ctx context.Context, uri string) (*admin.UrlBlob, error) {
	metadata, err := n.remoteDataStoreClient.Head(ctx, storage.DataReference(uri))
	if err != nil {
		return &admin.UrlBlob{}, errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to get metadata for uri: %s with err: %v", uri, err)
	}

	signedUri, err := n.remoteDataStoreClient.CreateSignedURL(ctx, storage.DataReference(uri), storage.SignedURLProperties{
		Scope:     stow.ClientMethodGet,
		ExpiresIn: n.presignDuration,
	})
	if err != nil {
		return &admin.UrlBlob{}, errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to get metadata for uri: %s with err: %v", uri, err)
	}

	return &admin.UrlBlob{
		Url:   signedUri.URL.String(),
		Bytes: metadata.Size(),
	}, nil
}

func NewAzureRemoteURL(remoteDataStoreClient storage.DataStore, presignDuration time.Duration) interfaces.RemoteURLInterface {
	return &AzureRemoteURL{
		remoteDataStoreClient: remoteDataStoreClient,
		presignDuration:       presignDuration,
	}
}
