package common

import (
	"context"
	commonMocks "github.com/flyteorg/flyte/flyteadmin/pkg/common/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	"github.com/golang/protobuf/proto"
	errs "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/googleapi"
	"google.golang.org/grpc/codes"
	"testing"
	"time"
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

var testExecID = &core.WorkflowExecutionIdentifier{
	Project: "project",
	Domain:  "domain",
	Name:    "name",
}

func TestOffloadLiteralMap(t *testing.T) {
	mockStorage := commonMocks.GetMockStorageClient()
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).WriteProtobufCb = func(ctx context.Context, reference storage.DataReference, opts storage.Options, msg proto.Message) error {
		assert.Equal(t, reference.String(), "s3://bucket/metadata/nested/project/domain/name/key")
		return nil
	}
	client := datastoreClient{
		storagePrefix: []string{"nested"},
		datastore:     mockStorage,
	}

	uri, err := client.OffloadExecutionLiteralMap(context.TODO(), literalMap, testExecID, "key")
	assert.NoError(t, err)
	assert.Equal(t, "s3://bucket/metadata/nested/project/domain/name/key", uri.String())
}

func TestOffloadLiteralMap_ConstructReferenceError(t *testing.T) {
	mockStorage := commonMocks.GetMockStorageClient()
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).ConstructReferenceCb = func(
		ctx context.Context, reference storage.DataReference, nestedKeys ...string) (storage.DataReference, error) {
		return "foo", errors.NewFlyteAdminError(codes.Internal, "foo")
	}

	client := datastoreClient{
		storagePrefix: []string{"nested"},
		datastore:     mockStorage,
	}

	_, err := client.OffloadExecutionLiteralMap(context.TODO(), literalMap, testExecID, "key")
	assert.Equal(t, err.(errors.FlyteAdminError).Code(), codes.Internal)
}

func TestOffloadLiteralMap_StorageFailure(t *testing.T) {
	mockStorage := commonMocks.GetMockStorageClient()
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).WriteProtobufCb = func(ctx context.Context, reference storage.DataReference, opts storage.Options, msg proto.Message) error {
		assert.Equal(t, reference.String(), "s3://bucket/metadata/nested/project/domain/name/key")
		return errors.NewFlyteAdminError(codes.Internal, "foo")
	}

	client := datastoreClient{
		storagePrefix: []string{"nested"},
		datastore:     mockStorage,
	}

	_, err := client.OffloadExecutionLiteralMap(context.TODO(), literalMap, testExecID, "key")
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

	client := datastoreClient{
		storagePrefix: []string{"nested"},
		datastore:     mockStorage,
	}

	_, err := client.offloadWithRetryDelayAndAttempts(context.TODO(), literalMap, time.Millisecond, expectedRetries, "nested", "key")
	assert.EqualValues(t, retries, expectedRetries+1)
	assert.Equal(t, err.(errors.FlyteAdminError).Code(), codes.Internal)
}
