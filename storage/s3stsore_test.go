package storage

import (
	"testing"

	"github.com/lyft/flytestdlib/promutils"

	"github.com/lyft/flytestdlib/config"
	"github.com/lyft/flytestdlib/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewS3RawStore(t *testing.T) {
	t.Run("Missing required config", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		_, err := NewDataStore(&Config{
			Type:          TypeMinio,
			InitContainer: "some-container",
			Connection: ConnectionConfig{
				Endpoint: config.URL{URL: utils.MustParseURL("http://minio:9000")},
			},
		}, testScope)

		assert.Error(t, err)
	})
}
