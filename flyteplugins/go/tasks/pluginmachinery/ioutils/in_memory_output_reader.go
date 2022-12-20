package ioutils

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
)

type InMemoryOutputReader struct {
	literals *core.LiteralMap
	DeckPath *storage.DataReference
	err      *io.ExecutionError
}

var _ io.OutputReader = InMemoryOutputReader{}

func (r InMemoryOutputReader) IsError(ctx context.Context) (bool, error) {
	return r.err != nil, nil
}

func (r InMemoryOutputReader) ReadError(ctx context.Context) (io.ExecutionError, error) {
	if r.err != nil {
		return *r.err, nil
	}

	return io.ExecutionError{}, fmt.Errorf("no execution error specified")
}

func (r InMemoryOutputReader) IsFile(_ context.Context) bool {
	return false
}

func (r InMemoryOutputReader) Exists(_ context.Context) (bool, error) {
	// TODO: should this return true if there is an error?
	return r.literals != nil, nil
}

func (r InMemoryOutputReader) Read(_ context.Context) (*core.LiteralMap, *io.ExecutionError, error) {
	return r.literals, r.err, nil
}

func (r InMemoryOutputReader) DeckExists(_ context.Context) (bool, error) {
	return r.DeckPath != nil, nil
}

func NewInMemoryOutputReader(literals *core.LiteralMap, DeckPath *storage.DataReference, err *io.ExecutionError) InMemoryOutputReader {
	return InMemoryOutputReader{
		literals: literals,
		DeckPath: DeckPath,
		err:      err,
	}
}
