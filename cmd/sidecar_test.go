package cmd

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flytecopilot/cmd/containerwatcher"
)

func TestUploadOptions_Upload(t *testing.T) {
	tmpFolderLocation := ""
	tmpPrefix := "upload_test"
	outputPath := "output"

	ctx := context.TODO()

	t.Run("uploadNoOutputs", func(t *testing.T) {
		tmpDir, err := ioutil.TempDir(tmpFolderLocation, tmpPrefix)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, os.RemoveAll(tmpDir))
		}()

		s := promutils.NewTestScope()
		store, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, s.NewSubScope("storage"))
		assert.NoError(t, err)
		uopts := UploadOptions{
			RootOptions: &RootOptions{
				Scope: s,
				Store: store,
			},
			remoteOutputsPrefix: outputPath,
			metadataFormat:      core.DataLoadingConfig_JSON.String(),
			uploadMode:          core.IOStrategy_UPLOAD_ON_EXIT.String(),
			startWatcherType:    containerwatcher.WatcherTypeFile,
			localDirectoryPath:  tmpDir,
		}

		assert.NoError(t, uopts.Sidecar(ctx))
	})

	t.Run("uploadBlobType-FileNotFound", func(t *testing.T) {
		tmpDir, err := ioutil.TempDir(tmpFolderLocation, tmpPrefix)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, os.RemoveAll(tmpDir))
		}()
		s := promutils.NewTestScope()
		store, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, s.NewSubScope("storage"))
		assert.NoError(t, err)

		iface := &core.TypedInterface{
			Outputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"x": {
						Type:        &core.LiteralType{Type: &core.LiteralType_Blob{Blob: &core.BlobType{Dimensionality: core.BlobType_SINGLE}}},
						Description: "example",
					},
				},
			},
		}
		d, err := proto.Marshal(iface)
		assert.NoError(t, err)

		uopts := UploadOptions{
			RootOptions: &RootOptions{
				Scope:           s,
				Store:           store,
				errorOutputName: "errors.pb",
			},
			remoteOutputsPrefix: outputPath,
			metadataFormat:      core.DataLoadingConfig_JSON.String(),
			uploadMode:          core.IOStrategy_UPLOAD_ON_EXIT.String(),
			startWatcherType:    containerwatcher.WatcherTypeNoop,
			exitWatcherType:     containerwatcher.WatcherTypeFile,
			typedInterface:      d,
			localDirectoryPath:  tmpDir,
		}

		success := path.Join(tmpDir, SuccessFile)
		assert.NoError(t, ioutil.WriteFile(success, []byte("done"), os.ModePerm))
		ok, err := containerwatcher.FileExists(success)
		assert.NoError(t, err)
		assert.True(t, ok, "sucessfile not created")
		assert.NoError(t, uopts.Sidecar(ctx))
		v, err := store.Head(ctx, "/output/errors.pb")
		assert.NoError(t, err)
		assert.True(t, v.Exists())
	})
}
