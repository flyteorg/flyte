package cmd

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/flytecopilot/cmd/containerwatcher"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
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
			startWatcherType:    containerwatcher.WatcherTypeNoop,
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
				Variables: []*core.VariableEntry{
					{
						Key: "x",
						Value: &core.Variable{
							Type:        &core.LiteralType{Type: &core.LiteralType_Blob{Blob: &core.BlobType{Dimensionality: core.BlobType_SINGLE}}},
							Description: "example",
						},
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
				errorOutputName: ioutils.ErrorsSuffix,
			},
			remoteOutputsPrefix: outputPath,
			metadataFormat:      core.DataLoadingConfig_JSON.String(),
			uploadMode:          core.IOStrategy_UPLOAD_ON_EXIT.String(),
			startWatcherType:    containerwatcher.WatcherTypeNoop,
			exitWatcherType:     containerwatcher.WatcherTypeNoop,
			typedInterface:      d,
			localDirectoryPath:  tmpDir,
		}

		assert.NoError(t, uopts.Sidecar(ctx))
		v, err := store.Head(ctx, storage.DataReference("/output/"+ioutils.ErrorsSuffix))
		assert.NoError(t, err)
		assert.True(t, v.Exists())
	})
}

// The failure a container declares is the one flyte must act on: writing a
// NON_RECOVERABLE user error to the error file is the only way a raw container
// can say "do not retry me", and until the document reaches the output reader
// intact the container is retried against the cluster's system-failure budget.
func TestSidecarForwardsTheContainersErrorDocument(t *testing.T) {
	ctx := context.TODO()
	tmpDir, err := os.MkdirTemp("", "upload_test")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, os.RemoveAll(tmpDir))
	}()

	document, err := proto.Marshal(&core.ErrorDocument{
		Error: &core.ContainerError{
			Code:    "BenchmarkFailed",
			Message: "the model rejected the prompt",
			Kind:    core.ContainerError_NON_RECOVERABLE,
			Origin:  core.ExecutionError_USER,
		},
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path.Join(tmpDir, ErrorFile), document, os.ModePerm)) // #nosec G306

	iface, err := proto.Marshal(&core.TypedInterface{
		Outputs: &core.VariableMap{
			Variables: []*core.VariableEntry{{
				Key: "x",
				Value: &core.Variable{
					Type: &core.LiteralType{Type: &core.LiteralType_Blob{Blob: &core.BlobType{Dimensionality: core.BlobType_SINGLE}}},
				},
			}},
		},
	})
	require.NoError(t, err)

	s := promutils.NewTestScope()
	store, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, s.NewSubScope("storage"))
	require.NoError(t, err)
	prefix := storage.DataReference("output")
	uopts := UploadOptions{
		RootOptions: &RootOptions{
			Scope:           s,
			Store:           store,
			errorOutputName: ioutils.ErrorsSuffix,
		},
		remoteOutputsPrefix: string(prefix),
		metadataFormat:      core.DataLoadingConfig_JSON.String(),
		uploadMode:          core.IOStrategy_UPLOAD_ON_EXIT.String(),
		startWatcherType:    containerwatcher.WatcherTypeNoop,
		exitWatcherType:     containerwatcher.WatcherTypeNoop,
		typedInterface:      iface,
		localDirectoryPath:  tmpDir,
	}

	require.NoError(t, uopts.Sidecar(ctx))

	reader := ioutils.NewRemoteFileOutputReader(ctx, store, ioutils.NewReadOnlyOutputFilePaths(ctx, store, prefix), 1024*1024)
	isErr, err := reader.IsError(ctx)
	require.NoError(t, err)
	require.True(t, isErr, "reader should see the error document the container wrote")

	ee, err := reader.ReadError(ctx)
	require.NoError(t, err)
	assert.False(t, ee.IsRecoverable, "the container declared a permanent failure")
	assert.Equal(t, core.ExecutionError_USER, ee.ExecutionError.GetKind())
	assert.Equal(t, "BenchmarkFailed", ee.ExecutionError.GetCode())
	assert.Equal(t, "the model rejected the prompt", ee.ExecutionError.GetMessage())
}
