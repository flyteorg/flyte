// Defines extensible storage interface.
// This package registers "storage" config section that maps to Config struct. Use NewDataStore(cfg) to initialize a
// DataStore with the provided config. The package provides default implementation to access local, S3 (and minio),
// and In-Memory storage. Use NewCompositeDataStore to swap any portions of the DataStore interface with an external
// implementation (e.g. a cached protobuf store). The underlying storage is provided by extensible "stow" library. You
// can use NewStowRawStore(cfg) to create a Raw store based on any other stow-supported configs (e.g. Azure Blob Storage)
package storage

import (
	"context"
	"strings"

	"io"
	"net/url"

	"github.com/golang/protobuf/proto"
)

// Defines a reference to data location.
type DataReference string

var emptyStore = DataStore{}

// Holder for recording storage options. It is used to pass Metadata (like headers for S3) and also tags or labels for
// objects
type Options struct {
	Metadata map[string]interface{}
}

// Placeholder for data reference metadata.
type Metadata interface {
	Exists() bool
	Size() int64
}

// A simplified interface for accessing and storing data in one of the Cloud stores.
// Today we rely on Stow for multi-cloud support, but this interface abstracts that part
type DataStore struct {
	ComposedProtobufStore
	ReferenceConstructor
}

// Defines a low level interface for accessing and storing bytes.
type RawStore interface {
	// returns a FQN DataReference with the configured base init container
	GetBaseContainerFQN(ctx context.Context) DataReference

	// Gets metadata about the reference. This should generally be a light weight operation.
	Head(ctx context.Context, reference DataReference) (Metadata, error)

	// Retrieves a byte array from the Blob store or an error
	ReadRaw(ctx context.Context, reference DataReference) (io.ReadCloser, error)

	// Stores a raw byte array.
	WriteRaw(ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader) error

	// Copies from source to destination.
	CopyRaw(ctx context.Context, source, destination DataReference, opts Options) error
}

// Defines an interface for building data reference paths.
type ReferenceConstructor interface {
	// Creates a new dataReference that matches the storage structure.
	ConstructReference(ctx context.Context, reference DataReference, nestedKeys ...string) (DataReference, error)
}

// Defines an interface for reading and writing protobuf messages
type ProtobufStore interface {
	// Retrieves the entire blob from blobstore and unmarshals it to the passed protobuf
	ReadProtobuf(ctx context.Context, reference DataReference, msg proto.Message) error

	// Serializes and stores the protobuf.
	WriteProtobuf(ctx context.Context, reference DataReference, opts Options, msg proto.Message) error
}

// A ProtobufStore needs a RawStore to get the RawData. This interface provides all the necessary components to make
// Protobuf fetching work
type ComposedProtobufStore interface {
	RawStore
	ProtobufStore
}

// Splits the data reference into parts.
func (r DataReference) Split() (scheme, container, key string, err error) {
	u, err := url.Parse(string(r))
	if err != nil {
		return "", "", "", err
	}

	return u.Scheme, u.Host, strings.Trim(u.Path, "/"), nil
}

func (r DataReference) String() string {
	return string(r)
}
