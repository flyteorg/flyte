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

func TestHandleBlobMultipart(t *testing.T) {
	t.Run("Successful Query", func(t *testing.T) {
		s, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)

		// Create one file at root level and another in a nested folder
		ref1 := storage.DataReference("s3://container/oz/a9ss8w4mnkk8zttr9p7z-n0-0/0fda6abb7c/root_file.txt")
		err = s.WriteRaw(context.Background(), ref1, 0, storage.Options{}, bytes.NewReader([]byte("root content")))
		assert.NoError(t, err)

		ref2 := storage.DataReference("s3://container/oz/a9ss8w4mnkk8zttr9p7z-n0-0/0fda6abb7c/nested/deep_file.txt")
		err = s.WriteRaw(context.Background(), ref2, 0, storage.Options{}, bytes.NewReader([]byte("nested content")))
		assert.NoError(t, err)

		d := Downloader{store: s}

		blob := &core.Blob{
			Uri: "s3://container/oz/a9ss8w4mnkk8zttr9p7z-n0-0/0fda6abb7c",
			Metadata: &core.BlobMetadata{
				Type: &core.BlobType{
					Dimensionality: core.BlobType_MULTIPART,
				},
			},
		}

		// Create temporary directory with inputs/my_var structure
		tmpDir, err := os.MkdirTemp("", "blob_test")
		assert.NoError(t, err)
		testPath := filepath.Join(tmpDir, "inputs", "my_var")
		defer func() {
			err := os.RemoveAll(tmpDir)
			if err != nil {
				t.Errorf("Failed to delete directory: %v", err)
			}
		}()

		result, err := d.handleBlob(context.Background(), blob, testPath)
		assert.NoError(t, err)
		assert.Equal(t, testPath, result)

		// Check if files were created and data written
		// With the new fix, files should be directly under testPath, not in a "folder" subdirectory
		expectedFiles := []string{
			filepath.Join(testPath, "root_file.txt"),
			filepath.Join(testPath, "nested", "deep_file.txt"),
		}

		for _, expectedFile := range expectedFiles {
			if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
				t.Errorf("expected file %s to exist", expectedFile)
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

		toPath := "./inputs"
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

		toPath := "./inputs"
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

func TestHandleScalar(t *testing.T) {
	t.Run("Handles Union Scalar with scalar value", func(t *testing.T) {
		d := Downloader{}

		scalar := &core.Scalar{
			Value: &core.Scalar_Union{
				Union: &core.Union{
					Value: &core.Literal{
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
				},
			},
		}

		resultValue, resultScalar, err := d.handleScalar(context.Background(), scalar, "./inputs", false)
		assert.NoError(t, err)
		assert.Equal(t, "string1", resultValue)
		assert.Equal(t, scalar, resultScalar)
	})
	t.Run("Handles Union Scalar with collection value", func(t *testing.T) {
		d := Downloader{}

		toPath := "./inputs"
		defer func() {
			err := os.RemoveAll(toPath)
			if err != nil {
				t.Errorf("Failed to delete directory: %v", err)
			}
		}()

		scalar := &core.Scalar{
			Value: &core.Scalar_Union{
				Union: &core.Union{
					Value: &core.Literal{
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
					},
				},
			},
		}

		resultValue, resultScalar, err := d.handleScalar(context.Background(), scalar, toPath, false)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"string1", "string2"}, resultValue)
		assert.Equal(t, scalar, resultScalar)
	})
}
