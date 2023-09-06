package ioutils

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
)

type cachedInputReader struct {
	io.InputReader
	cachedInputs *core.LiteralMap
}

func (c *cachedInputReader) Get(ctx context.Context) (*core.LiteralMap, error) {
	if c.cachedInputs == nil {
		in, err := c.InputReader.Get(ctx)
		if err != nil {
			return nil, err
		}
		c.cachedInputs = in
	}
	return c.cachedInputs, nil
}

// Creates a new Read-through cached Input Reader. the returned reader is not thread-safe
// It caches the inputs on a successful read from the underlying input reader
func NewCachedInputReader(ctx context.Context, in io.InputReader) io.InputReader {
	return &cachedInputReader{
		InputReader: in,
	}
}
