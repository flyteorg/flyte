package data

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func TestHandleBlobMultipart(t *testing.T) {
	t.Run("Successful Query", func(t *testing.T) {
		s, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)

		// Create one file at root level and another in a nested folder
		ref1 := storage.DataReference("mem://container/oz/a9ss8w4mnkk8zttr9p7z-n0-0/0fda6abb7c/root_file.txt")
		err = s.WriteRaw(context.Background(), ref1, 0, storage.Options{}, bytes.NewReader([]byte("root content")))
		assert.NoError(t, err)

		ref2 := storage.DataReference("mem://container/oz/a9ss8w4mnkk8zttr9p7z-n0-0/0fda6abb7c/nested/deep_file.txt")
		err = s.WriteRaw(context.Background(), ref2, 0, storage.Options{}, bytes.NewReader([]byte("nested content")))
		assert.NoError(t, err)

		d := Downloader{store: s}

		blob := &core.Blob{
			Uri: "mem://container/oz/a9ss8w4mnkk8zttr9p7z-n0-0/0fda6abb7c",
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
			Uri: "mem://container/folder",
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
	ref := storage.DataReference("mem://container/file")
	err = s.WriteRaw(context.Background(), ref, 0, storage.Options{}, bytes.NewReader([]byte{}))
	assert.NoError(t, err)

	d := Downloader{store: s}

	blob := &core.Blob{
		Uri: "mem://container/file",
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

func writeBlobLit(t *testing.T, s *storage.DataStore, name string) *core.Literal {
	ref := storage.DataReference("mem://container/" + name)
	assert.NoError(t, s.WriteRaw(context.Background(), ref, 0, storage.Options{}, bytes.NewReader([]byte("data"))))
	return &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{Value: &core.Scalar_Blob{Blob: &core.Blob{
		Uri:      string(ref),
		Metadata: &core.BlobMetadata{Type: &core.BlobType{Dimensionality: core.BlobType_SINGLE}},
	}}}}}
}

func collectionLit(lits ...*core.Literal) *core.Literal {
	return &core.Literal{Value: &core.Literal_Collection{Collection: &core.LiteralCollection{Literals: lits}}}
}

func TestNamedDirLayoutPreservesExtensions(t *testing.T) {
	s, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)
	d := Downloader{store: s, layout: core.DataLoadingConfig_NAMED_DIR}

	dir := "./nd_input"
	assert.NoError(t, os.MkdirAll(dir, os.ModePerm))
	defer func() { _ = os.RemoveAll(dir) }()

	inputs := &core.LiteralMap{Literals: map[string]*core.Literal{
		// list[File]: each element keeps its ORIGINAL basename in the per-input dir.
		"reads": collectionLit(writeBlobLit(t, s, "sample_R1.fastq.gz"), writeBlobLit(t, s, "sample_R2.fastq.gz")),
		// single File: staged inside a per-input dir under its original basename.
		"fasta": writeBlobLit(t, s, "genome.fasta"),
	}}
	_, _, err = d.RecursiveDownload(context.Background(), inputs, dir, true)
	assert.NoError(t, err)

	// Original names round-trip end-to-end (not 0.gz/1.gz).
	for _, rel := range []string{"reads/sample_R1.fastq.gz", "reads/sample_R2.fastq.gz", "fasta/genome.fasta"} {
		_, statErr := os.Stat(filepath.Join(dir, rel))
		assert.False(t, os.IsNotExist(statErr), "expected staged file %s", rel)
	}
}

func TestNamedDirLayoutDedupesCollidingBasenames(t *testing.T) {
	s, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)
	d := Downloader{store: s, layout: core.DataLoadingConfig_NAMED_DIR}

	dir := "./nd_collide"
	assert.NoError(t, os.MkdirAll(dir, os.ModePerm))
	defer func() { _ = os.RemoveAll(dir) }()

	// Two list elements with identical basenames (distinct random prefixes in
	// real URIs, same name here) must not clobber each other.
	inputs := &core.LiteralMap{Literals: map[string]*core.Literal{
		"reads": collectionLit(writeBlobLit(t, s, "a/reads.fastq.gz"), writeBlobLit(t, s, "b/reads.fastq.gz")),
	}}
	_, _, err = d.RecursiveDownload(context.Background(), inputs, dir, true)
	assert.NoError(t, err)

	// First keeps the name; the collision falls back to an index-prefixed name.
	for _, rel := range []string{"reads/reads.fastq.gz", "reads/1_reads.fastq.gz"} {
		_, statErr := os.Stat(filepath.Join(dir, rel))
		assert.False(t, os.IsNotExist(statErr), "expected staged file %s", rel)
	}
}

func TestDirectLayoutIsUnchanged(t *testing.T) {
	s, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)
	d := Downloader{store: s} // layout defaults to DIRECT

	dir := "./direct_input"
	assert.NoError(t, os.MkdirAll(dir, os.ModePerm))
	defer func() { _ = os.RemoveAll(dir) }()

	inputs := &core.LiteralMap{Literals: map[string]*core.Literal{
		"reads": collectionLit(writeBlobLit(t, s, "reads_1.fastq.gz"), writeBlobLit(t, s, "reads_2.fastq.gz")),
		"fasta": writeBlobLit(t, s, "genome.fasta"),
	}}
	_, _, err = d.RecursiveDownload(context.Background(), inputs, dir, true)
	assert.NoError(t, err)

	// list[File] elements stay at <index> (no extension); single File stays
	// directly at <var> (a file) — backwards-compatible.
	for _, fn := range []string{"0", "1"} {
		_, statErr := os.Stat(filepath.Join(dir, "reads", fn))
		assert.False(t, os.IsNotExist(statErr), "expected staged list element %s", fn)
	}
	info, statErr := os.Stat(filepath.Join(dir, "fasta"))
	assert.NoError(t, statErr)
	assert.False(t, info.IsDir(), "single File should stage directly as a file under DIRECT")
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
					Uri: "mem://container/offloaded",
				},
			},
		}

		inputs := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"input1": offloadedLiteral,
			},
		}

		// Mock reading the offloaded metadata
		err = s.WriteProtobuf(context.Background(), storage.DataReference("mem://container/offloaded"), storage.Options{}, &core.Literal{
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
					Uri: "mem://container/offloaded",
				},
			},
		}

		inputs := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"input1": offloadedLiteral,
			},
		}

		// Mock reading the offloaded metadata
		err = s.WriteProtobuf(context.Background(), storage.DataReference("mem://container/offloaded"), storage.Options{}, &core.Literal{
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
