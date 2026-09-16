package ioutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
)

func TestNewOutputSandbox(t *testing.T) {
	assert.Equal(t, NewRawOutputPaths(context.TODO(), "x").GetRawOutputPrefix(), storage.DataReference("x"))
}
