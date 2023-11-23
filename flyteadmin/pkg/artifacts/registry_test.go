package artifacts

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegistryNoClient(t *testing.T) {
	r := NewArtifactRegistry(context.Background(), nil)
	assert.Nil(t, r.GetClient())
}

type Parent struct {
	R *ArtifactRegistry
}

func TestPointerReceivers(t *testing.T) {
	p := Parent{}
	nilClient := p.R.GetClient()
	assert.Nil(t, nilClient)
}

func TestNilCheck(t *testing.T) {
	r := NewArtifactRegistry(context.Background(), nil)
	err := r.RegisterTrigger(context.Background(), nil)
	assert.NotNil(t, err)
}
