package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitializeClients(t *testing.T) {
	cfg := defaultConfig
	ctx := context.Background()
	err := SetConfig(&cfg)
	assert.NoError(t, err)
	cs, err := initializeClients(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, cs)
}
