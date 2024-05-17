package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadCACerts(t *testing.T) {

	t.Run("legal", func(t *testing.T) {
		x509Pool, err := readCACerts("testdata/root.pem")
		assert.NoError(t, err)
		assert.NotNil(t, x509Pool)
	})

	t.Run("illegal", func(t *testing.T) {
		x509Pool, err := readCACerts("testdata/invalid-root.pem")
		assert.NotNil(t, err)
		assert.Nil(t, x509Pool)
	})

	t.Run("non-existent", func(t *testing.T) {
		x509Pool, err := readCACerts("testdata/non-existent.pem")
		assert.NotNil(t, err)
		assert.Nil(t, x509Pool)
	})
}
