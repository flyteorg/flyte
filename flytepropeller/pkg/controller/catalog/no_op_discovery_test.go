package catalog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var noopDiscovery Client = &NoOpDiscovery{}

func TestNoopDiscovery_Get(t *testing.T) {
	ctx := context.Background()
	resp, err := noopDiscovery.Get(ctx, nil, "")
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.True(t, status.Code(err) == codes.NotFound)
}

func TestNoopDiscovery_Put(t *testing.T) {
	ctx := context.Background()
	err := noopDiscovery.Put(ctx, nil, nil, "", "")
	assert.Nil(t, err)
}
