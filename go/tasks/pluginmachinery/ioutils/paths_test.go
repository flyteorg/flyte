package ioutils

import (
	"context"
	"testing"

	"github.com/flyteorg/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
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

func TestGetTaskTemplatePath(t *testing.T) {
	store := storage.URLPathConstructor{}
	tmpl, err := GetTaskTemplatePath(context.TODO(), store, "s3://abc")
	assert.NoError(t, err)
	assert.Equal(t, tmpl, storage.DataReference("s3://abc/task.pb"))
}

func TestGetIndexLookupPath(t *testing.T) {
	store := storage.URLPathConstructor{}
	tmpl, err := GetIndexLookupPath(context.TODO(), store, "s3://abc")
	assert.NoError(t, err)
	assert.Equal(t, tmpl, storage.DataReference("s3://abc/indexlookup.pb"))
}
