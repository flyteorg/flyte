package ioutils

import "context"

// This interface allows shard selection for OutputSandbox.
type ShardSelector interface {
	GetShardPrefix(ctx context.Context, s []byte) (string, error)
}
