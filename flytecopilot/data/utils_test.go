package data

import (
	"bytes"
	"context"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
)

func TestIsFileReadable(t *testing.T) {
	tmpFolderLocation := ""
	tmpPrefix := "util_test"

	tmpDir, err := os.MkdirTemp(tmpFolderLocation, tmpPrefix)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.RemoveAll(tmpDir))
	}()
	p := path.Join(tmpDir, "x")
	f, i, err := IsFileReadable(p, false)
	assert.Error(t, err)
	assert.Empty(t, f)
	assert.Nil(t, i)

	assert.NoError(t, os.WriteFile(p, []byte("data"), os.ModePerm)) // #nosec G306
	f, i, err = IsFileReadable(p, false)
	assert.NoError(t, err)
	assert.Equal(t, p, f)
	assert.NotNil(t, i)
	assert.Equal(t, p, f)

	noExt := path.Join(tmpDir, "y")
	p = path.Join(tmpDir, "y.png")
	_, _, err = IsFileReadable(noExt, false)
	assert.Error(t, err)

	assert.NoError(t, os.WriteFile(p, []byte("data"), os.ModePerm)) // #nosec G306
	_, _, err = IsFileReadable(noExt, false)
	assert.Error(t, err)

	f, i, err = IsFileReadable(noExt, true)
	assert.NoError(t, err)
	assert.Equal(t, p, f)
	assert.NotNil(t, i)
	assert.Equal(t, p, f)
}

func TestUploadFile(t *testing.T) {
	tmpFolderLocation := ""
	tmpPrefix := "util_test"

	tmpDir, err := os.MkdirTemp(tmpFolderLocation, tmpPrefix)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.RemoveAll(tmpDir))
	}()

	exist := path.Join(tmpDir, "exist-file")
	data := []byte("data")
	l := int64(len(data))
	assert.NoError(t, os.WriteFile(exist, data, os.ModePerm)) // #nosec G306
	nonExist := path.Join(tmpDir, "non-exist-file")

	store, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	ctx := context.TODO()
	assert.NoError(t, UploadFileToStorage(ctx, exist, "exist", l, store))
	m, err := store.Head(ctx, "exist")
	assert.True(t, m.Exists())
	assert.NoError(t, err)

	assert.Error(t, UploadFileToStorage(ctx, nonExist, "nonExist", l, store))
}

func TestDownloadFromHttp(t *testing.T) {
	loc := storage.DataReference("https://raw.githubusercontent.com/flyteorg/flyte/master/README.md")
	badLoc := storage.DataReference("https://no-exist")
	f, err := DownloadFileFromHTTP(context.TODO(), loc)
	if assert.NoError(t, err) {
		if assert.NotNil(t, f) {
			f.Close()
		}
	}

	_, err = DownloadFileFromHTTP(context.TODO(), badLoc)
	assert.Error(t, err)
}

func TestDownloadFromStorage(t *testing.T) {
	store, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)
	ref := storage.DataReference("ref")

	f, err := DownloadFileFromStorage(context.TODO(), ref, store)
	assert.Error(t, err)
	assert.Nil(t, f)

	data := []byte("data")
	l := int64(len(data))

	assert.NoError(t, store.WriteRaw(context.TODO(), ref, l, storage.Options{}, bytes.NewReader(data)))
	f, err = DownloadFileFromStorage(context.TODO(), ref, store)
	if assert.NoError(t, err) {
		assert.NotNil(t, f)
		f.Close()
	}
}

func init() {
	labeled.SetMetricKeys("test")
}
