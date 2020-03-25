package ioutils

import (
	"context"
	"crypto/sha1" // #nosec
	"encoding/hex"

	"github.com/lyft/flytestdlib/storage"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io"
)

type precomputedRawOutputPaths struct {
	path storage.DataReference
}

func (r precomputedRawOutputPaths) GetRawOutputPrefix() storage.DataReference {
	return r.path
}

// Creates a deterministic RawOutputPath whose path is distributed based on the ShardSelector passed in.
// Determinism depends on the outputMetadataPath
// Potential performance problem, as creating a new RawPath creation may be expensive as it hashes the outputMetadataPath
// the final RawOutputPath is created in the shard selected by the sharder at the basePath and then appended by a hashed value of the outputMetadata
func NewShardedDeterministicRawOutputPath(ctx context.Context, sharder ShardSelector, basePrefix, outputMetadataPrefix storage.DataReference, store storage.ReferenceConstructor) (io.RawOutputPaths, error) {
	o := []byte(outputMetadataPrefix)
	prefix, err := sharder.GetShardPrefix(ctx, o)
	if err != nil {
		return nil, err
	}
	/* #nosec */
	// We use SHA1 for sheer speed instead of no collisions. As because of the shard Prefix + hash is pretty unique :)
	m := sha1.New()
	if _, err := m.Write(o); err != nil {
		return nil, err
	}
	path, err := store.ConstructReference(ctx, basePrefix, prefix, hex.EncodeToString(m.Sum(nil)))
	if err != nil {
		return nil, err
	}
	return precomputedRawOutputPaths{
		path: path,
	}, nil
}

// A simple Output sandbox at a given path
func NewRawOutputPaths(_ context.Context, outputSandboxPath storage.DataReference) io.RawOutputPaths {
	return precomputedRawOutputPaths{path: outputSandboxPath}
}

// Creates an OutputSandbox in the basePath using the uniqueID and a sharder
// This implementation is faster than the Randomized strategy
func NewShardedRawOutputPath(ctx context.Context, sharder ShardSelector, basePath storage.DataReference, uniqueID string, store storage.ReferenceConstructor) (io.RawOutputPaths, error) {
	o := []byte(uniqueID)
	prefix, err := sharder.GetShardPrefix(ctx, o)
	if err != nil {
		return nil, err
	}
	path, err := store.ConstructReference(ctx, basePath, prefix, uniqueID)
	if err != nil {
		return nil, err
	}
	return precomputedRawOutputPaths{
		path: path,
	}, nil
}
