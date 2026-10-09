package data

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto" //nolint: staticcheck
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/flyteorg/flyte/v2/flyteidl2/clients/go/coreutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/futures"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

const maxPrimitiveSize = 1024

// maxConcurrentFileUploads caps how many files are uploaded at once across all of a task's outputs. The
// storage client can hold a whole file in memory while uploading it (the S3 transfer manager reads up to
// its multipart threshold into memory), so uploading every file of a large directory output at once can
// exceed the sidecar's memory limit.
const maxConcurrentFileUploads = 8

type Unmarshal func(r io.Reader, msg proto.Message) error
type Uploader struct {
	format core.DataLoadingConfig_LiteralMapFormat
	mode   core.IOStrategy_UploadMode
	// TODO support multiple buckets
	store                   *storage.DataStore
	aggregateOutputFileName string
	errorFileName           string
	// fileUploads bounds the number of files uploaded concurrently; it is shared by every output.
	fileUploads *semaphore.Weighted
}

// RawContainerError is the failure raised by the raw container
type RawContainerError struct {
	Document *core.ErrorDocument
}

func (e RawContainerError) Error() string {
	return e.Document.GetError().GetMessage()
}

// readErrorDocument unmarshall the error file into flyte error document
// The unmarshall will fail if it is not a flyte error
func readErrorDocument(raw []byte) (*core.ErrorDocument, bool) {
	document := &core.ErrorDocument{}
	if err := proto.Unmarshal(raw, document); err != nil {
		return nil, false
	}
	if document.GetError().GetCode() == "" && document.GetError().GetMessage() == "" {
		return nil, false
	}
	return document, true
}

type dirFile struct {
	path string
	info os.FileInfo
	ref  storage.DataReference
}

func (u Uploader) handleSimpleType(_ context.Context, t core.SimpleType, filePath string) (*core.Literal, error) {
	fpath, info, err := IsFileReadable(filePath, true)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("expected file for type [%s], found dir at path [%s]", t.String(), filePath)
	}
	if info.Size() > maxPrimitiveSize {
		return nil, fmt.Errorf("maximum allowed filesize is [%d], but found [%d]", maxPrimitiveSize, info.Size())
	}
	b, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	return coreutils.MakeLiteralForSimpleType(t, string(b))
}

func (u Uploader) handleBlobType(ctx context.Context, localPath string, toPath storage.DataReference) (*core.Literal, error) {
	fpath, info, err := IsFileReadable(localPath, true)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		var files []dirFile
		err := filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				logger.Errorf(ctx, "encountered error when uploading multipart blob, %s", err)
				return err
			}
			if info.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(localPath, path)
			if err != nil {
				return err
			}
			keys := strings.Split(filepath.ToSlash(rel), "/")
			ref, err := u.store.ConstructReference(ctx, toPath, keys...)
			if err != nil {
				return err
			}
			files = append(files, dirFile{
				path: path,
				info: info,
				ref:  ref,
			})
			return nil
		})
		if err != nil {
			return nil, err
		}

		g, gCtx := errgroup.WithContext(ctx)
		for _, f := range files {
			g.Go(func() error {
				return u.uploadFile(gCtx, f.path, f.ref, f.info.Size())
			})
		}
		// TODO maybe we should have timeouts, or we can have a global timeout at the top level
		if err := g.Wait(); err != nil {
			return nil, err
		}

		return coreutils.MakeLiteralForBlob(toPath, true, ""), nil
	}
	size := info.Size()
	// Should we make this a go routine as well, so that we can introduce timeouts
	return coreutils.MakeLiteralForBlob(toPath, false, ""), u.uploadFile(ctx, fpath, toPath, size)
}

// uploadFile uploads one file once a slot is free among the concurrent file uploads.
func (u Uploader) uploadFile(ctx context.Context, filePath string, toPath storage.DataReference, size int64) error {
	if err := u.fileUploads.Acquire(ctx, 1); err != nil {
		return err
	}
	defer u.fileUploads.Release(1)
	return UploadFileToStorage(ctx, filePath, toPath, size, u.store)
}

func (u Uploader) RecursiveUpload(ctx context.Context, vars *core.VariableMap, fromPath string, metaOutputPath, dataRawPath storage.DataReference) error {
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errFile := path.Join(fromPath, u.errorFileName)
	// TODO(alex): The error like info.Size() > 1024*1024 should be non-retriable too
	if info, err := os.Stat(errFile); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else if info.Size() > 1024*1024 {
		return fmt.Errorf("error file too large %d", info.Size())
	} else if info.IsDir() {
		return fmt.Errorf("error file is a directory")
	} else {
		b, err := os.ReadFile(errFile)
		if err != nil {
			return err
		}
		if document, ok := readErrorDocument(b); ok {
			return RawContainerError{Document: document}
		}
		return errors.Errorf("User Error: %s", string(b))
	}

	varFutures := make(map[string]futures.Future, len(vars.GetVariables()))
	for _, variableEntry := range vars.GetVariables() {
		varName := variableEntry.GetKey()
		variable := variableEntry.GetValue()
		varPath := path.Join(fromPath, varName)
		varType := variable.GetType()
		switch varType.GetType().(type) {
		case *core.LiteralType_Blob:
			var varOutputPath storage.DataReference
			var err error
			if varName == u.aggregateOutputFileName {
				varOutputPath, err = u.store.ConstructReference(ctx, dataRawPath, "_"+varName)
			} else {
				varOutputPath, err = u.store.ConstructReference(ctx, dataRawPath, varName)
			}
			if err != nil {
				return err
			}
			varFutures[varName] = futures.NewAsyncFuture(childCtx, func(ctx2 context.Context) (interface{}, error) {
				return u.handleBlobType(ctx2, varPath, varOutputPath)
			})
		case *core.LiteralType_Simple:
			varFutures[varName] = futures.NewAsyncFuture(childCtx, func(ctx2 context.Context) (interface{}, error) {
				return u.handleSimpleType(ctx2, varType.GetSimple(), varPath)
			})
		default:
			return fmt.Errorf("currently CoPilot uploader does not support [%s], system error", varType)
		}
	}

	outputs := &core.LiteralMap{
		Literals: make(map[string]*core.Literal, len(varFutures)),
	}
	for k, f := range varFutures {
		logger.Infof(ctx, "Waiting for [%s] to complete (it may have a background upload too)", k)
		v, err := f.Get(ctx)
		if err != nil {
			logger.Errorf(ctx, "Failed to upload [%s], reason [%s]", k, err)
			return err
		}
		l, ok := v.(*core.Literal)
		if !ok {
			return fmt.Errorf("IllegalState, expected core.Literal, received [%s]", reflect.TypeOf(v))
		}
		outputs.Literals[k] = l
		logger.Infof(ctx, "Var [%s] completed", k)
	}

	logger.Infof(ctx, "Uploading final outputs to [%s]", metaOutputPath)
	if err := u.store.WriteProtobuf(ctx, metaOutputPath, storage.Options{}, outputs); err != nil {
		logger.Errorf(ctx, "Failed to upload final outputs file to [%s], err [%s]", metaOutputPath, err)
		return err
	}
	logger.Infof(ctx, "Uploaded final outputs to [%s]", metaOutputPath)
	return nil
}

func NewUploader(_ context.Context, store *storage.DataStore, format core.DataLoadingConfig_LiteralMapFormat, mode core.IOStrategy_UploadMode, errorFileName string) Uploader {
	return Uploader{
		format:        format,
		store:         store,
		errorFileName: errorFileName,
		mode:          mode,
		fileUploads:   semaphore.NewWeighted(maxConcurrentFileUploads),
	}
}
