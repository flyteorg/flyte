package data

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func TestHandleBlobMultipartReportsFailedCopies(t *testing.T) {
	store, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	ref := storage.DataReference("s3://container/prefix/file.txt")
	payload := []byte("payload")
	err = store.WriteRaw(context.Background(), ref, int64(len(payload)), storage.Options{}, bytes.NewReader(payload))
	assert.NoError(t, err)

	downloader := Downloader{store: store}
	blob := &core.Blob{
		Uri: "s3://container/prefix",
		Metadata: &core.BlobMetadata{
			Type: &core.BlobType{Dimensionality: core.BlobType_MULTIPART},
		},
	}

	blockedPath := filepath.Join(t.TempDir(), "blocked")
	err = os.WriteFile(blockedPath, []byte("x"), 0o600)
	assert.NoError(t, err)

	result, err := downloader.handleBlob(context.Background(), blob, blockedPath)
	assert.Nil(t, result)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "Failed to copy 1 out of 1 remote files")
		assert.Contains(t, err.Error(), "Failed to close 1 writers")
	}
}
