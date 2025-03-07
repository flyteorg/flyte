package ioutils

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	regErrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/runtime/protoiface"
	"google.golang.org/protobuf/types/known/timestamppb"

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

func TestExistsTooBig(t *testing.T) {
	ctx := context.TODO()
	opath := &pluginsIOMock.OutputFilePaths{}
	opath.EXPECT().GetErrorPath().Return("")
	deckPath := "some.file"
	opath.EXPECT().GetOutputPath().Return(storage.DataReference(deckPath))

	t.Run("too large", func(t *testing.T) {
		store := &storageMocks.ComposedProtobufStore{}
		store.EXPECT().Head(ctx, storage.DataReference("some.file")).Return(MemoryMetadata{
			exists: true,
			size:   2,
		}, nil)

		r := RemoteFileOutputReader{
			outPath:        opath,
			store:          store,
			maxPayloadSize: 1,
		}

		_, err := r.Exists(ctx)
		assert.Error(t, err)
		assert.True(t, regErrors.Is(err, ErrRemoteFileExceedsMaxSize))
	})
}

func TestReadOrigin(t *testing.T) {
	ctx := context.TODO()

	opath := &pluginsIOMock.OutputFilePaths{}
	opath.EXPECT().GetErrorPath().Return("")
	deckPath := "deck.html"
	opath.EXPECT().GetDeckPath().Return(storage.DataReference(deckPath))

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
		store.EXPECT().ReadProtobuf(mock.Anything, mock.Anything, mock.Anything).Run(func(ctx context.Context, reference storage.DataReference, incomingErrorDoc protoiface.MessageV1) {
			assert.NotNil(t, incomingErrorDoc)
			casted := incomingErrorDoc.(*core.ErrorDocument)
			casted.Error = errorDoc.GetError()
		}).Return(nil)

		store.EXPECT().Head(ctx, storage.DataReference("deck.html")).Return(MemoryMetadata{
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
		store.EXPECT().ReadProtobuf(mock.Anything, mock.Anything, mock.Anything).Run(func(ctx context.Context, reference storage.DataReference, incomingErrorDoc protoiface.MessageV1) {
			assert.NotNil(t, incomingErrorDoc)
			casted := incomingErrorDoc.(*core.ErrorDocument)
			casted.Error = errorDoc.GetError()
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
		outputPaths.EXPECT().GetErrorPath().Return("s3://errors/error.pb")

		store := &storageMocks.ComposedProtobufStore{}
		store.EXPECT().ReadProtobuf(mock.Anything, mock.Anything, mock.Anything).Run(func(ctx context.Context, errorFilePath storage.DataReference, incomingErrorDoc protoiface.MessageV1) {
			workerIdx, err := strconv.Atoi(strings.Split(strings.Split(errorFilePath.String(), "-")[1], ".")[0])
			assert.NoError(t, err)
			errorDoc := &core.ErrorDocument{
				Error: &core.ContainerError{
					Code:      "red",
					Message:   fmt.Sprintf("hi-%d", workerIdx),
					Kind:      core.ContainerError_NON_RECOVERABLE,
					Origin:    core.ExecutionError_USER,
					Worker:    fmt.Sprintf("worker-%d", workerIdx),
					Timestamp: timestamppb.New(time.Unix(int64(100-workerIdx%2), 0)),
				},
			}
			assert.NotNil(t, incomingErrorDoc)
			casted := incomingErrorDoc.(*core.ErrorDocument)
			casted.Error = errorDoc.GetError()
		}).Return(nil)

		store.EXPECT().List(ctx, storage.DataReference("s3://errors/error"), 1000, storage.NewCursorAtStart()).Return(
			[]storage.DataReference{"s3://errors/error-0.pb", "s3://errors/error-1.pb", "s3://errors/error-2.pb"}, storage.NewCursorAtEnd(), nil)

		store.EXPECT().Head(ctx, storage.DataReference("s3://errors/error-0.pb")).Return(MemoryMetadata{
			exists: true,
		}, nil)

		store.EXPECT().Head(ctx, storage.DataReference("s3://errors/error-1.pb")).Return(MemoryMetadata{
			exists: true,
		}, nil)

		store.EXPECT().Head(ctx, storage.DataReference("s3://errors/error-2.pb")).Return(MemoryMetadata{
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
		assert.Equal(t, "hi-1", executionError.Message)
		assert.Equal(t, "worker-1", executionError.Worker)
		assert.Equal(t, timestamppb.New(time.Unix(99, 0)), executionError.Timestamp)
		assert.False(t, executionError.IsRecoverable)
	})

	t.Run("multi-user-error-backward-compat", func(t *testing.T) {
		outputPaths := &pluginsIOMock.OutputFilePaths{}
		outputPaths.EXPECT().GetErrorPath().Return("s3://errors/error.pb")

		store := &storageMocks.ComposedProtobufStore{}
		store.EXPECT().ReadProtobuf(mock.Anything, mock.Anything, mock.Anything).Run(func(ctx context.Context, reference storage.DataReference, incomingErrorDoc protoiface.MessageV1) {
			errorDoc := &core.ErrorDocument{
				Error: &core.ContainerError{
					Code:    "red",
					Message: "hi",
					Kind:    core.ContainerError_NON_RECOVERABLE,
					Origin:  core.ExecutionError_USER,
				},
			}
			assert.NotNil(t, incomingErrorDoc)
			casted := incomingErrorDoc.(*core.ErrorDocument)
			casted.Error = errorDoc.GetError()
		}).Return(nil)

		store.EXPECT().List(ctx, storage.DataReference("s3://errors/error"), 1000, storage.NewCursorAtStart()).Return(
			[]storage.DataReference{"s3://errors/error.pb"}, storage.NewCursorAtEnd(), nil)

		store.EXPECT().Head(ctx, storage.DataReference("s3://errors/error.pb")).Return(MemoryMetadata{
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
		assert.Equal(t, "hi", executionError.Message)
		assert.Equal(t, "", executionError.Worker)
		assert.Nil(t, executionError.Timestamp)
		assert.False(t, executionError.IsRecoverable)
	})
}
