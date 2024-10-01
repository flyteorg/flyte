package ioutils

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginsIOMock "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	storageMocks "github.com/flyteorg/flyte/flytestdlib/storage/mocks"
)

type MemoryMetadata struct {
	exists     bool
	size       int64
	etag       string
	contentMD5 string
}

func (m MemoryMetadata) ContentMD5() string {
	return m.contentMD5
}

func (m MemoryMetadata) Size() int64 {
	return m.size
}

func (m MemoryMetadata) Exists() bool {
	return m.exists
}

func (m MemoryMetadata) Etag() string {
	return m.etag
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

		maxPayloadSize := int64(0)
		r := NewRemoteFileOutputReader(
			ctx,
			store,
			opath,
			maxPayloadSize,
		)

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

		maxPayloadSize := int64(0)
		r := NewRemoteFileOutputReader(
			ctx,
			store,
			opath,
			maxPayloadSize,
		)

		ee, err := r.ReadError(ctx)
		assert.NoError(t, err)
		assert.Equal(t, core.ExecutionError_SYSTEM, ee.Kind)
		assert.True(t, ee.IsRecoverable)
	})

	t.Run("multi-user-error", func(t *testing.T) {
		outputPaths := &pluginsIOMock.OutputFilePaths{}
		outputPaths.OnGetErrorPath().Return("s3://errors/error.pb")

		store := &storageMocks.ComposedProtobufStore{}
		store.OnReadProtobufMatch(mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			errorDoc := &core.ErrorDocument{
				Error: &core.ContainerError{
					Code:    "red",
					Message: "hi",
					Kind:    core.ContainerError_NON_RECOVERABLE,
					Origin:  core.ExecutionError_USER,
				},
			}
			errorFilePath := args.Get(1)
			incomingErrorDoc := args.Get(2)
			assert.NotNil(t, incomingErrorDoc)
			casted := incomingErrorDoc.(*core.ErrorDocument)
			casted.Error = errorDoc.Error
			casted.Error.Message = fmt.Sprintf("%s-%s", casted.Error.Message, errorFilePath)
		}).Return(nil)

		store.OnList(ctx, storage.DataReference("s3://errors/error"), 1000, storage.NewCursorAtStart()).Return(
			[]storage.DataReference{"error-0.pb", "error-1.pb", "error-2.pb"}, storage.NewCursorAtEnd(), nil)

		maxPayloadSize := int64(0)
		r := NewRemoteFileOutputReaderWithErrorAggregationStrategy(
			ctx,
			store,
			outputPaths,
			maxPayloadSize,
			k8s.EarliestErrorAggregationStrategy,
		)

		hasError, err := r.IsError(ctx)
		assert.NoError(t, err)
		assert.True(t, hasError)

		executionError, err := r.ReadError(ctx)
		assert.NoError(t, err)
		assert.Equal(t, core.ExecutionError_USER, executionError.Kind)
		assert.Equal(t, "red", executionError.Code)
		assert.Equal(t, "hi-error-2.pb", executionError.Message)
		assert.False(t, executionError.IsRecoverable)
	})
}
