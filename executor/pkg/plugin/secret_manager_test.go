package plugin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoopSecretManager(t *testing.T) {
	m := NewNoopSecretManager()
	assert.NotNil(t, m)

	// No backend: a lookup fails with an error instead of returning an empty secret, so a connector
	// task that references a secret fails loudly rather than running with a blank value.
	val, err := m.Get(context.Background(), "my-secret")
	assert.Error(t, err)
	assert.Empty(t, val)
}
