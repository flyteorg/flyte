package common

import (
	"context"
	"testing"
	"time"

	errs "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/googleapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"

	commonMocks "github.com/flyteorg/flyte/flyteadmin/pkg/common/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var literalMap = &core.LiteralMap{
	Literals: map[string]*core.Literal{
		"foo": {
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Primitive{
						Primitive: &core.Primitive{
							Value: &core.Primitive_Integer{
								Integer: 4,
							},
						},
					},
				},
			},
		},
	},
}

func TestOffloadLiteralMap(t *testing.T) {
	mockStorage := commonMocks.GetMockStorageClient()
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).WriteProtobufCb = func(ctx context.Context, reference storage.DataReference, opts storage.Options, msg proto.Message) error {
		assert.Equal(t, reference.String(), "s3://bucket/metadata/nested/key")
		return nil
	}

	uri, err := OffloadLiteralMap(context.TODO(), mockStorage, literalMap, "nested", "key")
	assert.NoError(t, err)
	assert.Equal(t, "s3://bucket/metadata/nested/key", uri.String())
}

func TestOffloadLiteralMap_ConstructReferenceError(t *testing.T) {
	mockStorage := commonMocks.GetMockStorageClient()
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).ConstructReferenceCb = func(
		ctx context.Context, reference storage.DataReference, nestedKeys ...string) (storage.DataReference, error) {
		return "foo", errors.NewFlyteAdminError(codes.Internal, "foo")
	}
	_, err := OffloadLiteralMap(context.TODO(), mockStorage, literalMap, "nested", "key")
	assert.Equal(t, err.(errors.FlyteAdminError).Code(), codes.Internal)
}

func TestOffloadLiteralMap_StorageFailure(t *testing.T) {
	mockStorage := commonMocks.GetMockStorageClient()
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).WriteProtobufCb = func(ctx context.Context, reference storage.DataReference, opts storage.Options, msg proto.Message) error {
		assert.Equal(t, reference.String(), "s3://bucket/metadata/nested/key")
		return errors.NewFlyteAdminError(codes.Internal, "foo")
	}
	_, err := OffloadLiteralMap(context.TODO(), mockStorage, literalMap, "nested", "key")
	assert.Equal(t, err.(errors.FlyteAdminError).Code(), codes.Internal)
}

func TestOffloadLiteralMap_RetryOn409(t *testing.T) {
	mockStorage := commonMocks.GetMockStorageClient()
	retries := 0
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).WriteProtobufCb = func(ctx context.Context, reference storage.DataReference, opts storage.Options, msg proto.Message) error {
		retries++
		assert.Equal(t, reference.String(), "s3://bucket/metadata/nested/key")
		return errs.Wrapf(&googleapi.Error{Code: 409}, "Failed to write data [%vb] to path [%v].", 10, "size")
	}
	expectedRetries := 2
	_, err := OffloadLiteralMapWithRetryDelayAndAttempts(context.TODO(), mockStorage, literalMap, time.Millisecond, expectedRetries, "nested", "key")
	assert.EqualValues(t, retries, expectedRetries+1)
	assert.Equal(t, err.(errors.FlyteAdminError).Code(), codes.Internal)
}
