package utils

import (
	"context"
	"fmt"
	"io"

	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type FailingRawStore struct {
}

func (FailingRawStore) CopyRaw(context.Context, storage.DataReference, storage.DataReference, storage.Options) error {
	return fmt.Errorf("failed to copy raw")
}

func (FailingRawStore) CreateSignedURL(context.Context, storage.DataReference, storage.SignedURLProperties) (storage.SignedURLResponse, error) {
	return storage.SignedURLResponse{}, fmt.Errorf("failed to create signed url")
}

func (FailingRawStore) GetBaseContainerFQN(context.Context) storage.DataReference {
	return ""
}

func (FailingRawStore) Head(context.Context, storage.DataReference) (storage.Metadata, error) {
	return nil, fmt.Errorf("failed metadata fetch")
}

func (FailingRawStore) List(context.Context, storage.DataReference, int, storage.Cursor) ([]storage.DataReference, storage.Cursor, error) {
	return nil, storage.NewCursorAtEnd(), fmt.Errorf("Not implemented yet")
}

func (FailingRawStore) ReadRaw(context.Context, storage.DataReference) (io.ReadCloser, error) {
	return nil, fmt.Errorf("failed read raw")
}

func (FailingRawStore) WriteRaw(context.Context, storage.DataReference, int64, storage.Options, io.Reader) error {
	return fmt.Errorf("failed write raw")
}

func (FailingRawStore) Delete(context.Context, storage.DataReference) error {
	return fmt.Errorf("failed to delete")
}
