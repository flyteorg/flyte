package ioutils

import (
	"context"
	"testing"

	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
)

func TestRemoteFileOutputWriter(t *testing.T) {
	ctx := context.TODO()
	memStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.Nil(t, err)

	outputPrefix := storage.DataReference("output")
	rawOutputPrefix := storage.DataReference("sandbox")
	previousCheckpointPath := storage.DataReference("checkpoint")

	checkpointPath := NewCheckpointRemoteFilePaths(
		ctx,
		memStore,
		outputPrefix,
		NewRawOutputPaths(ctx, rawOutputPrefix),
		previousCheckpointPath,
	)

	t.Run("Test NewCheckpointRemoteFilePaths", func(t *testing.T) {
		assert.Equal(t, previousCheckpointPath, checkpointPath.GetPreviousCheckpointsPrefix())
		assert.Equal(t, outputPrefix, checkpointPath.GetOutputPrefixPath())

		assert.Equal(t, constructPath(memStore, rawOutputPrefix, CheckpointPrefix), checkpointPath.GetCheckpointPrefix())
		assert.Equal(t, constructPath(memStore, outputPrefix, OutputsSuffix), checkpointPath.GetOutputPath())
		assert.Equal(t, constructPath(memStore, outputPrefix, deckSuffix), checkpointPath.GetDeckPath())
		assert.Equal(t, constructPath(memStore, outputPrefix, ErrorsSuffix), checkpointPath.GetErrorPath())
		assert.Equal(t, constructPath(memStore, outputPrefix, FuturesSuffix), checkpointPath.GetFuturesPath())
	})

	t.Run("Test NewRemoteFileOutputWriter", func(t *testing.T) {
		p := NewRemoteFileOutputWriter(ctx, memStore, checkpointPath)

		assert.Equal(t, constructPath(memStore, rawOutputPrefix, CheckpointPrefix), p.GetCheckpointPrefix())
		assert.Equal(t, constructPath(memStore, outputPrefix, OutputsSuffix), p.GetOutputPath())
		assert.Equal(t, constructPath(memStore, outputPrefix, deckSuffix), p.GetDeckPath())
		assert.Equal(t, constructPath(memStore, outputPrefix, ErrorsSuffix), p.GetErrorPath())
	})
}
