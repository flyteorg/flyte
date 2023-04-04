package catalog

import (
	"context"

	"github.com/flyteorg/flytestdlib/bitarray"

	"github.com/flyteorg/flytestdlib/errors"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
)

type ResponseStatus uint8

const (
	ResponseStatusNotReady ResponseStatus = iota
	ResponseStatusReady
)

const (
	ErrResponseNotReady errors.ErrorCode = "RESPONSE_NOT_READY"
	ErrSystemError      errors.ErrorCode = "SYSTEM_ERROR"
)

type UploadRequest struct {
	Key              Key
	ArtifactData     io.OutputReader
	ArtifactMetadata Metadata
}

type ReadyHandler func(ctx context.Context, future Future)

// A generic Future interface to represent async operations results
type Future interface {
	// Gets the response status for the future. If the future represents multiple operations, the status will only be
	// ready if all of them are.
	GetResponseStatus() ResponseStatus

	// Sets a callback handler to be called when the future status changes to ready.
	OnReady(handler ReadyHandler)

	GetResponseError() error
}

// Catalog Sidecar future to represent async process of uploading catalog artifacts.
type UploadFuture interface {
	Future
}

// Catalog Download Request to represent async operation download request.
type DownloadRequest struct {
	Key    Key
	Target io.OutputWriter
}

// Catalog download future to represent async process of downloading catalog artifacts.
type DownloadFuture interface {
	Future

	// Gets the actual response from the future. This will return an error if the future isn't ready yet.
	GetResponse() (DownloadResponse, error)
}

// Catalog download response.
type DownloadResponse interface {
	// Gets a bit set representing which items from the request were cached.
	GetCachedResults() *bitarray.BitSet

	// Gets the total size of the cached result.
	GetResultsSize() int

	// A convenience method to retrieve the number of cached items.
	GetCachedCount() int
}

// An interface that helps async interaction with catalog service
type AsyncClient interface {
	// Returns if an entry exists for the given task and input. It returns the data as a LiteralMap
	Download(ctx context.Context, requests ...DownloadRequest) (outputFuture DownloadFuture, err error)

	// Adds a new entry to catalog for the given task execution context and the generated output
	Upload(ctx context.Context, requests ...UploadRequest) (putFuture UploadFuture, err error)
}

var _ AsyncClient = AsyncClientImpl{}
