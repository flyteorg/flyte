package ioutils

import (
	"context"
	"testing"

	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	pluginsIOMock "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	storageMocks "github.com/flyteorg/flytestdlib/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MemoryMetadata struct {
	exists bool
	size   int64
}

func (m MemoryMetadata) Size() int64 {
	return m.size
}

func (m MemoryMetadata) Exists() bool {
	return m.exists
}

func TestReadOrigin(t *testing.T) {
	ctx := context.TODO()

	opath := &pluginsIOMock.OutputFilePaths{}
	opath.OnGetErrorPath().Return("")
	deckPath := "deck.html"
	opath.OnGetDeckPath().Return(storage.DataReference(deckPath))

	t.Run("user", func(t *testing.T) {
		errorDoc := &core.ErrorDocument{
			Error: &core.ContainerError{
				Code:    "red",
				Message: "hi",
				Kind:    core.ContainerError_NON_RECOVERABLE,
				Origin:  core.ExecutionError_USER,
			},
		}
		store := &storageMocks.ComposedProtobufStore{}
		store.OnReadProtobufMatch(mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			incomingErrorDoc := args.Get(2)
			assert.NotNil(t, incomingErrorDoc)
			casted := incomingErrorDoc.(*core.ErrorDocument)
			casted.Error = errorDoc.Error
		}).Return(nil)

		store.OnHead(ctx, storage.DataReference("deck.html")).Return(MemoryMetadata{
			exists: true,
		}, nil)

		r := RemoteFileOutputReader{
			outPath:        opath,
			store:          store,
			maxPayloadSize: 0,
		}

		ee, err := r.ReadError(ctx)
		assert.NoError(t, err)
		assert.Equal(t, core.ExecutionError_USER, ee.Kind)
		assert.False(t, ee.IsRecoverable)
		exists, err := r.DeckExists(ctx)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("system", func(t *testing.T) {
		errorDoc := &core.ErrorDocument{
			Error: &core.ContainerError{
				Code:    "red",
				Message: "hi",
				Kind:    core.ContainerError_RECOVERABLE,
				Origin:  core.ExecutionError_SYSTEM,
			},
		}
		store := &storageMocks.ComposedProtobufStore{}
		store.OnReadProtobufMatch(mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			incomingErrorDoc := args.Get(2)
			assert.NotNil(t, incomingErrorDoc)
			casted := incomingErrorDoc.(*core.ErrorDocument)
			casted.Error = errorDoc.Error
		}).Return(nil)

		r := RemoteFileOutputReader{
			outPath:        opath,
			store:          store,
			maxPayloadSize: 0,
		}

		ee, err := r.ReadError(ctx)
		assert.NoError(t, err)
		assert.Equal(t, core.ExecutionError_SYSTEM, ee.Kind)
		assert.True(t, ee.IsRecoverable)
	})
}
