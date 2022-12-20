package util

import (
	"context"
	"testing"

	commonMocks "github.com/flyteorg/flyteadmin/pkg/common/mocks"
	urlMocks "github.com/flyteorg/flyteadmin/pkg/data/mocks"
	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

var testLiteralMap = &core.LiteralMap{
	Literals: map[string]*core.Literal{
		"foo-1": coreutils.MustMakeLiteral("foo-value-1"),
	},
}

const testOutputsURI = "s3://foo/bar/outputs.pb"

func TestShouldFetchData(t *testing.T) {
	t.Run("local config", func(t *testing.T) {
		assert.True(t, shouldFetchData(&interfaces.RemoteDataConfig{
			Scheme:         common.Local,
			MaxSizeInBytes: 100,
		}, admin.UrlBlob{
			Url:   "s3://data",
			Bytes: 200,
		}))
	})
	t.Run("no config", func(t *testing.T) {
		assert.True(t, shouldFetchData(&interfaces.RemoteDataConfig{
			Scheme:         common.None,
			MaxSizeInBytes: 100,
		}, admin.UrlBlob{
			Url:   "s3://data",
			Bytes: 200,
		}))
	})
	t.Run("no config", func(t *testing.T) {
		assert.True(t, shouldFetchData(&interfaces.RemoteDataConfig{
			Scheme: common.None,
		}, admin.UrlBlob{
			Url:   "s3://data",
			Bytes: 200,
		}))
	})
	t.Run("max size under limit", func(t *testing.T) {
		assert.True(t, shouldFetchData(&interfaces.RemoteDataConfig{
			Scheme:         common.AWS,
			MaxSizeInBytes: 1000,
		}, admin.UrlBlob{
			Url:   "s3://data",
			Bytes: 200,
		}))
	})
	t.Run("max size over limit", func(t *testing.T) {
		assert.False(t, shouldFetchData(&interfaces.RemoteDataConfig{
			Scheme:         common.AWS,
			MaxSizeInBytes: 100,
		}, admin.UrlBlob{
			Url:   "s3://data",
			Bytes: 200,
		}))
	})
	t.Run("empty url config", func(t *testing.T) {
		assert.False(t, shouldFetchData(&interfaces.RemoteDataConfig{
			Scheme:         common.AWS,
			MaxSizeInBytes: 100,
		}, admin.UrlBlob{
			Bytes: 200,
		}))
	})
}

func TestShouldFetchOutputData(t *testing.T) {
	t.Run("local config", func(t *testing.T) {
		assert.True(t, shouldFetchOutputData(&interfaces.RemoteDataConfig{
			Scheme:         common.Local,
			MaxSizeInBytes: 100,
		}, admin.UrlBlob{
			Url:   "s3://data",
			Bytes: 200,
		}, "s3://foo/bar.txt"))
	})
	t.Run("max size under limit", func(t *testing.T) {
		assert.True(t, shouldFetchOutputData(&interfaces.RemoteDataConfig{
			Scheme:         common.AWS,
			MaxSizeInBytes: 1000,
		}, admin.UrlBlob{
			Url:   "s3://data",
			Bytes: 200,
		}, "s3://foo/bar.txt"))
	})
	t.Run("output uri empty", func(t *testing.T) {
		assert.False(t, shouldFetchOutputData(&interfaces.RemoteDataConfig{
			Scheme:         common.AWS,
			MaxSizeInBytes: 1000,
		}, admin.UrlBlob{
			Url:   "s3://data",
			Bytes: 200,
		}, ""))
	})
}

func TestGetInputs(t *testing.T) {
	inputsURI := "s3://foo/bar/inputs.pb"

	expectedURLBlob := admin.UrlBlob{
		Url:   "s3://foo/signed/inputs.pb",
		Bytes: 1000,
	}

	mockRemoteURL := urlMocks.NewMockRemoteURL()
	mockRemoteURL.(*urlMocks.MockRemoteURL).GetCallback = func(ctx context.Context, uri string) (admin.UrlBlob, error) {
		assert.Equal(t, inputsURI, uri)
		return expectedURLBlob, nil
	}
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

	t.Run("should sign URL", func(t *testing.T) {
		remoteDataConfig.SignedURL = interfaces.SignedURL{
			Enabled: true,
		}
		fullInputs, inputURLBlob, err := GetInputs(context.TODO(), mockRemoteURL, &remoteDataConfig, mockStorage, inputsURI)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(fullInputs, testLiteralMap))
		assert.True(t, proto.Equal(inputURLBlob, &expectedURLBlob))
	})
	t.Run("should not sign URL", func(t *testing.T) {
		remoteDataConfig.SignedURL = interfaces.SignedURL{
			Enabled: false,
		}
		fullInputs, inputURLBlob, err := GetInputs(context.TODO(), mockRemoteURL, &remoteDataConfig, mockStorage, inputsURI)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(fullInputs, testLiteralMap))
		assert.Empty(t, inputURLBlob)
	})
}

func TestGetOutputs(t *testing.T) {
	expectedURLBlob := admin.UrlBlob{
		Url:   "s3://foo/signed/outputs.pb",
		Bytes: 1000,
	}

	mockRemoteURL := urlMocks.NewMockRemoteURL()
	mockRemoteURL.(*urlMocks.MockRemoteURL).GetCallback = func(ctx context.Context, uri string) (admin.UrlBlob, error) {
		assert.Equal(t, testOutputsURI, uri)
		return expectedURLBlob, nil
	}

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

		fullOutputs, outputURLBlob, err := GetOutputs(context.TODO(), mockRemoteURL, &remoteDataConfig, mockStorage, closure)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(fullOutputs, testLiteralMap))
		assert.True(t, proto.Equal(outputURLBlob, &expectedURLBlob))
	})
	t.Run("offloaded outputs without signed URL", func(t *testing.T) {
		remoteDataConfig.SignedURL = interfaces.SignedURL{
			Enabled: false,
		}

		fullOutputs, outputURLBlob, err := GetOutputs(context.TODO(), mockRemoteURL, &remoteDataConfig, mockStorage, closure)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(fullOutputs, testLiteralMap))
		assert.Empty(t, outputURLBlob)
	})
	t.Run("inline outputs", func(t *testing.T) {
		mockRemoteURL := urlMocks.NewMockRemoteURL()
		mockRemoteURL.(*urlMocks.MockRemoteURL).GetCallback = func(ctx context.Context, uri string) (admin.UrlBlob, error) {
			t.Fatal("Should not fetch a remote URL for outputs stored inline for an execution model")
			return admin.UrlBlob{}, nil
		}
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

		fullOutputs, outputURLBlob, err := GetOutputs(context.TODO(), mockRemoteURL, &remoteDataConfig, mockStorage, closure)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(fullOutputs, testLiteralMap))
		assert.Empty(t, outputURLBlob)
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
