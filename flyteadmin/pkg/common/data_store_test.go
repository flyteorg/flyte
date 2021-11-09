package common

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"google.golang.org/grpc/codes"

	commonMocks "github.com/flyteorg/flyteadmin/pkg/common/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
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
