package utils

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func TestFailingRawStore(t *testing.T) {
	ctx := context.TODO()
	f := FailingRawStore{}
	_, err := f.Head(ctx, "")
	assert.Error(t, err)

	c := f.GetBaseContainerFQN(ctx)
	assert.Equal(t, storage.DataReference(""), c)

	_, err = f.ReadRaw(ctx, "")
	assert.Error(t, err)

	assert.Error(t, f.WriteRaw(ctx, "", 0, storage.Options{}, bytes.NewReader(nil)))

	assert.Error(t, f.CopyRaw(ctx, "", "", storage.Options{}))

	assert.Error(t, f.Delete(ctx, ""))
}
