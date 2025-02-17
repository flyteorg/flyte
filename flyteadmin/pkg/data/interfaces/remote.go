package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery-v2 --name=RemoteURLInterface --output=../mocks --case=underscore --with-expecter

// Defines an interface for fetching pre-signed URLs.
type RemoteURLInterface interface {
	// TODO: Refactor for URI to be of type DataReference. We should package a FromString-like function in flytestdlib
	Get(ctx context.Context, uri string) (*admin.UrlBlob, error)
}
