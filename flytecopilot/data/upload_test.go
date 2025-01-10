package data

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
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
			Variables: map[string]*core.Variable{
				"x": {
					Type: &core.LiteralType{Type: &core.LiteralType_Blob{Blob: &core.BlobType{Dimensionality: core.BlobType_SINGLE}}},
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
}
