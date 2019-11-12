package datacatalog

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/catalog"
	mocks2 "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/pkg/errors"

	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lyft/flytepropeller/pkg/controller/nodes/task/catalog/datacatalog/mocks"
)

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
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
		"test": {
			Type: &core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_STRING,
				},
			},
		},
	},
}

var typedInterface = core.TypedInterface{
	Inputs:  variableMap,
	Outputs: variableMap,
}

var sampleKey = catalog.Key{
	Identifier:     core.Identifier{Project: "project", Domain: "domain", Name: "name"},
	TypedInterface: typedInterface,
	CacheVersion:   "1.0.0",
}

var noInputOutputKey = catalog.Key{
	Identifier:   core.Identifier{Project: "project", Domain: "domain", Name: "name"},
	CacheVersion: "1.0.0",
}

var datasetID = &datacatalog.DatasetID{
	Project: "project",
	Domain:  "domain",
	Name:    "flyte_task-name",
	Version: "1.0.0-ue5g6uuI-ue5g6uuI",
}

func assertGrpcErr(t *testing.T, err error, code codes.Code) {
	assert.Equal(t, code, status.Code(errors.Cause(err)), "Got err: %s", err)
}

func TestCatalog_Get(t *testing.T) {

	ctx := context.Background()

	t.Run("No results, no Dataset", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(newStringLiteral("output"), nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		catalogClient := &CatalogClient{
			client: mockClient,
		}
		mockClient.On("GetDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetDatasetRequest) bool {
				assert.EqualValues(t, datasetID, o.Dataset)
				return true
			}),
		).Return(nil, status.Error(codes.NotFound, "test not found"))
		newKey := sampleKey
		newKey.InputReader = ir
		resp, err := catalogClient.Get(ctx, newKey)
		assert.Error(t, err)

		assertGrpcErr(t, err, codes.NotFound)
		assert.Nil(t, resp)
	})

	t.Run("No results, no Artifact", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		catalogClient := &CatalogClient{
			client: mockClient,
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

		newKey := sampleKey
		newKey.InputReader = ir
		outputs, err := catalogClient.Get(ctx, newKey)
		assert.Nil(t, outputs)
		assert.Error(t, err)
	})

	t.Run("Found w/ tag", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		catalogClient := &CatalogClient{
			client: mockClient,
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

		newKey := sampleKey
		newKey.InputReader = ir
		resp, err := catalogClient.Get(ctx, newKey)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Found w/ tag no inputs or outputs", func(t *testing.T) {
		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
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

		resp, err := discovery.Get(ctx, noInputOutputKey)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		v, e, err := resp.Read(ctx)
		assert.NoError(t, err)
		assert.Nil(t, e)
		assert.Len(t, v.Literals, 0)
	})
}

func TestCatalog_Put(t *testing.T) {
	ctx := context.Background()

	t.Run("Create new cached execution", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
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
		newKey := sampleKey
		newKey.InputReader = ir
		or := ioutils.NewInMemoryOutputReader(sampleParameters, nil)
		err := discovery.Put(ctx, newKey, or, catalog.Metadata{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Name: "test",
			},
			TaskExecutionIdentifier: nil,
		})
		assert.NoError(t, err)
	})

	t.Run("Create new cached execution with no inputs/outputs", func(t *testing.T) {
		mockClient := &mocks.DataCatalogClient{}
		catalogClient := &CatalogClient{
			client: mockClient,
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
		err := catalogClient.Put(ctx, noInputOutputKey, &mocks2.OutputReader{}, catalog.Metadata{})
		assert.NoError(t, err)
	})

	t.Run("Create new cached execution with existing dataset", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
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
		newKey := sampleKey
		newKey.InputReader = ir
		or := ioutils.NewInMemoryOutputReader(sampleParameters, nil)
		err := discovery.Put(ctx, newKey, or, catalog.Metadata{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Name: "test",
			},
			TaskExecutionIdentifier: nil,
		})
		assert.NoError(t, err)
		assert.True(t, createDatasetCalled)
		assert.True(t, createArtifactCalled)
		assert.True(t, addTagCalled)
	})

}
