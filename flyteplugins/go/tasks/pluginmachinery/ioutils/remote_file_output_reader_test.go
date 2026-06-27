package ioutils

import (
	"context"
	"testing"

	pluginsIOMock "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	storageMocks "github.com/flyteorg/flyte/v2/flytestdlib/storage/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/runtime/protoiface"
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
		store.EXPECT().ReadProtobuf(mock.Anything, mock.Anything, mock.Anything).Run(func(ctx context.Context, ref storage.DataReference, msg protoiface.MessageV1) {
			assert.NotNil(t, msg)
			casted := msg.(*core.ErrorDocument)
			casted.Error = errorDoc.Error
		}).Return(nil)

		store.EXPECT().Head(ctx, storage.DataReference("deck.html")).Return(MemoryMetadata{
			exists: true,
		}, nil)

		r := RemoteFileOutputReader{
			OutPath:        opath,
			store:          store,
			maxPayloadSize: 0,
		}

		ee, err := r.ReadError(ctx)
		assert.NoError(t, err)
		assert.Equal(t, core.ExecutionError_USER, ee.Kind)
		assert.False(t, ee.IsRecoverable)
		// Proto-level Recoverability mirrors the container's kind so the bit
		// travels on the wire to downstream services. Without this, all errors
		// surface as the proto3 zero (NON_RECOVERABLE) regardless of intent.
		assert.Equal(t, core.ContainerError_NON_RECOVERABLE, ee.GetRecoverability())
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
		store.EXPECT().ReadProtobuf(mock.Anything, mock.Anything, mock.Anything).Run(func(ctx context.Context, ref storage.DataReference, msg protoiface.MessageV1) {
			assert.NotNil(t, msg)
			casted := msg.(*core.ErrorDocument)
			casted.Error = errorDoc.Error
		}).Return(nil)

		r := RemoteFileOutputReader{
			OutPath:        opath,
			store:          store,
			maxPayloadSize: 0,
		}

		ee, err := r.ReadError(ctx)
		assert.NoError(t, err)
		assert.Equal(t, core.ExecutionError_SYSTEM, ee.Kind)
		assert.True(t, ee.IsRecoverable)
		assert.Equal(t, core.ContainerError_RECOVERABLE, ee.GetRecoverability())
	})
}
