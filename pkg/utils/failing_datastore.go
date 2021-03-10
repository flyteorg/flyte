package utils

import (
	"context"
	"fmt"
	"io"

	"github.com/flyteorg/flytestdlib/storage"
)

type FailingRawStore struct {
}

func (FailingRawStore) CopyRaw(ctx context.Context, source, destination storage.DataReference, opts storage.Options) error {
	return fmt.Errorf("failed to copy raw")
}

func (FailingRawStore) GetBaseContainerFQN(ctx context.Context) storage.DataReference {
	return ""
}

func (FailingRawStore) Head(ctx context.Context, reference storage.DataReference) (storage.Metadata, error) {
	return nil, fmt.Errorf("failed metadata fetch")
}

func (FailingRawStore) ReadRaw(ctx context.Context, reference storage.DataReference) (io.ReadCloser, error) {
	return nil, fmt.Errorf("failed read raw")
}

func (FailingRawStore) WriteRaw(ctx context.Context, reference storage.DataReference, size int64, opts storage.Options, raw io.Reader) error {
	return fmt.Errorf("failed write raw")
}
