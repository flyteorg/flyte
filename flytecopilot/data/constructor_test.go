package data

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func TestNewDownloaderSanitizesConcurrencyPerCPU(t *testing.T) {
	store, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	d := NewDownloader(context.Background(), store, core.DataLoadingConfig_YAML, core.IOStrategy_DOWNLOAD_EAGER, 0)
	assert.Equal(t, core.DataLoadingConfig_YAML, d.format)
	assert.Equal(t, store, d.store)
	assert.Equal(t, core.IOStrategy_DOWNLOAD_EAGER, d.mode)
	assert.Equal(t, DefaultConcurrencyPerCPU, d.concurrencyPerCPU)

	d = NewDownloader(context.Background(), store, core.DataLoadingConfig_JSON, core.IOStrategy_DOWNLOAD_EAGER, 2)
	assert.Equal(t, 2, d.concurrencyPerCPU)
}

func TestNewUploaderSanitizesConcurrencyPerCPU(t *testing.T) {
	store, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	u := NewUploader(context.Background(), store, core.DataLoadingConfig_YAML, core.IOStrategy_UPLOAD_ON_EXIT, "errors.pb", -10)
	assert.Equal(t, core.DataLoadingConfig_YAML, u.format)
	assert.Equal(t, store, u.store)
	assert.Equal(t, core.IOStrategy_UPLOAD_ON_EXIT, u.mode)
	assert.Equal(t, "errors.pb", u.errorFileName)
	assert.Equal(t, DefaultConcurrencyPerCPU, u.concurrencyPerCPU)

	u = NewUploader(context.Background(), store, core.DataLoadingConfig_JSON, core.IOStrategy_UPLOAD_ON_EXIT, "errors.pb", 3)
	assert.Equal(t, 3, u.concurrencyPerCPU)
}

func TestRecursiveDownloadReturnsEmptyMapForNilOrEmptyInputs(t *testing.T) {
	d := Downloader{}

	varMap, literalMap, err := d.RecursiveDownload(context.Background(), nil, t.TempDir(), true)
	assert.NoError(t, err)
	assert.NotNil(t, varMap)
	assert.Empty(t, varMap)
	assert.Nil(t, literalMap)

	varMap, literalMap, err = d.RecursiveDownload(context.Background(), &core.LiteralMap{}, t.TempDir(), true)
	assert.NoError(t, err)
	assert.NotNil(t, varMap)
	assert.Empty(t, varMap)
	assert.Nil(t, literalMap)
}
