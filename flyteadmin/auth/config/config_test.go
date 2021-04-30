package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/fosite"
)

func TestHashFlyteClientSecret(t *testing.T) {
	hasher := &fosite.BCrypt{WorkFactor: 6}
	res, err := hasher.Hash(context.Background(), []byte("foobar"))
	assert.NoError(t, err)
	t.Log(string(res))
}
