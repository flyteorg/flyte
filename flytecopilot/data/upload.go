package data

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/futures"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"
)

const maxPrimitiveSize = 1024

type Unmarshal func(r io.Reader, msg proto.Message) error
type Uploader struct {
	format core.DataLoadingConfig_LiteralMapFormat
	mode   core.IOStrategy_UploadMode
	// TODO support multiple buckets
	store                   *storage.DataStore
	aggregateOutputFileName string
	errorFileName           string
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
	b, err := ioutil.ReadFile(fpath)
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
				logger.Warnf(ctx, "Currently nested directories are not supported in multipart blob uploads, for directory @ %s", path)
			} else {
				ref, err := u.store.ConstructReference(ctx, toPath, info.Name())
				if err != nil {
					return err
				}
				files = append(files, dirFile{
					path: path,
					info: info,
					ref:  ref,
				})
			}
			return nil
		})
		if err != nil {
			return nil, err
		}

		childCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		fileUploader := make([]futures.Future, 0, len(files))
		for _, f := range files {
			pth := f.path
			ref := f.ref
			size := f.info.Size()
			fileUploader = append(fileUploader, futures.NewAsyncFuture(childCtx, func(i2 context.Context) (i interface{}, e error) {
				return nil, UploadFileToStorage(i2, pth, ref, size, u.store)
			}))
		}

		for _, f := range fileUploader {
			// TODO maybe we should have timeouts, or we can have a global timeout at the top level
			_, err := f.Get(ctx)
			if err != nil {
				return nil, err
			}
		}

		return coreutils.MakeLiteralForBlob(toPath, false, ""), nil
	}
	size := info.Size()
	// Should we make this a go routine as well, so that we can introduce timeouts
	return coreutils.MakeLiteralForBlob(toPath, false, ""), UploadFileToStorage(ctx, fpath, toPath, size, u.store)
}

func (u Uploader) RecursiveUpload(ctx context.Context, vars *core.VariableMap, fromPath string, metaOutputPath, dataRawPath storage.DataReference) error {
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errFile := path.Join(fromPath, u.errorFileName)
	if info, err := os.Stat(errFile); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else if info.Size() > 1024*1024 {
		return fmt.Errorf("error file too large %d", info.Size())
	} else if info.IsDir() {
		return fmt.Errorf("error file is a directory")
	} else {
		b, err := ioutil.ReadFile(errFile)
		if err != nil {
			return err
		}
		return errors.Errorf("User Error: %s", string(b))
	}

	varFutures := make(map[string]futures.Future, len(vars.Variables))
	for varName, variable := range vars.Variables {
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
	}
}
