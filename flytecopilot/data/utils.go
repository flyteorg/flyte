package data

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/pkg/errors"
)

// Checks if the given filepath is a valid and existing file path. If ignoreExtension is true, then the dir + basepath is checked for existence
// ignoring the extension.
// In the return the first return value is the actual path that exists (with the extension), second argument is the file info and finally the error
func IsFileReadable(fpath string, ignoreExtension bool) (string, os.FileInfo, error) {
	info, err := os.Stat(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			if ignoreExtension {
				logger.Infof(context.TODO(), "looking for any extensions")
				matches, err := filepath.Glob(fpath + ".*")
				if err == nil && len(matches) == 1 {
					logger.Infof(context.TODO(), "Extension match found [%s]", matches[0])
					info, err = os.Stat(matches[0])
					if err == nil {
						return matches[0], info, nil
					}
				} else {
					logger.Errorf(context.TODO(), "Extension match not found [%v,%v]", err, matches)
				}
			}
			return "", nil, errors.Wrapf(err, "file not found at path [%s]", fpath)
		}
		if os.IsPermission(err) {
			return "", nil, errors.Wrapf(err, "unable to read file [%s], Flyte does not have permissions", fpath)
		}
		return "", nil, errors.Wrapf(err, "failed to read file")
	}
	return fpath, info, nil
}

// Uploads a file to the data store.
func UploadFileToStorage(ctx context.Context, filePath string, toPath storage.DataReference, size int64, store *storage.DataStore) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			logger.Errorf(ctx, "failed to close blob file at path [%s]", filePath)
		}
	}()
	return store.WriteRaw(ctx, toPath, size, storage.Options{}, f)
}

func DownloadFileFromStorage(ctx context.Context, ref storage.DataReference, store *storage.DataStore) (io.ReadCloser, error) {
	// We should probably directly use stow!??
	m, err := store.Head(ctx, ref)
	if err != nil {
		return nil, errors.Wrapf(err, "failed when looking up Blob")
	}
	if m.Exists() {
		r, err := store.ReadRaw(ctx, ref)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read Blob from storage")
		}
		return r, err

	}
	return nil, fmt.Errorf("incorrect blob reference, does not exist")
}

// Downloads data from the given HTTP URL. If context is canceled then the request will be canceled.
func DownloadFileFromHTTP(ctx context.Context, ref storage.DataReference) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ref.String(), nil)
	if err != nil {
		logger.Errorf(ctx, "failed to create new http request with context, %s", err)
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to download from url :%s", ref)
	}
	return resp.Body, nil
}
