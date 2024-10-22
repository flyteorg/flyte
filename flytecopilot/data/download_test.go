package data

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"

	"github.com/stretchr/testify/assert"
)

func TestHandleBlobMultipart(t *testing.T) {
	t.Run("Successful Query", func(t *testing.T) {
		s, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)
		ref := storage.DataReference("s3://container/folder/file1")
		s.WriteRaw(context.Background(), ref, 0, storage.Options{}, bytes.NewReader([]byte{}))
		ref = storage.DataReference("s3://container/folder/file2")
		s.WriteRaw(context.Background(), ref, 0, storage.Options{}, bytes.NewReader([]byte{}))

		d := Downloader{store: s}

		blob := &core.Blob{
			Uri: "s3://container/folder",
			Metadata: &core.BlobMetadata{
				Type: &core.BlobType{
					Dimensionality: core.BlobType_MULTIPART,
				},
			},
		}

		toPath := "./inputs"
		defer func() {
			err := os.RemoveAll(toPath)
			if err != nil {
				t.Errorf("Failed to delete directory: %v", err)
			}
		}()

		result, err := d.handleBlob(context.Background(), blob, toPath)
		assert.NoError(t, err)
		assert.Equal(t, toPath, result)

		// Check if files were created and data written
		for _, file := range []string{"file1", "file2"} {
			if _, err := os.Stat(filepath.Join(toPath, "folder", file)); os.IsNotExist(err) {
				t.Errorf("expected file %s to exist", file)
			}
		}
	})

	t.Run("No Items", func(t *testing.T) {
		s, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)

		d := Downloader{store: s}

		blob := &core.Blob{
			Uri: "s3://container/folder",
			Metadata: &core.BlobMetadata{
				Type: &core.BlobType{
					Dimensionality: core.BlobType_MULTIPART,
				},
			},
		}

		toPath := "./inputs"
		defer func() {
			err := os.RemoveAll(toPath)
			if err != nil {
				t.Errorf("Failed to delete directory: %v", err)
			}
		}()

		result, err := d.handleBlob(context.Background(), blob, toPath)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestHandleBlobSinglePart(t *testing.T) {
	s, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)
	ref := storage.DataReference("s3://container/file")
	s.WriteRaw(context.Background(), ref, 0, storage.Options{}, bytes.NewReader([]byte{}))

	d := Downloader{store: s}

	blob := &core.Blob{
		Uri: "s3://container/file",
		Metadata: &core.BlobMetadata{
			Type: &core.BlobType{
				Dimensionality: core.BlobType_SINGLE,
			},
		},
	}

	toPath := "./input"
	defer func() {
		err := os.RemoveAll(toPath)
		if err != nil {
			t.Errorf("Failed to delete file: %v", err)
		}
	}()

	result, err := d.handleBlob(context.Background(), blob, toPath)
	assert.NoError(t, err)
	assert.Equal(t, toPath, result)

	// Check if files were created and data written
	if _, err := os.Stat(toPath); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist", toPath)
	}
}

func TestHandleBlobHTTP(t *testing.T) {
	s, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)
	d := Downloader{store: s}

	blob := &core.Blob{
		Uri: "https://raw.githubusercontent.com/flyteorg/flyte/master/README.md",
		Metadata: &core.BlobMetadata{
			Type: &core.BlobType{
				Dimensionality: core.BlobType_SINGLE,
			},
		},
	}

	toPath := "./input"
	defer func() {
		err := os.RemoveAll(toPath)
		if err != nil {
			t.Errorf("Failed to delete file: %v", err)
		}
	}()

	result, err := d.handleBlob(context.Background(), blob, toPath)
	assert.NoError(t, err)
	assert.Equal(t, toPath, result)

	// Check if files were created and data written
	if _, err := os.Stat(toPath); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist", toPath)
	}
}
