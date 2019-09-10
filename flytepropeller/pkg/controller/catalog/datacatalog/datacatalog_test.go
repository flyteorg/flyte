package datacatalog

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/controller/catalog/datacatalog/mocks"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
}

func createInmemoryStore(t testing.TB) *storage.DataStore {
	cfg := storage.Config{
		Type: storage.TypeMemory,
	}

	d, err := storage.NewDataStore(&cfg, promutils.NewTestScope())
	assert.NoError(t, err)

	return d
}

func newStringLiteral(value string) *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: &core.Primitive{
						Value: &core.Primitive_StringValue{
							StringValue: value,
						},
					},
				},
			},
		},
	}
}

var sampleParameters = &core.LiteralMap{Literals: map[string]*core.Literal{
	"out1": newStringLiteral("output1-stringval"),
}}

var variableMap = &core.VariableMap{
	Variables: map[string]*core.Variable{
		"test": &core.Variable{
			Type: &core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_STRING,
				},
			},
		},
	},
}

var typedInterface = &core.TypedInterface{
	Inputs:  variableMap,
	Outputs: variableMap,
}

var sampleTask = &core.TaskTemplate{
	Id:        &core.Identifier{Project: "project", Domain: "domain", Name: "name"},
	Interface: typedInterface,
	Metadata: &core.TaskMetadata{
		DiscoveryVersion: "1.0.0",
	},
}

var noInputOutputTask = &core.TaskTemplate{
	Id: &core.Identifier{Project: "project", Domain: "domain", Name: "name"},
	Metadata: &core.TaskMetadata{
		DiscoveryVersion: "1.0.0",
	},
	Interface: &core.TypedInterface{},
}

var datasetID = &datacatalog.DatasetID{
	Project: "project",
	Domain:  "domain",
	Name:    "flyte_task-name",
	Version: "1.0.0-ue5g6uuI-ue5g6uuI",
}

func assertGrpcErr(t *testing.T, err error, code codes.Code) {
	assert.Equal(t, code, status.Code(err))
}

func TestCatalog_Get(t *testing.T) {

	ctx := context.Background()
	testFile := storage.DataReference("test-data.pb")

	t.Run("Empty interface returns err", func(t *testing.T) {
		dataStore := createInmemoryStore(t)

		mockClient := &mocks.DataCatalogClient{}
		catalogClient := &CatalogClient{
			client: mockClient,
			store:  dataStore,
		}
		taskWithoutInterface := &core.TaskTemplate{
			Id: &core.Identifier{Project: "project", Domain: "domain", Name: "name"},
			Metadata: &core.TaskMetadata{
				DiscoveryVersion: "1.0.0",
			},
		}
		_, err := catalogClient.Get(ctx, taskWithoutInterface, testFile)
		assert.Error(t, err)
	})

	t.Run("No results, no Dataset", func(t *testing.T) {
		dataStore := createInmemoryStore(t)
		err := dataStore.WriteProtobuf(ctx, testFile, storage.Options{}, newStringLiteral("output"))
		assert.NoError(t, err)

		mockClient := &mocks.DataCatalogClient{}
		catalogClient := &CatalogClient{
			client: mockClient,
			store:  dataStore,
		}
		mockClient.On("GetDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetDatasetRequest) bool {
				assert.EqualValues(t, datasetID, o.Dataset)
				return true
			}),
		).Return(nil, status.Error(codes.NotFound, "test not found"))
		resp, err := catalogClient.Get(ctx, sampleTask, testFile)
		assert.Error(t, err)

		assertGrpcErr(t, err, codes.NotFound)
		assert.Nil(t, resp)
	})

	t.Run("No results, no Artifact", func(t *testing.T) {
		dataStore := createInmemoryStore(t)

		err := dataStore.WriteProtobuf(ctx, testFile, storage.Options{}, sampleParameters)
		assert.NoError(t, err)

		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
			store:  dataStore,
		}

		sampleDataSet := &datacatalog.Dataset{
			Id: datasetID,
		}
		mockClient.On("GetDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetDatasetRequest) bool {
				assert.EqualValues(t, datasetID, o.Dataset)
				return true
			}),
		).Return(&datacatalog.GetDatasetResponse{Dataset: sampleDataSet}, nil, "")

		mockClient.On("GetArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetArtifactRequest) bool {
				return true
			}),
		).Return(nil, status.Error(codes.NotFound, ""))

		outputs, err := discovery.Get(ctx, sampleTask, testFile)
		assert.Nil(t, outputs)
		assert.Error(t, err)
	})

	t.Run("Found w/ tag", func(t *testing.T) {
		dataStore := createInmemoryStore(t)

		err := dataStore.WriteProtobuf(ctx, testFile, storage.Options{}, sampleParameters)
		assert.NoError(t, err)

		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
			store:  dataStore,
		}

		sampleDataSet := &datacatalog.Dataset{
			Id: datasetID,
		}

		mockClient.On("GetDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetDatasetRequest) bool {
				assert.EqualValues(t, datasetID, o.Dataset)
				return true
			}),
		).Return(&datacatalog.GetDatasetResponse{Dataset: sampleDataSet}, nil)

		sampleArtifactData := &datacatalog.ArtifactData{
			Name:  "test",
			Value: newStringLiteral("output1-stringval"),
		}
		sampleArtifact := &datacatalog.Artifact{
			Id:      "test-artifact",
			Dataset: sampleDataSet.Id,
			Data:    []*datacatalog.ArtifactData{sampleArtifactData},
		}
		mockClient.On("GetArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetArtifactRequest) bool {
				assert.EqualValues(t, datasetID, o.Dataset)
				assert.Equal(t, "flyte_cached-BE6CZsMk6N3ExR_4X9EuwBgj2Jh2UwasXK3a_pM9xlY", o.GetTagName())
				return true
			}),
		).Return(&datacatalog.GetArtifactResponse{Artifact: sampleArtifact}, nil)

		resp, err := discovery.Get(ctx, sampleTask, testFile)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Found w/ tag no inputs or outputs", func(t *testing.T) {
		dataStore := createInmemoryStore(t)

		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
			store:  dataStore,
		}

		sampleDataSet := &datacatalog.Dataset{
			Id: &datacatalog.DatasetID{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
				Version: "1.0.0-V-K42BDF-V-K42BDF",
			},
		}

		mockClient.On("GetDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetDatasetRequest) bool {
				assert.EqualValues(t, "1.0.0-V-K42BDF-V-K42BDF", o.Dataset.Version)
				return true
			}),
		).Return(&datacatalog.GetDatasetResponse{Dataset: sampleDataSet}, nil)

		sampleArtifact := &datacatalog.Artifact{
			Id:      "test-artifact",
			Dataset: sampleDataSet.Id,
			Data:    []*datacatalog.ArtifactData{},
		}
		mockClient.On("GetArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetArtifactRequest) bool {
				assert.EqualValues(t, "1.0.0-V-K42BDF-V-K42BDF", o.Dataset.Version)
				assert.Equal(t, "flyte_cached-m4vFNUOHOFEFIiZSyOyid92TkWFFBDha4UOkkBb47XU", o.GetTagName())
				return true
			}),
		).Return(&datacatalog.GetArtifactResponse{Artifact: sampleArtifact}, nil)

		resp, err := discovery.Get(ctx, noInputOutputTask, testFile)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Literals, 0)
	})
}

func TestCatalog_Put(t *testing.T) {
	ctx := context.Background()

	execID := &core.TaskExecutionIdentifier{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "runID",
			},
		},
		TaskId: &core.Identifier{
			Project: "project",
			Domain:  "domain",
			Name:    "taskRunName",
			Version: "taskRunVersion",
		},
	}

	testFile := storage.DataReference("test-data.pb")

	t.Run("Create new cached execution", func(t *testing.T) {
		dataStore := createInmemoryStore(t)

		err := dataStore.WriteProtobuf(ctx, testFile, storage.Options{}, sampleParameters)
		assert.NoError(t, err)

		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
			store:  dataStore,
		}

		mockClient.On("CreateDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateDatasetRequest) bool {
				assert.True(t, proto.Equal(o.Dataset.Id, datasetID))
				return true
			}),
		).Return(&datacatalog.CreateDatasetResponse{}, nil)

		mockClient.On("CreateArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateArtifactRequest) bool {
				_, parseErr := uuid.Parse(o.Artifact.Id)
				assert.NoError(t, parseErr)
				assert.EqualValues(t, 1, len(o.Artifact.Data))
				assert.EqualValues(t, "out1", o.Artifact.Data[0].Name)
				assert.EqualValues(t, newStringLiteral("output1-stringval"), o.Artifact.Data[0].Value)
				return true
			}),
		).Return(&datacatalog.CreateArtifactResponse{}, nil)

		mockClient.On("AddTag",
			ctx,
			mock.MatchedBy(func(o *datacatalog.AddTagRequest) bool {
				assert.EqualValues(t, "flyte_cached-BE6CZsMk6N3ExR_4X9EuwBgj2Jh2UwasXK3a_pM9xlY", o.Tag.Name)
				return true
			}),
		).Return(&datacatalog.AddTagResponse{}, nil)
		err = discovery.Put(ctx, sampleTask, execID, testFile, testFile)
		assert.NoError(t, err)
	})

	t.Run("Create new cached execution with no inputs/outputs", func(t *testing.T) {
		dataStore := createInmemoryStore(t)
		mockClient := &mocks.DataCatalogClient{}
		catalogClient := &CatalogClient{
			client: mockClient,
			store:  dataStore,
		}

		mockClient.On("CreateDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateDatasetRequest) bool {
				assert.Equal(t, "1.0.0-V-K42BDF-V-K42BDF", o.Dataset.Id.Version)
				return true
			}),
		).Return(&datacatalog.CreateDatasetResponse{}, nil)

		mockClient.On("CreateArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateArtifactRequest) bool {
				assert.EqualValues(t, 0, len(o.Artifact.Data))
				return true
			}),
		).Return(&datacatalog.CreateArtifactResponse{}, nil)

		mockClient.On("AddTag",
			ctx,
			mock.MatchedBy(func(o *datacatalog.AddTagRequest) bool {
				assert.EqualValues(t, "flyte_cached-m4vFNUOHOFEFIiZSyOyid92TkWFFBDha4UOkkBb47XU", o.Tag.Name)
				return true
			}),
		).Return(&datacatalog.AddTagResponse{}, nil)
		err := catalogClient.Put(ctx, noInputOutputTask, execID, "", "")
		assert.NoError(t, err)
	})

	t.Run("Create new cached execution with existing dataset", func(t *testing.T) {
		dataStore := createInmemoryStore(t)

		err := dataStore.WriteProtobuf(ctx, testFile, storage.Options{}, sampleParameters)
		assert.NoError(t, err)

		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
			store:  dataStore,
		}

		createDatasetCalled := false
		mockClient.On("CreateDataset",
			ctx,
			mock.MatchedBy(func(_ *datacatalog.CreateDatasetRequest) bool {
				createDatasetCalled = true
				return true
			}),
		).Return(nil, status.Error(codes.AlreadyExists, "test dataset already exists"))

		createArtifactCalled := false
		mockClient.On("CreateArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateArtifactRequest) bool {
				_, parseErr := uuid.Parse(o.Artifact.Id)
				assert.NoError(t, parseErr)
				assert.EqualValues(t, 1, len(o.Artifact.Data))
				assert.EqualValues(t, "out1", o.Artifact.Data[0].Name)
				assert.EqualValues(t, newStringLiteral("output1-stringval"), o.Artifact.Data[0].Value)
				createArtifactCalled = true
				return true
			}),
		).Return(&datacatalog.CreateArtifactResponse{}, nil)

		addTagCalled := false
		mockClient.On("AddTag",
			ctx,
			mock.MatchedBy(func(o *datacatalog.AddTagRequest) bool {
				assert.EqualValues(t, "flyte_cached-BE6CZsMk6N3ExR_4X9EuwBgj2Jh2UwasXK3a_pM9xlY", o.Tag.Name)
				addTagCalled = true
				return true
			}),
		).Return(&datacatalog.AddTagResponse{}, nil)
		err = discovery.Put(ctx, sampleTask, execID, testFile, testFile)
		assert.NoError(t, err)
		assert.True(t, createDatasetCalled)
		assert.True(t, createArtifactCalled)
		assert.True(t, addTagCalled)
	})

}
