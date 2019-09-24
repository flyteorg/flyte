package interfaces

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

// Defines an interface for fetching pre-signed URLs.
type RemoteURLInterface interface {
	Get(ctx context.Context, uri string) (admin.UrlBlob, error)
}
