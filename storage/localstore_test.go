package storage

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/lyft/flytestdlib/promutils"

	"github.com/lyft/flytestdlib/config"
	"github.com/lyft/flytestdlib/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewLocalStore(t *testing.T) {
	t.Run("Valid config", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		store, err := newLocalRawStore(&Config{
			Connection: ConnectionConfig{
				Endpoint: config.URL{URL: utils.MustParseURL("./")},
			},
			InitContainer: "testdata",
		}, testScope.NewSubScope("x"))

		assert.NoError(t, err)
		assert.NotNil(t, store)

		// Stow local store expects the full path after the container portion (looks like a bug to me)
		rc, err := store.ReadRaw(context.TODO(), DataReference("file://testdata/testdata/config.yaml"))
		assert.NoError(t, err)
		assert.NotNil(t, rc)
		assert.NoError(t, rc.Close())
	})

	t.Run("Invalid config", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		_, err := newLocalRawStore(&Config{}, testScope)
		assert.Error(t, err)
	})

	t.Run("Initialize container", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		tmpDir, err := ioutil.TempDir("", "stdlib_local")
		assert.NoError(t, err)

		stats, err := os.Stat(tmpDir)
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		store, err := newLocalRawStore(&Config{
			Connection: ConnectionConfig{
				Endpoint: config.URL{URL: utils.MustParseURL(tmpDir)},
			},
			InitContainer: "tmp",
		}, testScope.NewSubScope("y"))

		assert.NoError(t, err)
		assert.NotNil(t, store)

		stats, err = os.Stat(filepath.Join(tmpDir, "tmp"))
		assert.NoError(t, err)
		assert.True(t, stats.IsDir())
	})
}
