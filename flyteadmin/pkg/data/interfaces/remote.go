package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

// Defines an interface for fetching pre-signed URLs.
type RemoteURLInterface interface {
	// TODO: Refactor for URI to be of type DataReference. We should package a FromString-like function in flytestdlib
	Get(ctx context.Context, uri string) (admin.UrlBlob, error)
}
