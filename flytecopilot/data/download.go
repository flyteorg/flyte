package data

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"

	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/pkg/errors"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/futures"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type Downloader struct {
	format core.DataLoadingConfig_LiteralMapFormat
	store  *storage.DataStore
	// TODO support download mode
	mode core.IOStrategy_DownloadMode
}

// TODO add timeout and rate limit
// TODO use chunk to download
func (d Downloader) handleBlob(ctx context.Context, blob *core.Blob, toPath string) (interface{}, error) {
	/*
	   handleBlob handles the retrieval and local storage of blob data, including support for both single and multipart blob types.
	   For multipart blobs, it lists all parts recursively and spawns concurrent goroutines to download each part while managing file I/O in parallel.

	   - The function begins by validating the blob URI and categorizing the blob type (single or multipart).
	   - In the multipart case, it recursively lists all blob parts and launches goroutines to download and save each part.
	     Goroutine closure and I/O success tracking are managed to avoid resource leaks.
	   - For single-part blobs, it directly downloads and writes the data to the specified path.

	   Life Cycle:
	   1. Blob URI                    -> Blob Metadata Type check       -> Recursive List parts if Multipart          -> Launch goroutines to download parts
	      (input blob object)            (determine multipart/single)       (List API, handles recursive case)              (each part handled in parallel)
	   2. Download part or full blob   -> Save locally with error checks -> Handle reader/writer closures             -> Return local path or error
	      (download each part)             (error on write or directory)     (close streams safely, track success)        (completion or report missing closures)
	*/

	blobRef := storage.DataReference(blob.Uri)
	scheme, _, _, err := blobRef.Split()
	if err != nil {
		return nil, errors.Wrapf(err, "Blob uri incorrectly formatted")
	}

	if blob.GetMetadata().GetType().Dimensionality == core.BlobType_MULTIPART {
		// Collect all parts of the multipart blob recursively (List API handles nested directories)
		// Set maxItems to 100 as a parameter for the List API, enabling batch retrieval of items until all are downloaded
		maxItems := 100
		cursor := storage.NewCursorAtStart()
		var items []storage.DataReference
		var absPaths []string
		for {
			items, cursor, err = d.store.List(ctx, blobRef, maxItems, cursor)
			if err != nil || len(items) == 0 {
				logger.Errorf(ctx, "failed to collect items from multipart blob [%s]", blobRef)
				return nil, err
			}
			for _, item := range items {
				absPaths = append(absPaths, item.String())
			}
			if storage.IsCursorEnd(cursor) {
				break
			}
		}

		// Track the count of successful downloads and the total number of items
		downloadSuccess := 0
		itemCount := len(absPaths)
		// Track successful closures of readers and writers in deferred functions
		readerCloseSuccessCount := 0
		writerCloseSuccessCount := 0
		// We use Mutex to avoid race conditions when updating counters and creating directories
		var mu sync.Mutex
		var wg sync.WaitGroup
		for _, absPath := range absPaths {
			absPath := absPath

			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() {
					if err := recover(); err != nil {
						logger.Errorf(ctx, "recover receives error: %s", err)
					}
				}()

				ref := storage.DataReference(absPath)
				reader, err := DownloadFileFromStorage(ctx, ref, d.store)
				if err != nil {
					logger.Errorf(ctx, "Failed to download from ref [%s]", ref)
					return
				}
				defer func() {
					err := reader.Close()
					if err != nil {
						logger.Errorf(ctx, "failed to close Blob read stream @ref [%s]. Error: %s", ref, err)
					}
					mu.Lock()
					readerCloseSuccessCount += 1
					mu.Unlock()
				}()

				_, _, k, err := ref.Split()
				if err != nil {
					logger.Errorf(ctx, "Failed to parse ref [%s]", ref)
					return
				}
				newPath := filepath.Join(toPath, k)
				dir := filepath.Dir(newPath)

				mu.Lock()
				// os.MkdirAll creates the specified directory structure if it doesn’t already exist
				// 0777: the directory can be read and written by anyone
				os.MkdirAll(dir, 0777)
				mu.Unlock()
				writer, err := os.Create(newPath)
				if err != nil {
					logger.Errorf(ctx, "failed to open file at path %s", newPath)
					return
				}
				defer func() {
					err := writer.Close()
					if err != nil {
						logger.Errorf(ctx, "failed to close File write stream. Error: %s", err)
					}
					mu.Lock()
					writerCloseSuccessCount += 1
					mu.Unlock()
				}()

				_, err = io.Copy(writer, reader)
				if err != nil {
					logger.Errorf(ctx, "failed to write remote data to local filesystem")
					return
				}
				mu.Lock()
				downloadSuccess += 1
				mu.Unlock()
			}()
		}
		// Go routines are synchronized with a WaitGroup to prevent goroutine leaks.
		wg.Wait()
		if downloadSuccess != itemCount || readerCloseSuccessCount != itemCount || writerCloseSuccessCount != itemCount {
			return nil, errors.Errorf(
				"Failed to copy %d out of %d remote files from [%s] to local [%s]. Failed to close %d readers; Failed to close %d writers.",
				itemCount-downloadSuccess, itemCount, blobRef, toPath, itemCount-readerCloseSuccessCount, itemCount-writerCloseSuccessCount,
			)
		}
		logger.Infof(ctx, "successfully copied %d remote files from [%s] to local [%s]", downloadSuccess, blobRef, toPath)
		return toPath, nil
	} else if blob.GetMetadata().GetType().Dimensionality == core.BlobType_SINGLE {
		// reader should be declared here (avoid being shared across all goroutines)
		var reader io.ReadCloser
		if scheme == "http" || scheme == "https" {
			reader, err = DownloadFileFromHTTP(ctx, blobRef)
		} else {
			reader, err = DownloadFileFromStorage(ctx, blobRef, d.store)
		}
		if err != nil {
			logger.Errorf(ctx, "Failed to download from ref [%s]", blobRef)
			return nil, err
		}
		defer func() {
			err := reader.Close()
			if err != nil {
				logger.Errorf(ctx, "failed to close Blob read stream @ref [%s]. Error: %s", blobRef, err)
			}
		}()

		writer, err := os.Create(toPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open file at path %s", toPath)
		}
		defer func() {
			err := writer.Close()
			if err != nil {
				logger.Errorf(ctx, "failed to close File write stream. Error: %s", err)
			}
		}()
		v, err := io.Copy(writer, reader)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to write remote data to local filesystem")
		}
		logger.Infof(ctx, "Successfully copied [%d] bytes remote data from [%s] to local [%s]", v, blobRef, toPath)
		return toPath, nil
	}

	return nil, errors.Errorf("unexpected Blob type encountered")
}

func (d Downloader) handleSchema(ctx context.Context, schema *core.Schema, toFilePath string) (interface{}, error) {
	return d.handleBlob(ctx, &core.Blob{Uri: schema.Uri, Metadata: &core.BlobMetadata{Type: &core.BlobType{Dimensionality: core.BlobType_MULTIPART}}}, toFilePath)
}

func (d Downloader) handleBinary(_ context.Context, b *core.Binary, toFilePath string, writeToFile bool) (interface{}, error) {
	// maybe we should return a map
	v := b.GetValue()
	if writeToFile {
		return v, ioutil.WriteFile(toFilePath, v, os.ModePerm)
	}
	return v, nil
}

func (d Downloader) handleError(_ context.Context, b *core.Error, toFilePath string, writeToFile bool) (interface{}, error) {
	// maybe we should return a map
	if writeToFile {
		return b.Message, ioutil.WriteFile(toFilePath, []byte(b.Message), os.ModePerm)
	}
	return b.Message, nil
}

func (d Downloader) handleGeneric(ctx context.Context, b *structpb.Struct, toFilePath string, writeToFile bool) (interface{}, error) {
	if writeToFile && b != nil {
		m := jsonpb.Marshaler{}
		writer, err := os.Create(toFilePath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open file at path %s", toFilePath)
		}
		defer func() {
			err := writer.Close()
			if err != nil {
				logger.Errorf(ctx, "failed to close File write stream. Error: %s", err)
			}
		}()
		return b, m.Marshal(writer, b)
	}
	return b, nil
}

// Returns the primitive value in Golang native format and if the filePath is not empty, then writes the value to the given file path.
func (d Downloader) handlePrimitive(primitive *core.Primitive, toFilePath string, writeToFile bool) (interface{}, error) {

	var toByteArray func() ([]byte, error)
	var v interface{}
	var err error

	switch primitive.Value.(type) {
	case *core.Primitive_StringValue:
		v = primitive.GetStringValue()
		toByteArray = func() ([]byte, error) {
			return []byte(primitive.GetStringValue()), nil
		}
	case *core.Primitive_Boolean:
		v = primitive.GetBoolean()
		toByteArray = func() ([]byte, error) {
			return []byte(strconv.FormatBool(primitive.GetBoolean())), nil
		}
	case *core.Primitive_Integer:
		v = primitive.GetInteger()
		toByteArray = func() ([]byte, error) {
			return []byte(strconv.FormatInt(primitive.GetInteger(), 10)), nil
		}
	case *core.Primitive_FloatValue:
		v = primitive.GetFloatValue()
		toByteArray = func() ([]byte, error) {
			return []byte(strconv.FormatFloat(primitive.GetFloatValue(), 'f', -1, 64)), nil
		}
	case *core.Primitive_Datetime:
		v, err = ptypes.Timestamp(primitive.GetDatetime())
		if err != nil {
			return nil, err
		}
		toByteArray = func() ([]byte, error) {
			return []byte(ptypes.TimestampString(primitive.GetDatetime())), nil
		}
	case *core.Primitive_Duration:
		v, err := ptypes.Duration(primitive.GetDuration())
		if err != nil {
			return nil, err
		}
		toByteArray = func() ([]byte, error) {
			return []byte(v.String()), nil
		}
	default:
		v = nil
		toByteArray = func() ([]byte, error) {
			return []byte("null"), nil
		}
	}
	if writeToFile {
		b, err := toByteArray()
		if err != nil {
			return nil, err
		}
		return v, ioutil.WriteFile(toFilePath, b, os.ModePerm)
	}
	return v, nil
}

func (d Downloader) handleScalar(ctx context.Context, scalar *core.Scalar, toFilePath string, writeToFile bool) (interface{}, *core.Scalar, error) {
	switch scalar.GetValue().(type) {
	case *core.Scalar_Primitive:
		p := scalar.GetPrimitive()
		i, err := d.handlePrimitive(p, toFilePath, writeToFile)
		return i, scalar, err
	case *core.Scalar_Blob:
		b := scalar.GetBlob()
		i, err := d.handleBlob(ctx, b, toFilePath)
		return i, &core.Scalar{Value: &core.Scalar_Blob{Blob: &core.Blob{Metadata: b.Metadata, Uri: toFilePath}}}, err
	case *core.Scalar_Schema:
		b := scalar.GetSchema()
		i, err := d.handleSchema(ctx, b, toFilePath)
		return i, &core.Scalar{Value: &core.Scalar_Schema{Schema: &core.Schema{Type: b.Type, Uri: toFilePath}}}, err
	case *core.Scalar_Binary:
		b := scalar.GetBinary()
		i, err := d.handleBinary(ctx, b, toFilePath, writeToFile)
		return i, scalar, err
	case *core.Scalar_Error:
		b := scalar.GetError()
		i, err := d.handleError(ctx, b, toFilePath, writeToFile)
		return i, scalar, err
	case *core.Scalar_Generic:
		b := scalar.GetGeneric()
		i, err := d.handleGeneric(ctx, b, toFilePath, writeToFile)
		return i, scalar, err
	case *core.Scalar_NoneType:
		if writeToFile {
			return nil, scalar, ioutil.WriteFile(toFilePath, []byte("null"), os.ModePerm)
		}
		return nil, scalar, nil
	default:
		return nil, nil, fmt.Errorf("unsupported scalar type [%v]", reflect.TypeOf(scalar.GetValue()))
	}
}

func (d Downloader) handleLiteral(ctx context.Context, lit *core.Literal, filePath string, writeToFile bool) (interface{}, *core.Literal, error) {
	switch lit.GetValue().(type) {
	case *core.Literal_Scalar:
		v, s, err := d.handleScalar(ctx, lit.GetScalar(), filePath, writeToFile)
		if err != nil {
			return nil, nil, err
		}
		return v, &core.Literal{Value: &core.Literal_Scalar{
			Scalar: s,
		}}, nil
	case *core.Literal_Collection:
		v, c2, err := d.handleCollection(ctx, lit.GetCollection(), filePath, writeToFile)
		if err != nil {
			return nil, nil, err
		}
		return v, &core.Literal{Value: &core.Literal_Collection{
			Collection: c2,
		}}, nil
	case *core.Literal_Map:
		v, m, err := d.RecursiveDownload(ctx, lit.GetMap(), filePath, writeToFile)
		if err != nil {
			return nil, nil, err
		}
		return v, &core.Literal{Value: &core.Literal_Map{
			Map: m,
		}}, nil
	default:
		return nil, nil, fmt.Errorf("unsupported literal type [%v]", reflect.TypeOf(lit.GetValue()))
	}
}

// Collection should be stored as a top level list file and may have accompanying files?
func (d Downloader) handleCollection(ctx context.Context, c *core.LiteralCollection, dir string, writePrimitiveToFile bool) ([]interface{}, *core.LiteralCollection, error) {
	if c == nil || len(c.Literals) == 0 {
		return []interface{}{}, c, nil
	}
	var collection []interface{}
	litCollection := &core.LiteralCollection{}
	for i, lit := range c.Literals {
		filePath := path.Join(dir, strconv.Itoa(i))
		v, lit, err := d.handleLiteral(ctx, lit, filePath, writePrimitiveToFile)
		if err != nil {
			return nil, nil, err
		}
		collection = append(collection, v)
		litCollection.Literals = append(litCollection.Literals, lit)
	}
	return collection, litCollection, nil
}

type downloadedResult struct {
	lit *core.Literal
	v   interface{}
}

func (d Downloader) RecursiveDownload(ctx context.Context, inputs *core.LiteralMap, dir string, writePrimitiveToFile bool) (VarMap, *core.LiteralMap, error) {
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	if inputs == nil || len(inputs.Literals) == 0 {
		return VarMap{}, nil, nil
	}
	f := make(FutureMap, len(inputs.Literals))
	for variable, literal := range inputs.Literals {
		varPath := path.Join(dir, variable)
		lit := literal
		f[variable] = futures.NewAsyncFuture(childCtx, func(ctx2 context.Context) (interface{}, error) {
			v, lit, err := d.handleLiteral(ctx2, lit, varPath, writePrimitiveToFile)
			if err != nil {
				return nil, err
			}
			return downloadedResult{lit: lit, v: v}, nil
		})
	}

	m := &core.LiteralMap{
		Literals: make(map[string]*core.Literal),
	}
	vmap := make(VarMap, len(f))
	for variable, future := range f {
		logger.Infof(ctx, "Waiting for [%s] to be persisted", variable)
		v, err := future.Get(childCtx)
		if err != nil {
			logger.Errorf(ctx, "Failed to persist [%s], err %s", variable, err)
			if err == futures.ErrAsyncFutureCanceled {
				logger.Errorf(ctx, "Future was canceled, possibly Timeout!")
			}
			return nil, nil, errors.Wrapf(err, "variable [%s] download/store failed", variable)
		}
		dr := v.(downloadedResult)
		vmap[variable] = dr.v
		m.Literals[variable] = dr.lit
		logger.Infof(ctx, "Completed persisting [%s]", variable)
	}

	return vmap, m, nil
}

func (d Downloader) DownloadInputs(ctx context.Context, inputRef storage.DataReference, outputDir string) error {
	logger.Infof(ctx, "Downloading inputs from [%s]", inputRef)
	defer logger.Infof(ctx, "Exited downloading inputs from [%s]", inputRef)
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		logger.Errorf(ctx, "Failed to create output directories, err: %s", err)
		return err
	}
	inputs := &core.LiteralMap{}
	err := d.store.ReadProtobuf(ctx, inputRef, inputs)
	if err != nil {
		logger.Errorf(ctx, "Failed to download inputs from [%s], err [%s]", inputRef, err)
		return errors.Wrapf(err, "failed to download input metadata message from remote store")
	}
	varMap, lMap, err := d.RecursiveDownload(ctx, inputs, outputDir, true)
	if err != nil {
		return errors.Wrapf(err, "failed to download input variable from remote store")
	}

	// We will always write the protobuf
	b, err := proto.Marshal(lMap)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(outputDir, "inputs.pb"), b, os.ModePerm); err != nil {
		return err
	}

	if d.format == core.DataLoadingConfig_JSON {
		m, err := json.Marshal(varMap)
		if err != nil {
			return errors.Wrapf(err, "failed to marshal out inputs")
		}
		return ioutil.WriteFile(path.Join(outputDir, "inputs.json"), m, os.ModePerm)
	}
	if d.format == core.DataLoadingConfig_YAML {
		m, err := yaml.Marshal(varMap)
		if err != nil {
			return errors.Wrapf(err, "failed to marshal out inputs")
		}
		return ioutil.WriteFile(path.Join(outputDir, "inputs.yaml"), m, os.ModePerm)
	}
	return nil
}

func NewDownloader(_ context.Context, store *storage.DataStore, format core.DataLoadingConfig_LiteralMapFormat, mode core.IOStrategy_DownloadMode) Downloader {
	return Downloader{
		format: format,
		store:  store,
		mode:   mode,
	}
}
