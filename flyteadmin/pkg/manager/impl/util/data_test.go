package util

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	commonMocks "github.com/flyteorg/flyte/flyteadmin/pkg/common/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var testLiteralMap = &core.LiteralMap{
	Literals: map[string]*core.Literal{
		"foo-1": coreutils.MustMakeLiteral("foo-value-1"),
	},
}

const testOutputsURI = "s3://foo/bar/outputs.pb"

func TestGetInputs(t *testing.T) {
	inputsURI := "s3://foo/bar/inputs.pb"

	remoteDataConfig := interfaces.RemoteDataConfig{
		MaxSizeInBytes: 2000,
	}

	mockStorage := commonMocks.GetMockStorageClient()
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).ReadProtobufCb = func(
		ctx context.Context, reference storage.DataReference, msg proto.Message) error {
		assert.Equal(t, inputsURI, reference.String())
		marshalled, _ := proto.Marshal(testLiteralMap)
		_ = proto.Unmarshal(marshalled, msg)
		return nil
	}

	remoteDataConfig.SignedURL = interfaces.SignedURL{
		Enabled: true,
	}
	fullInputs := GetInputs(context.TODO(), mockStorage, inputsURI)
	assert.True(t, proto.Equal(fullInputs, testLiteralMap))
}

func TestGetOutputs(t *testing.T) {
	remoteDataConfig := interfaces.RemoteDataConfig{
		MaxSizeInBytes: 2000,
	}
	mockStorage := commonMocks.GetMockStorageClient()
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).ReadProtobufCb = func(
		ctx context.Context, reference storage.DataReference, msg proto.Message) error {
		assert.Equal(t, testOutputsURI, reference.String())
		marshalled, _ := proto.Marshal(testLiteralMap)
		_ = proto.Unmarshal(marshalled, msg)
		return nil
	}
	closure := &admin.NodeExecutionClosure{
		OutputResult: &admin.NodeExecutionClosure_OutputUri{
			OutputUri: testOutputsURI,
		},
	}
	t.Run("offloaded outputs with signed URL", func(t *testing.T) {
		remoteDataConfig.SignedURL = interfaces.SignedURL{
			Enabled: true,
		}

		fullOutputs := GetOutputs(context.TODO(), &remoteDataConfig, mockStorage, closure)
		assert.True(t, proto.Equal(fullOutputs, testLiteralMap))
	})
	t.Run("offloaded outputs without signed URL", func(t *testing.T) {
		remoteDataConfig.SignedURL = interfaces.SignedURL{
			Enabled: false,
		}

		fullOutputs := GetOutputs(context.TODO(), &remoteDataConfig, mockStorage, closure)
		assert.True(t, proto.Equal(fullOutputs, testLiteralMap))
	})
	t.Run("inline outputs", func(t *testing.T) {
		remoteDataConfig := interfaces.RemoteDataConfig{}
		remoteDataConfig.MaxSizeInBytes = 2000

		mockStorage := commonMocks.GetMockStorageClient()
		mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).ReadProtobufCb = func(
			ctx context.Context, reference storage.DataReference, msg proto.Message) error {
			t.Fatal("Should not fetch when outputs stored inline for an execution model")
			return nil
		}
		closure := &admin.NodeExecutionClosure{
			OutputResult: &admin.NodeExecutionClosure_OutputData{
				OutputData: testLiteralMap,
			},
		}

		fullOutputs := GetOutputs(context.TODO(), &remoteDataConfig, mockStorage, closure)
		assert.True(t, proto.Equal(fullOutputs, testLiteralMap))
	})
}

func TestWorkflowExecutionClosure(t *testing.T) {
	t.Run("outputs offloaded", func(t *testing.T) {
		workflowExecutionClosure := admin.ExecutionClosure{
			OutputResult: &admin.ExecutionClosure_Outputs{
				Outputs: &admin.LiteralMapBlob{
					Data: &admin.LiteralMapBlob_Uri{
						Uri: testOutputsURI,
					},
				},
			},
		}
		closureImpl := ToExecutionClosureInterface(&workflowExecutionClosure)
		assert.Equal(t, testOutputsURI, closureImpl.GetOutputUri())
	})
	t.Run("outputs inline", func(t *testing.T) {
		workflowExecutionClosure := admin.ExecutionClosure{
			OutputResult: &admin.ExecutionClosure_OutputData{
				OutputData: testLiteralMap,
			},
		}
		closureImpl := ToExecutionClosureInterface(&workflowExecutionClosure)
		assert.Empty(t, closureImpl.GetOutputUri())
		assert.True(t, proto.Equal(testLiteralMap, closureImpl.GetOutputData()))
	})
	t.Run("outputs inline - historical/deprecated format", func(t *testing.T) {
		workflowExecutionClosure := admin.ExecutionClosure{
			OutputResult: &admin.ExecutionClosure_Outputs{
				Outputs: &admin.LiteralMapBlob{
					Data: &admin.LiteralMapBlob_Values{
						Values: testLiteralMap,
					},
				},
			},
		}
		closureImpl := ToExecutionClosureInterface(&workflowExecutionClosure)
		assert.Empty(t, closureImpl.GetOutputUri())
		assert.True(t, proto.Equal(testLiteralMap, closureImpl.GetOutputData()))
	})
}
