package files

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindConfigFiles(t *testing.T) {
	t.Run("Find config-* group", func(t *testing.T) {
		files := FindConfigFiles([]string{filepath.Join("testdata", "config*.yaml")})
		assert.Equal(t, 2, len(files))
	})

	t.Run("Find other-group-* group", func(t *testing.T) {
		files := FindConfigFiles([]string{filepath.Join("testdata", "other-group*.yaml")})
		assert.Equal(t, 2, len(files))
	})

	t.Run("Absolute path", func(t *testing.T) {
		files := FindConfigFiles([]string{filepath.Join("testdata", "other-group-1.yaml")})
		assert.Equal(t, 1, len(files))

		files = FindConfigFiles([]string{filepath.Join("testdata", "other-group-3.yaml")})
		assert.Equal(t, 0, len(files))
	})
}
