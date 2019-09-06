package catalog

import (
	"context"
	"testing"

	"github.com/lyft/flyteidl/clients/go/datacatalog/mocks"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	labeled.SetMetricKeys(contextutils.TaskIDKey)
}

func createInmemoryDataStore(t testing.TB, scope promutils.Scope) *storage.DataStore {
	cfg := storage.Config{
		Type: storage.TypeMemory,
	}
	d, err := storage.NewDataStore(&cfg, scope)
	assert.NoError(t, err)
	return d
}

func newIntegerPrimitive(value int64) *core.Primitive {
	return &core.Primitive{Value: &core.Primitive_Integer{Integer: value}}
}

func newScalarInteger(value int64) *core.Scalar {
	return &core.Scalar{
		Value: &core.Scalar_Primitive{
			Primitive: newIntegerPrimitive(value),
		},
	}
}

func newIntegerLiteral(value int64) *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: newScalarInteger(value),
		},
	}
}

func TestTransformToInputParameters(t *testing.T) {
	paramsMap := make(map[string]*core.Literal)
	paramsMap["out1"] = newIntegerLiteral(200)

	params, err := TransformToInputParameters(context.Background(), &core.LiteralMap{
		Literals: paramsMap,
	})
	assert.Nil(t, err)
	assert.Equal(t, "out1", params[0].Name)
	assert.Equal(t, "c6i2T7NODjwnlxmXKRCNDk/AN4pZpRGGFX49kT6DT/c=", params[0].Value)
}

func TestTransformToOutputParameters(t *testing.T) {
	paramsMap := make(map[string]*core.Literal)
	paramsMap["out1"] = newIntegerLiteral(100)

	params, err := TransformToOutputParameters(context.Background(), &core.LiteralMap{
		Literals: paramsMap,
	})
	assert.Nil(t, err)
	assert.Equal(t, "out1", params[0].Name)
	assert.Equal(t, "CgQKAghk", params[0].Value)
}

func TestTransformFromParameters(t *testing.T) {
	params := []*datacatalog.Parameter{
		{Name: "out1", Value: "CgQKAghk"},
	}
	literalMap, err := TransformFromParameters(params)
	assert.Nil(t, err)

	val, exists := literalMap.Literals["out1"]
	assert.True(t, exists)
	assert.Equal(t, int64(100), val.GetScalar().GetPrimitive().GetInteger())
}

func TestLegacyDiscovery_Get(t *testing.T) {
	ctx := context.Background()

	paramMap := &core.LiteralMap{Literals: map[string]*core.Literal{
		"out1": newIntegerLiteral(100),
	}}
	task := &core.TaskTemplate{
		Id: &core.Identifier{Project: "project", Domain: "domain", Name: "name"},
		Metadata: &core.TaskMetadata{
			DiscoveryVersion: "0.0.1",
		},
		Interface: &core.TypedInterface{
			Inputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"out1": &core.Variable{
						Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
					},
				},
			},
			Outputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"out1": &core.Variable{
						Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
					},
				},
			},
		},
	}

	inputPath := storage.DataReference("test-data/inputs.pb")

	t.Run("notfound", func(t *testing.T) {
		mockClient := &mocks.ArtifactsClient{}
		store := createInmemoryDataStore(t, promutils.NewScope("get_test_notfound"))

		err := store.WriteProtobuf(ctx, inputPath, storage.Options{}, paramMap)
		assert.NoError(t, err)

		discovery := LegacyDiscovery{client: mockClient, store: store}
		mockClient.On("Get",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetRequest) bool {
				assert.Equal(t, o.GetArtifactId().Name, "project:domain:name")
				params, err := TransformToInputParameters(context.Background(), paramMap)
				assert.NoError(t, err)
				assert.Equal(t, o.GetArtifactId().Inputs, params)
				return true
			}),
			mock.MatchedBy(func(opts grpc.CallOption) bool { return true }),
			mock.MatchedBy(func(opts grpc.CallOption) bool { return true }),
		).Return(nil, status.Error(codes.NotFound, ""))
		resp, err := discovery.Get(ctx, task, inputPath)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("found", func(t *testing.T) {
		mockClient := &mocks.ArtifactsClient{}
		store := createInmemoryDataStore(t, promutils.NewScope("get_test_found"))

		err := store.WriteProtobuf(ctx, inputPath, storage.Options{}, paramMap)
		assert.NoError(t, err)

		discovery := LegacyDiscovery{client: mockClient, store: store}
		outputs, err := TransformToOutputParameters(context.Background(), paramMap)
		assert.NoError(t, err)
		response := &datacatalog.GetResponse{
			Artifact: &datacatalog.Artifact{
				Outputs: outputs,
			},
		}
		mockClient.On("Get",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetRequest) bool {
				assert.Equal(t, o.GetArtifactId().Name, "project:domain:name")
				assert.Equal(t, o.GetArtifactId().Version, "0.0.1")
				params, err := TransformToInputParameters(context.Background(), paramMap)
				assert.NoError(t, err)
				assert.Equal(t, o.GetArtifactId().Inputs, params)
				return true
			}),
			mock.MatchedBy(func(opts grpc.CallOption) bool { return true }),
			mock.MatchedBy(func(opts grpc.CallOption) bool { return true }),
		).Return(response, nil)
		resp, err := discovery.Get(ctx, task, inputPath)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		val, exists := resp.Literals["out1"]
		assert.True(t, exists)
		assert.Equal(t, int64(100), val.GetScalar().GetPrimitive().GetInteger())
	})
}

func TestLegacyDiscovery_Put(t *testing.T) {
	ctx := context.Background()

	inputPath := storage.DataReference("test-data/inputs.pb")
	outputPath := storage.DataReference("test-data/ouputs.pb")

	paramMap := &core.LiteralMap{Literals: map[string]*core.Literal{
		"out1": newIntegerLiteral(100),
	}}
	task := &core.TaskTemplate{
		Id: &core.Identifier{Project: "project", Domain: "domain", Name: "name"},
		Metadata: &core.TaskMetadata{
			DiscoveryVersion: "0.0.1",
		},
		Interface: &core.TypedInterface{
			Inputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"out1": &core.Variable{
						Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
					},
				},
			},
			Outputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"out1": &core.Variable{
						Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
					},
				},
			},
		},
	}

	execID := &core.TaskExecutionIdentifier{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "runID",
			},
		},
	}

	t.Run("failed", func(t *testing.T) {
		mockClient := &mocks.ArtifactsClient{}
		store := createInmemoryDataStore(t, promutils.NewScope("put_test_failed"))
		discovery := LegacyDiscovery{client: mockClient, store: store}
		mockClient.On("Create",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateRequest) bool {
				assert.Equal(t, o.GetRef().Name, "project:domain:name")
				assert.Equal(t, o.GetReferenceId(), "project:domain:runID")
				inputs, err := TransformToInputParameters(context.Background(), paramMap)
				assert.NoError(t, err)
				outputs, err := TransformToOutputParameters(context.Background(), paramMap)
				assert.NoError(t, err)
				assert.Equal(t, o.GetRef().Inputs, inputs)
				assert.Equal(t, o.GetOutputs(), outputs)

				return true
			}),
			mock.MatchedBy(func(opts grpc.CallOption) bool { return true }),
			mock.MatchedBy(func(opts grpc.CallOption) bool { return true }),
		).Return(nil, status.Error(codes.AlreadyExists, ""))

		err := store.WriteProtobuf(ctx, inputPath, storage.Options{}, paramMap)
		assert.NoError(t, err)
		err = store.WriteProtobuf(ctx, outputPath, storage.Options{}, paramMap)
		assert.NoError(t, err)

		err = discovery.Put(ctx, task, execID, inputPath, outputPath)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		store := createInmemoryDataStore(t, promutils.NewScope("put_test_success"))
		mockClient := &mocks.ArtifactsClient{}
		discovery := LegacyDiscovery{client: mockClient, store: store}
		mockClient.On("Create",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateRequest) bool {
				assert.Equal(t, o.GetRef().Name, "project:domain:name")
				assert.Equal(t, o.GetRef().Version, "0.0.1")
				inputs, err := TransformToInputParameters(context.Background(), paramMap)
				assert.NoError(t, err)
				outputs, err := TransformToOutputParameters(context.Background(), paramMap)
				assert.NoError(t, err)
				assert.Equal(t, o.GetRef().Inputs, inputs)
				assert.Equal(t, o.GetOutputs(), outputs)

				return true
			}),
			mock.MatchedBy(func(opts grpc.CallOption) bool { return true }),
			mock.MatchedBy(func(opts grpc.CallOption) bool { return true }),
		).Return(nil, nil)

		err := store.WriteProtobuf(ctx, inputPath, storage.Options{}, paramMap)
		assert.NoError(t, err)
		err = store.WriteProtobuf(ctx, outputPath, storage.Options{}, paramMap)
		assert.NoError(t, err)

		err = discovery.Put(ctx, task, execID, inputPath, outputPath)
		assert.NoError(t, err)
	})
}
