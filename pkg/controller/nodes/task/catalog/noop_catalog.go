package catalog

import (
	"context"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io"
)

type NOOPCatalog struct {
}

func (n NOOPCatalog) Get(ctx context.Context, key catalog.Key) (io.OutputReader, error) {
	return nil, nil
}

func (n NOOPCatalog) Put(ctx context.Context, key catalog.Key, reader io.OutputReader, metadata catalog.Metadata) error {
	return nil
}
