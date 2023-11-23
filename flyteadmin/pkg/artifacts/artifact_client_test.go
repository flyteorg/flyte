package artifacts

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEmpty(t *testing.T) {
	c := InitializeArtifactClient(context.Background(), nil)
	assert.Nil(t, c)
}
