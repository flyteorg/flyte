// Package storage defines extensible storage interface.
// This package registers "storage" config section that maps to Config struct. Use NewDataStore(cfg) to initialize a
// DataStore with the provided config. The package provides default implementation to access local, S3 (and minio),
// and In-Memory storage. Use NewCompositeDataStore to swap any portions of the DataStore interface with an external
// implementation (e.g. a cached protobuf store). The underlying storage is provided by extensible "stow" library. You
// can use NewStowRawStore(cfg) to create a Raw store based on any other stow-supported configs (e.g. Azure Blob Storage)
package storage

import (
	"context"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/stow"
)

// DataReference defines a reference to data location.
type DataReference string

var emptyStore = DataStore{}

// Options holds storage options. It is used to pass Metadata (like headers for S3) and also tags or labels for
// objects
type Options struct {
	Metadata map[string]interface{}
}

// Metadata is a placeholder for data reference metadata.
type Metadata interface {
	Exists() bool
	Size() int64
	Etag() string
	// ContentMD5 retrieves the value of a special metadata tag added by the system that
	// contains the MD5 of the uploaded file. If there is no metadata attached
	// or that `FlyteContentMD5` key isn't set, ContentMD5 will return empty.
	ContentMD5() string
}

type CursorState int

const (
	// Enum representing state of the cursor
	AtStartCursorState     CursorState = 0
	AtEndCursorState       CursorState = 1
	AtCustomPosCursorState CursorState = 2
)

type Cursor struct {
	cursorState    CursorState
	customPosition string
}

func NewCursorAtStart() Cursor {
	return Cursor{
		cursorState:    AtStartCursorState,
		customPosition: "",
	}
}

func NewCursorAtEnd() Cursor {
	return Cursor{
		cursorState:    AtEndCursorState,
		customPosition: "",
	}
}

func NewCursorFromCustomPosition(customPosition string) Cursor {
	return Cursor{
		cursorState:    AtCustomPosCursorState,
		customPosition: customPosition,
	}
}

// DataStore is a simplified interface for accessing and storing data in one of the Cloud stores.
// Today we rely on Stow for multi-cloud support, but this interface abstracts that part
type DataStore struct {
	ComposedProtobufStore
	ReferenceConstructor
	metrics *dataStoreMetrics
}

// SignedURLProperties encapsulates properties about the signedURL operation.
type SignedURLProperties struct {
	// Scope defines the permission level allowed for the generated URL.
	Scope stow.ClientMethod
	// ExpiresIn defines the expiration duration for the URL. It's strongly recommended setting it.
	ExpiresIn time.Duration
	// ContentMD5 defines the expected hash of the generated file. It's strongly recommended setting it.
	ContentMD5 string
	// AddContentMD5Metadata Add ContentMD5 to the metadata of signed URL if true.
	AddContentMD5Metadata bool
}

type SignedURLResponse struct {
	URL                    url.URL
	RequiredRequestHeaders map[string]string
}

//go:generate mockery -name RawStore -case=underscore

// RawStore defines a low level interface for accessing and storing bytes.
type RawStore interface {
	// GetBaseContainerFQN returns a FQN DataReference with the configured base init container
	GetBaseContainerFQN(ctx context.Context) DataReference

	// CreateSignedURL creates a signed url with the provided properties.
	CreateSignedURL(ctx context.Context, reference DataReference, properties SignedURLProperties) (SignedURLResponse, error)

	// Head gets metadata about the reference. This should generally be a light weight operation.
	Head(ctx context.Context, reference DataReference) (Metadata, error)

	// List gets a list of items given a prefix, using a paginated API
	List(ctx context.Context, reference DataReference, maxItems int, cursor Cursor) ([]DataReference, Cursor, error)

	// ReadRaw retrieves a byte array from the Blob store or an error
	ReadRaw(ctx context.Context, reference DataReference) (io.ReadCloser, error)

	// WriteRaw stores a raw byte array.
	WriteRaw(ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader) error

	// CopyRaw copies from source to destination.
	CopyRaw(ctx context.Context, source, destination DataReference, opts Options) error

	// Delete removes the referenced data from the blob store.
	Delete(ctx context.Context, reference DataReference) error
}

//go:generate mockery -name ReferenceConstructor -case=underscore

// ReferenceConstructor defines an interface for building data reference paths.
type ReferenceConstructor interface {
	// ConstructReference creates a new dataReference that matches the storage structure.
	ConstructReference(ctx context.Context, reference DataReference, nestedKeys ...string) (DataReference, error)

	// FromSignedURL constructs a data reference from a signed URL
	//FromSignedURL(ctx context.Context, signedURL string) (DataReference, error)
}

// ProtobufStore defines an interface for reading and writing protobuf messages
type ProtobufStore interface {
	// ReadProtobuf retrieves the entire blob from blobstore and unmarshals it to the passed protobuf
	ReadProtobuf(ctx context.Context, reference DataReference, msg proto.Message) error

	// WriteProtobuf serializes and stores the protobuf.
	WriteProtobuf(ctx context.Context, reference DataReference, opts Options, msg proto.Message) error
}

//go:generate mockery -name ComposedProtobufStore -case=underscore

// ComposedProtobufStore interface includes all the necessary data to allow a ProtobufStore to interact with storage
// through a RawStore.
type ComposedProtobufStore interface {
	RawStore
	ProtobufStore
}

// Split splits the data reference into parts.
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
