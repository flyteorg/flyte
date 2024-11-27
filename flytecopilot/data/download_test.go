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

const toPath = "./inputs"

func TestHandleBlobMultipart(t *testing.T) {
	t.Run("Successful Query", func(t *testing.T) {
		s, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)
		ref := storage.DataReference("s3://container/folder/file1")
		err = s.WriteRaw(context.Background(), ref, 0, storage.Options{}, bytes.NewReader([]byte{}))
		assert.NoError(t, err)
		ref = storage.DataReference("s3://container/folder/file2")
		err = s.WriteRaw(context.Background(), ref, 0, storage.Options{}, bytes.NewReader([]byte{}))
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
	err = s.WriteRaw(context.Background(), ref, 0, storage.Options{}, bytes.NewReader([]byte{}))
	assert.NoError(t, err)

	d := Downloader{store: s}

	blob := &core.Blob{
		Uri: "s3://container/file",
		Metadata: &core.BlobMetadata{
			Type: &core.BlobType{
				Dimensionality: core.BlobType_SINGLE,
			},
		},
	}

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

func TestRecursiveDownload(t *testing.T) {
	t.Run("OffloadedMetadataContainsCollectionOfStrings", func(t *testing.T) {
		s, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)

		d := Downloader{store: s}

		offloadedLiteral := &core.Literal{
			Value: &core.Literal_OffloadedMetadata{
				OffloadedMetadata: &core.LiteralOffloadedMetadata{
					Uri: "s3://container/offloaded",
				},
			},
		}

		inputs := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"input1": offloadedLiteral,
			},
		}

		// Mock reading the offloaded metadata
		err = s.WriteProtobuf(context.Background(), storage.DataReference("s3://container/offloaded"), storage.Options{}, &core.Literal{
			Value: &core.Literal_Collection{
				Collection: &core.LiteralCollection{
					Literals: []*core.Literal{
						{
							Value: &core.Literal_Scalar{
								Scalar: &core.Scalar{
									Value: &core.Scalar_Primitive{
										Primitive: &core.Primitive{
											Value: &core.Primitive_StringValue{
												StringValue: "string1",
											},
										},
									},
								},
							},
						},
						{
							Value: &core.Literal_Scalar{
								Scalar: &core.Scalar{
									Value: &core.Scalar_Primitive{
										Primitive: &core.Primitive{
											Value: &core.Primitive_StringValue{
												StringValue: "string2",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		})
		assert.NoError(t, err)

		defer func() {
			err := os.RemoveAll(toPath)
			if err != nil {
				t.Errorf("Failed to delete directory: %v", err)
			}
		}()

		varMap, lMap, err := d.RecursiveDownload(context.Background(), inputs, toPath, true)
		assert.NoError(t, err)
		assert.NotNil(t, varMap)
		assert.NotNil(t, lMap)
		assert.Equal(t, []interface{}{"string1", "string2"}, varMap["input1"])
		// Check if files were created and data written
		for _, file := range []string{"0", "1"} {
			if _, err := os.Stat(filepath.Join(toPath, "input1", file)); os.IsNotExist(err) {
				t.Errorf("expected file %s to exist", file)
			}
		}
	})

	t.Run("OffloadedMetadataContainsMapOfStringString", func(t *testing.T) {
		s, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)

		d := Downloader{store: s}

		offloadedLiteral := &core.Literal{
			Value: &core.Literal_OffloadedMetadata{
				OffloadedMetadata: &core.LiteralOffloadedMetadata{
					Uri: "s3://container/offloaded",
				},
			},
		}

		inputs := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"input1": offloadedLiteral,
			},
		}

		// Mock reading the offloaded metadata
		err = s.WriteProtobuf(context.Background(), storage.DataReference("s3://container/offloaded"), storage.Options{}, &core.Literal{
			Value: &core.Literal_Map{
				Map: &core.LiteralMap{
					Literals: map[string]*core.Literal{
						"key1": {
							Value: &core.Literal_Scalar{
								Scalar: &core.Scalar{
									Value: &core.Scalar_Primitive{
										Primitive: &core.Primitive{
											Value: &core.Primitive_StringValue{
												StringValue: "value1",
											},
										},
									},
								},
							},
						},
						"key2": {
							Value: &core.Literal_Scalar{
								Scalar: &core.Scalar{
									Value: &core.Scalar_Primitive{
										Primitive: &core.Primitive{
											Value: &core.Primitive_StringValue{
												StringValue: "value2",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		})
		assert.NoError(t, err)
		defer func() {
			err := os.RemoveAll(toPath)
			if err != nil {
				t.Errorf("Failed to delete directory: %v", err)
			}
		}()

		varMap, lMap, err := d.RecursiveDownload(context.Background(), inputs, toPath, true)
		assert.NoError(t, err)
		assert.NotNil(t, varMap)
		assert.NotNil(t, lMap)
		assert.Equal(t, "value1", varMap["input1"].(VarMap)["key1"])
		assert.Equal(t, "value2", varMap["input1"].(VarMap)["key2"])

		for _, file := range []string{"key1", "key2"} {
			if _, err := os.Stat(filepath.Join(toPath, "input1", file)); os.IsNotExist(err) {
				t.Errorf("expected file %s to exist", file)
			}
		}
	})
}
