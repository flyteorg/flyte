package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
)

// volumeLocator is the part of a Volume's published metadata object this
// service needs. The object carries the whole serialized Volume; everything
// not listed here is deliberately ignored, so the two sides can evolve
// independently.
type volumeLocator struct {
	Name string `json:"name"`
	// Where this volume's chunks live, relative to the bucket root.
	//
	// It cannot be derived from Name, and that is the single most expensive
	// thing to get wrong here. The volume client stores chunks under the
	// metadata's *format* name, and a fork copies the format forward while
	// taking a new Volume name -- so a fork's chunks stay under its ancestor's
	// name. A policy built from Name would authenticate perfectly and then be
	// denied on the first chunk read, which looks like a platform fault rather
	// than a scoping mistake. So the client records it, and this refuses to
	// guess when it is absent.
	ChunkPrefix string `json:"chunk_prefix"`
	// Where version metadata objects are written, as a URI prefix. Note it is
	// the PREFIX, not this version's own directory: sealing a new version
	// writes a sibling under it, so a session scoped to the version it mounted
	// could read its index and then be refused the upload that commits it.
	MetadataPrefix string `json:"metadata_prefix"`
	Bucket         string `json:"bucket"`
}

type resolvedVolume struct {
	name     string
	prefixes []string
}

// maxLocatorBytes caps the metadata object read. A locator is a small JSON
// document; anything larger is not one, and reading it into memory on the word
// of the caller who named it is how a credential endpoint becomes a memory
// exhaustion endpoint.
const maxLocatorBytes = 1 << 20

func (s *Service) readVolumeLocator(ctx context.Context, locator string, bucket string) (resolvedVolume, error) {
	if !sameBucket(locator, bucket) {
		// Refused before the read, not after: fetching an object because a
		// caller named it is the thing to avoid, whatever the object says.
		return resolvedVolume{}, connect.NewError(connect.CodePermissionDenied, fmt.Errorf(
			"the locator is not in this data plane's store, so credentials for it cannot be vended here"))
	}

	rc, err := s.dataStore.ReadRaw(ctx, storage.DataReference(locator))
	if err != nil {
		logger.Warnf(ctx, "could not read volume locator %q: %v", locator, err)
		return resolvedVolume{}, connect.NewError(connect.CodeNotFound, fmt.Errorf(
			"no volume metadata could be read at the given locator"))
	}
	defer func() {
		if cerr := rc.Close(); cerr != nil {
			logger.Debugf(ctx, "closing volume locator read: %v", cerr)
		}
	}()

	raw, err := io.ReadAll(io.LimitReader(rc, maxLocatorBytes))
	if err != nil {
		return resolvedVolume{}, connect.NewError(connect.CodeInternal, fmt.Errorf(
			"could not read the volume metadata object"))
	}

	var vol volumeLocator
	if err := json.Unmarshal(raw, &vol); err != nil {
		return resolvedVolume{}, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf(
			"the object at that locator is not volume metadata"))
	}

	if vol.ChunkPrefix == "" {
		return resolvedVolume{}, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf(
			"this volume's metadata does not record where its chunks live (chunk_prefix), which a "+
				"scoped credential has to know and cannot infer -- a fork's chunks stay under its "+
				"ancestor's name. It was written by a client older than that field; mount it with "+
				"ambient credentials, or re-commit it with a current client"))
	}

	prefixes := []string{strings.Trim(vol.ChunkPrefix, "/")}
	if metaPrefix := bucketSubpath(vol.MetadataPrefix); metaPrefix != "" {
		prefixes = append(prefixes, metaPrefix)
	}
	return resolvedVolume{name: vol.Name, prefixes: prefixes}, nil
}

// sameBucket reports whether a URI points into the given bucket (which is a
// bare bucket/container name).
func sameBucket(uri, bucket string) bool {
	if bucket == "" {
		return false
	}
	return bucketOf(uri) == bucket
}

// bucketOf returns the bucket (or container) name from a storage URI.
func bucketOf(uri string) string {
	rest := uri
	if i := strings.Index(rest, "://"); i >= 0 {
		rest = rest[i+3:]
	}
	// Azure's abfss://<container>@<account>.../prefix: the container is what
	// plays the bucket's part.
	if at := strings.Index(rest, "@"); at >= 0 && at < strings.IndexByte(rest+"/", '/') {
		return rest[:at]
	}
	if slash := strings.IndexByte(rest, '/'); slash >= 0 {
		return rest[:slash]
	}
	return rest
}

// bucketSubpath returns the path inside the bucket, with no leading or
// trailing slash.
func bucketSubpath(uri string) string {
	rest := uri
	if i := strings.Index(rest, "://"); i >= 0 {
		rest = rest[i+3:]
	}
	if slash := strings.IndexByte(rest, '/'); slash >= 0 {
		return strings.Trim(rest[slash+1:], "/")
	}
	return ""
}
