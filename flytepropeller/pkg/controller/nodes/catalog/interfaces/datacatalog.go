package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	futureFileReaderInterfaces "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/interfaces"
)

//go:generate mockery --name CatalogClient --output ../mocks --case=snake --with-expecter

// FutureClient extends Client interface to support asynchronous operations
type CatalogClient interface {
	catalog.Client
	// GetFuture asynchronously retrieves an artifact for the given key
	GetFuture(ctx context.Context, key catalog.Key) (catalog.Entry, error)
	// PutFuture asynchronously stores new data using the specified key
	PutFuture(ctx context.Context, key catalog.Key, futureReader futureFileReaderInterfaces.FutureFileReaderInterface, metadata catalog.Metadata) (catalog.Status, error)
	// UpdateFuture asynchronously updates existing data at the specified key
	UpdateFuture(ctx context.Context, key catalog.Key, futureReader futureFileReaderInterfaces.FutureFileReaderInterface, metadata catalog.Metadata) (catalog.Status, error)
}
