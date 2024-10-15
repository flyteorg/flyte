package ioutils

import (
	"context"
	"fmt"
	"strings"
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
		r, err := NewRemoteFileOutputReader(
			ctx,
			store,
			opath,
			maxPayloadSize,
		)
		assert.NoError(t, err)

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
		r, err := NewRemoteFileOutputReader(
			ctx,
			store,
			opath,
			maxPayloadSize,
		)
		assert.NoError(t, err)

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
			errorFilePath := args.Get(1).(storage.DataReference)
			workerIdx := strings.Split(strings.Split(errorFilePath.String(), "-")[1], ".")[0]
			errorDoc := &core.ErrorDocument{
				Error: &core.ContainerError{
					Code:    "red",
					Message: fmt.Sprintf("hi-%s", workerIdx),
					Kind:    core.ContainerError_NON_RECOVERABLE,
					Origin:  core.ExecutionError_USER,
					Worker:  fmt.Sprintf("worker-%s", workerIdx),
				},
			}
			incomingErrorDoc := args.Get(2)
			assert.NotNil(t, incomingErrorDoc)
			casted := incomingErrorDoc.(*core.ErrorDocument)
			casted.Error = errorDoc.Error
		}).Return(nil)

		store.OnList(ctx, storage.DataReference("s3://errors/error"), 1000, storage.NewCursorAtStart()).Return(
			[]storage.DataReference{"error-0.pb", "error-1.pb", "error-2.pb"}, storage.NewCursorAtEnd(), nil)

		store.OnHead(ctx, storage.DataReference("error-0.pb")).Return(MemoryMetadata{
			exists: true,
		}, nil)

		store.OnHead(ctx, storage.DataReference("error-1.pb")).Return(MemoryMetadata{
			exists: true,
		}, nil)

		store.OnHead(ctx, storage.DataReference("error-2.pb")).Return(MemoryMetadata{
			exists: true,
		}, nil)

		maxPayloadSize := int64(0)
		r, err := NewRemoteFileOutputReaderWithErrorAggregationStrategy(
			ctx,
			store,
			outputPaths,
			maxPayloadSize,
			k8s.EarliestErrorAggregationStrategy,
		)
		assert.NoError(t, err)

		hasError, err := r.IsError(ctx)
		assert.NoError(t, err)
		assert.True(t, hasError)

		executionError, err := r.ReadError(ctx)
		assert.NoError(t, err)
		assert.Equal(t, core.ExecutionError_USER, executionError.Kind)
		assert.Equal(t, "red", executionError.Code)
		assert.Equal(t, "hi-2", executionError.Message)
		assert.Equal(t, "worker-2", executionError.Worker)
		assert.False(t, executionError.IsRecoverable)
	})
}
