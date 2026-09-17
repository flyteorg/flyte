package ioutils

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
)

func TestConstructCheckpointPath(t *testing.T) {
	store := storage.URLPathConstructor{}
	assert.Equal(t, ConstructCheckpointPath(store, "s3://my-bucket/base"),
		storage.DataReference("s3://my-bucket/base/_flytecheckpoints"))
	assert.Equal(t, ConstructCheckpointPath(store, "s3://my-bucket/base2/"),
		storage.DataReference("s3://my-bucket/base2/_flytecheckpoints"))
	assert.Equal(t, ConstructCheckpointPath(store, ""),
		storage.DataReference(""))
}
