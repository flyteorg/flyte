package cmd

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"
)

func TestDownloadOptions_Download(t *testing.T) {

	tmpFolderLocation := ""
	tmpPrefix := "download_test"
	inputPath := "input/inputs.pb"
	outputPath := "output"

	ctx := context.TODO()
	dopts := DownloadOptions{
		remoteInputsPath:    inputPath,
		remoteOutputsPrefix: outputPath,
		metadataFormat:      core.DataLoadingConfig_JSON.String(),
		downloadMode:        core.IOStrategy_DOWNLOAD_EAGER.String(),
	}

	collectFile := func(d string) []string {
		var files []string
		assert.NoError(t, filepath.Walk(d, func(path string, info os.FileInfo, err error) error {
			if !strings.Contains(info.Name(), tmpPrefix) {
				files = append(files, info.Name())
			} // Skip tmp folder
			return nil
		}))
		sort.Strings(files)
		return files
	}

	t.Run("emptyInputs", func(t *testing.T) {
		tmpDir, err := ioutil.TempDir(tmpFolderLocation, tmpPrefix)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, os.RemoveAll(tmpDir))
		}()
		dopts.localDirectoryPath = tmpDir

		s := promutils.NewTestScope()
		store, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, s.NewSubScope("storage"))
		assert.NoError(t, err)
		dopts.RootOptions = &RootOptions{
			Scope: s,
			Store: store,
		}

		assert.NoError(t, store.WriteProtobuf(ctx, storage.DataReference(inputPath), storage.Options{}, &core.LiteralMap{}))
		assert.NoError(t, dopts.Download(ctx))

		assert.Len(t, collectFile(tmpDir), 0)
	})

	t.Run("primitiveInputs", func(t *testing.T) {
		tmpDir, err := ioutil.TempDir(tmpFolderLocation, tmpPrefix)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, os.RemoveAll(tmpDir))
		}()
		dopts.localDirectoryPath = tmpDir

		s := promutils.NewTestScope()
		store, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, s.NewSubScope("storage"))
		assert.NoError(t, err)
		dopts.RootOptions = &RootOptions{
			Scope: s,
			Store: store,
		}

		assert.NoError(t, store.WriteProtobuf(ctx, storage.DataReference(inputPath), storage.Options{}, &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"x": coreutils.MustMakePrimitiveLiteral(1),
				"y": coreutils.MustMakePrimitiveLiteral("hello"),
			},
		}))
		assert.NoError(t, dopts.Download(ctx), "Download Operation failed")
		assert.Equal(t, []string{"inputs.json", "inputs.pb", "x", "y"}, collectFile(tmpDir))
	})

	t.Run("primitiveAndBlobInputs", func(t *testing.T) {
		tmpDir, err := ioutil.TempDir(tmpFolderLocation, tmpPrefix)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, os.RemoveAll(tmpDir))
		}()
		dopts.localDirectoryPath = tmpDir

		s := promutils.NewTestScope()
		store, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, s.NewSubScope("storage"))
		assert.NoError(t, err)
		dopts.RootOptions = &RootOptions{
			Scope: s,
			Store: store,
		}

		blobLoc := storage.DataReference("blob-loc")
		br := bytes.NewBuffer([]byte("Hello World!"))
		assert.NoError(t, store.WriteRaw(ctx, blobLoc, int64(br.Len()), storage.Options{}, br))
		assert.NoError(t, store.WriteProtobuf(ctx, storage.DataReference(inputPath), storage.Options{}, &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"x": coreutils.MustMakePrimitiveLiteral(1),
				"y": coreutils.MustMakePrimitiveLiteral("hello"),
				"blob": {Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Blob{
							Blob: &core.Blob{
								Uri: blobLoc.String(),
								Metadata: &core.BlobMetadata{
									Type: &core.BlobType{
										Dimensionality: core.BlobType_SINGLE,
										Format:         ".xyz",
									},
								},
							},
						},
					},
				}},
			},
		}))
		assert.NoError(t, dopts.Download(ctx), "Download Operation failed")
		assert.ElementsMatch(t, []string{"inputs.json", "inputs.pb", "x", "y", "blob"}, collectFile(tmpDir))
	})

	t.Run("primitiveAndMissingBlobInputs", func(t *testing.T) {
		tmpDir, err := ioutil.TempDir(tmpFolderLocation, tmpPrefix)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, os.RemoveAll(tmpDir))
		}()
		dopts.localDirectoryPath = tmpDir

		s := promutils.NewTestScope()
		store, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, s.NewSubScope("storage"))
		assert.NoError(t, err)
		dopts.RootOptions = &RootOptions{
			Scope:           s,
			Store:           store,
			errorOutputName: "errors.pb",
		}

		assert.NoError(t, store.WriteProtobuf(ctx, storage.DataReference(inputPath), storage.Options{}, &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"x": coreutils.MustMakePrimitiveLiteral(1),
				"y": coreutils.MustMakePrimitiveLiteral("hello"),
				"blob": {Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Blob{
							Blob: &core.Blob{
								Uri: "blob",
								Metadata: &core.BlobMetadata{
									Type: &core.BlobType{
										Dimensionality: core.BlobType_SINGLE,
										Format:         ".xyz",
									},
								},
							},
						},
					},
				}},
			},
		}))
		err = dopts.Download(ctx)
		assert.NoError(t, err, "Download Operation failed")
		errFile, err := store.ConstructReference(ctx, storage.DataReference(outputPath), "errors.pb")
		assert.NoError(t, err)
		errProto := &core.ErrorDocument{}
		err = store.ReadProtobuf(ctx, errFile, errProto)
		assert.NoError(t, err)
		if assert.NotNil(t, errProto.Error) {
			assert.Equal(t, core.ContainerError_RECOVERABLE, errProto.Error.Kind)
		}
	})
}
