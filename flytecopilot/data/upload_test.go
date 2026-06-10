package data

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func TestUploader_RecursiveUpload(t *testing.T) {

	tmpFolderLocation := ""
	tmpPrefix := "upload_test"

	t.Run("upload-blob", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp(tmpFolderLocation, tmpPrefix)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, os.RemoveAll(tmpDir))
		}()

		vmap := &core.VariableMap{
			Variables: []*core.VariableEntry{
				{
					Key: "x",
					Value: &core.Variable{
						Type: &core.LiteralType{Type: &core.LiteralType_Blob{Blob: &core.BlobType{Dimensionality: core.BlobType_SINGLE}}},
					},
				},
			},
		}

		data := []byte("data")
		assert.NoError(t, os.WriteFile(path.Join(tmpDir, "x"), data, os.ModePerm)) // #nosec G306
		fmt.Printf("Written to %s ", path.Join(tmpDir, "x"))

		store, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)

		outputRef := storage.DataReference("output")
		rawRef := storage.DataReference("raw")
		u := NewUploader(context.TODO(), store, core.DataLoadingConfig_JSON, core.IOStrategy_UPLOAD_ON_EXIT, "error")
		assert.NoError(t, u.RecursiveUpload(context.TODO(), vmap, tmpDir, outputRef, rawRef))

		outputs := &core.LiteralMap{}
		assert.NoError(t, store.ReadProtobuf(context.TODO(), outputRef, outputs))
		assert.Len(t, outputs.GetLiterals(), 1)
		assert.NotNil(t, outputs.GetLiterals()["x"])
		assert.NotNil(t, outputs.GetLiterals()["x"].GetScalar())
		assert.NotNil(t, outputs.GetLiterals()["x"].GetScalar().GetBlob())
		ref := storage.DataReference(outputs.GetLiterals()["x"].GetScalar().GetBlob().GetUri())
		r, err := store.ReadRaw(context.TODO(), ref)
		assert.NoError(t, err, "%s does not exist", ref)
		defer r.Close()
		b, err := io.ReadAll(r)
		assert.NoError(t, err)
		assert.Equal(t, string(data), string(b), "content dont match")
	})

	t.Run("upload-multipart-nested", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp(tmpFolderLocation, tmpPrefix)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, os.RemoveAll(tmpDir))
		}()

		// A multipart blob output "x" whose directory contains a root file
		// and files nested in subdirectories (including a basename collision
		// across two different subdirs).
		varDir := path.Join(tmpDir, "x")
		assert.NoError(t, os.MkdirAll(path.Join(varDir, "nested"), os.ModePerm)) // #nosec G301
		assert.NoError(t, os.MkdirAll(path.Join(varDir, "other"), os.ModePerm))  // #nosec G301
		contents := map[string]string{
			"root.txt":        "root",
			"nested/deep.txt": "deep",
			"nested/dup.txt":  "dup-in-nested",
			"other/dup.txt":   "dup-in-other",
		}
		for rel, body := range contents {
			assert.NoError(t, os.WriteFile(path.Join(varDir, rel), []byte(body), os.ModePerm)) // #nosec G306
		}

		vmap := &core.VariableMap{
			Variables: []*core.VariableEntry{
				{
					Key: "x",
					Value: &core.Variable{
						Type: &core.LiteralType{Type: &core.LiteralType_Blob{Blob: &core.BlobType{Dimensionality: core.BlobType_MULTIPART}}},
					},
				},
			},
		}

		store, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)

		outputRef := storage.DataReference("output")
		rawRef := storage.DataReference("raw")
		u := NewUploader(context.TODO(), store, core.DataLoadingConfig_JSON, core.IOStrategy_UPLOAD_ON_EXIT, "error")
		assert.NoError(t, u.RecursiveUpload(context.TODO(), vmap, tmpDir, outputRef, rawRef))

		outputs := &core.LiteralMap{}
		assert.NoError(t, store.ReadProtobuf(context.TODO(), outputRef, outputs))
		assert.Len(t, outputs.GetLiterals(), 1)
		blob := outputs.GetLiterals()["x"].GetScalar().GetBlob()
		assert.NotNil(t, blob)
		assert.Equal(t, core.BlobType_MULTIPART, blob.GetMetadata().GetType().GetDimensionality())

		// Every file must land at a structure-preserving key under the blob
		// prefix; the two dup.txt files must not collide.
		base := storage.DataReference(blob.GetUri())
		for rel, body := range contents {
			ref, err := store.ConstructReference(context.TODO(), base, strings.Split(rel, "/")...)
			assert.NoError(t, err)
			r, err := store.ReadRaw(context.TODO(), ref)
			assert.NoError(t, err, "%s does not exist", ref)
			b, err := io.ReadAll(r)
			assert.NoError(t, err)
			assert.NoError(t, r.Close())
			assert.Equal(t, body, string(b), "content mismatch for %s", rel)
		}
	})
}
