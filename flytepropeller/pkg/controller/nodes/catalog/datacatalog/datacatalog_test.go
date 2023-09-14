package datacatalog

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/clients/go/datacatalog/mocks"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog"
	mocks2 "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/ptypes"
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
	Identifier:     core.Identifier{ResourceType: core.ResourceType_TASK, Project: "project", Domain: "domain", Name: "name", Version: "version"},
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

	sampleArtifactData := &datacatalog.ArtifactData{
		Name:  "test",
		Value: newStringLiteral("output1-stringval"),
	}

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
				assert.EqualValues(t, datasetID.String(), o.Dataset.String())
				return true
			}),
		).Return(nil, status.Error(codes.NotFound, "test not found"))
		newKey := sampleKey
		newKey.InputReader = ir
		resp, err := catalogClient.Get(ctx, newKey)
		assert.Error(t, err)

		assertGrpcErr(t, err, codes.NotFound)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_DISABLED, resp.GetStatus().GetCacheStatus())
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
				assert.EqualValues(t, datasetID.String(), o.Dataset.String())
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
		resp, err := catalogClient.Get(ctx, newKey)
		assert.Error(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_DISABLED, resp.GetStatus().GetCacheStatus())
	})

	t.Run("Found w/ tag", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		catalogClient := &CatalogClient{
			client: mockClient,
		}

		taskID := &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Name:         sampleKey.Identifier.Name,
				Project:      sampleKey.Identifier.Project,
				Domain:       sampleKey.Identifier.Domain,
				Version:      "ver",
			},
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Name:    "wf",
					Project: "p1",
					Domain:  "d1",
				},
				NodeId: "n",
			},
			RetryAttempt: 1,
		}
		sampleDataSet := &datacatalog.Dataset{
			Id:       datasetID,
			Metadata: GetDatasetMetadataForSource(taskID),
		}

		mockClient.On("GetDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetDatasetRequest) bool {
				assert.EqualValues(t, datasetID, o.Dataset)
				return true
			}),
		).Return(&datacatalog.GetDatasetResponse{Dataset: sampleDataSet}, nil)

		sampleArtifact := &datacatalog.Artifact{
			Id:       "test-artifact",
			Dataset:  sampleDataSet.Id,
			Data:     []*datacatalog.ArtifactData{sampleArtifactData},
			Metadata: GetArtifactMetadataForSource(taskID),
			Tags: []*datacatalog.Tag{
				{
					Name:       "x",
					ArtifactId: "y",
				},
			},
		}

		assert.Equal(t, taskID.NodeExecutionId.ExecutionId.Name, sampleArtifact.GetMetadata().KeyMap[execNameKey])
		assert.Equal(t, taskID.NodeExecutionId.NodeId, sampleArtifact.GetMetadata().KeyMap[execNodeIDKey])
		assert.Equal(t, taskID.NodeExecutionId.ExecutionId.Project, sampleArtifact.GetMetadata().KeyMap[execProjectKey])
		assert.Equal(t, taskID.NodeExecutionId.ExecutionId.Domain, sampleArtifact.GetMetadata().KeyMap[execDomainKey])
		assert.Equal(t, strconv.Itoa(int(taskID.RetryAttempt)), sampleArtifact.GetMetadata().KeyMap[execTaskAttemptKey])

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
		assert.Equal(t, core.CatalogCacheStatus_CACHE_HIT.String(), resp.GetStatus().GetCacheStatus().String())
		assert.NotNil(t, resp.GetStatus().GetMetadata().DatasetId)
		assert.Equal(t, core.ResourceType_DATASET, resp.GetStatus().GetMetadata().DatasetId.ResourceType)
		assert.Equal(t, datasetID.Name, resp.GetStatus().GetMetadata().DatasetId.Name)
		assert.Equal(t, datasetID.Project, resp.GetStatus().GetMetadata().DatasetId.Project)
		assert.Equal(t, datasetID.Domain, resp.GetStatus().GetMetadata().DatasetId.Domain)
		assert.Equal(t, datasetID.Version, resp.GetStatus().GetMetadata().DatasetId.Version)
		assert.NotNil(t, resp.GetStatus().GetMetadata().ArtifactTag)
		assert.NotNil(t, resp.GetStatus().GetMetadata().SourceExecution)
		sourceTID := resp.GetStatus().GetMetadata().GetSourceTaskExecution()
		assert.Equal(t, taskID.TaskId.String(), sourceTID.TaskId.String())
		assert.Equal(t, taskID.RetryAttempt, sourceTID.RetryAttempt)
		assert.Equal(t, taskID.NodeExecutionId.String(), sourceTID.NodeExecutionId.String())
	})

	t.Run("Found expired artifact", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		catalogClient := &CatalogClient{
			client:      mockClient,
			maxCacheAge: time.Hour,
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
		createdAt, err := ptypes.TimestampProto(time.Now().Add(time.Minute * -61))
		assert.NoError(t, err)

		sampleArtifact := &datacatalog.Artifact{
			Id:        "test-artifact",
			Dataset:   sampleDataSet.Id,
			Data:      []*datacatalog.ArtifactData{sampleArtifactData},
			CreatedAt: createdAt,
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
		assert.Error(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_DISABLED, resp.GetStatus().GetCacheStatus())

		getStatus, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, getStatus.Code(), codes.NotFound)
	})

	t.Run("Found non-expired artifact", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		catalogClient := &CatalogClient{
			client:      mockClient,
			maxCacheAge: time.Hour,
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
		createdAt, err := ptypes.TimestampProto(time.Now().Add(time.Minute * -59))
		assert.NoError(t, err)

		sampleArtifact := &datacatalog.Artifact{
			Id:        "test-artifact",
			Dataset:   sampleDataSet.Id,
			Data:      []*datacatalog.ArtifactData{sampleArtifactData},
			CreatedAt: createdAt,
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
				Version: "1.0.0-GKw-c0Pw-GKw-c0Pw",
			},
		}

		mockClient.On("GetDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetDatasetRequest) bool {
				assert.EqualValues(t, "1.0.0-GKw-c0Pw-GKw-c0Pw", o.Dataset.Version)
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
				assert.EqualValues(t, "1.0.0-GKw-c0Pw-GKw-c0Pw", o.Dataset.Version)
				assert.Equal(t, "flyte_cached-GKw-c0PwFokMUQ6T-TUmEWnZ4_VlQ2Qpgw-vCTT0-OQ", o.GetTagName())
				return true
			}),
		).Return(&datacatalog.GetArtifactResponse{Artifact: sampleArtifact}, nil)

		assert.False(t, discovery.maxCacheAge > time.Duration(0))

		resp, err := discovery.Get(ctx, noInputOutputKey)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		assert.Equal(t, core.CatalogCacheStatus_CACHE_HIT.String(), resp.GetStatus().GetCacheStatus().String())
		v, e, err := resp.GetOutputs().Read(ctx)
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
		or := ioutils.NewInMemoryOutputReader(sampleParameters, nil, nil)
		s, err := discovery.Put(ctx, newKey, or, catalog.Metadata{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Name: "test",
			},
			TaskExecutionIdentifier: nil,
		})
		assert.NoError(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_POPULATED, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
		assert.Equal(t, "flyte_cached-BE6CZsMk6N3ExR_4X9EuwBgj2Jh2UwasXK3a_pM9xlY", s.GetMetadata().ArtifactTag.Name)
	})

	t.Run("Create new cached execution with no inputs/outputs", func(t *testing.T) {
		mockClient := &mocks.DataCatalogClient{}
		catalogClient := &CatalogClient{
			client: mockClient,
		}

		mockClient.On("CreateDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateDatasetRequest) bool {
				assert.Equal(t, "1.0.0-GKw-c0Pw-GKw-c0Pw", o.Dataset.Id.Version)
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
				assert.EqualValues(t, "flyte_cached-GKw-c0PwFokMUQ6T-TUmEWnZ4_VlQ2Qpgw-vCTT0-OQ", o.Tag.Name)
				return true
			}),
		).Return(&datacatalog.AddTagResponse{}, nil)
		s, err := catalogClient.Put(ctx, noInputOutputKey, &mocks2.OutputReader{}, catalog.Metadata{})
		assert.NoError(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_POPULATED, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
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
		or := ioutils.NewInMemoryOutputReader(sampleParameters, nil, nil)
		s, err := discovery.Put(ctx, newKey, or, catalog.Metadata{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Name: "test",
			},
			TaskExecutionIdentifier: nil,
		})
		assert.NoError(t, err)
		assert.True(t, createDatasetCalled)
		assert.True(t, createArtifactCalled)
		assert.True(t, addTagCalled)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_POPULATED, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
	})
}

func TestCatalog_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("Overwrite existing cached execution", func(t *testing.T) {
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

		mockClient.On("UpdateArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.UpdateArtifactRequest) bool {
				assert.True(t, proto.Equal(o.Dataset, datasetID))
				assert.IsType(t, &datacatalog.UpdateArtifactRequest_TagName{}, o.QueryHandle)
				assert.Equal(t, tagName, o.GetTagName())
				return true
			}),
		).Return(&datacatalog.UpdateArtifactResponse{ArtifactId: "test-artifact"}, nil)

		taskID := &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Name:         sampleKey.Identifier.Name,
				Project:      sampleKey.Identifier.Project,
				Domain:       sampleKey.Identifier.Domain,
				Version:      "version",
			},
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Name:    "wf",
					Project: "p1",
					Domain:  "d1",
				},
				NodeId: "unknown", // not set in Put request below --> defaults to "unknown"
			},
			RetryAttempt: 0,
		}

		newKey := sampleKey
		newKey.InputReader = ir
		or := ioutils.NewInMemoryOutputReader(sampleParameters, nil, nil)
		s, err := discovery.Update(ctx, newKey, or, catalog.Metadata{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Name:    taskID.NodeExecutionId.ExecutionId.Name,
				Domain:  taskID.NodeExecutionId.ExecutionId.Domain,
				Project: taskID.NodeExecutionId.ExecutionId.Project,
			},
			TaskExecutionIdentifier: &core.TaskExecutionIdentifier{
				TaskId:          &sampleKey.Identifier,
				NodeExecutionId: taskID.NodeExecutionId,
				RetryAttempt:    0,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_POPULATED, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
		assert.Equal(t, tagName, s.GetMetadata().ArtifactTag.Name)
		sourceTID := s.GetMetadata().GetSourceTaskExecution()
		assert.Equal(t, taskID.TaskId.String(), sourceTID.TaskId.String())
		assert.Equal(t, taskID.RetryAttempt, sourceTID.RetryAttempt)
		assert.Equal(t, taskID.NodeExecutionId.String(), sourceTID.NodeExecutionId.String())
	})

	t.Run("Overwrite non-existing execution", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
		}

		createDatasetCalled := false
		mockClient.On("CreateDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateDatasetRequest) bool {
				assert.True(t, proto.Equal(o.Dataset.Id, datasetID))
				createDatasetCalled = true
				return true
			}),
		).Return(&datacatalog.CreateDatasetResponse{}, nil)

		updateArtifactCalled := false
		mockClient.On("UpdateArtifact", ctx, mock.Anything).Run(func(args mock.Arguments) {
			updateArtifactCalled = true
		}).Return(nil, status.New(codes.NotFound, "missing entity of type Artifact with identifier id").Err())

		createArtifactCalled := false
		mockClient.On("CreateArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateArtifactRequest) bool {
				_, parseErr := uuid.Parse(o.Artifact.Id)
				assert.NoError(t, parseErr)
				assert.True(t, proto.Equal(o.Artifact.Dataset, datasetID))
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

		taskID := &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Name:         sampleKey.Identifier.Name,
				Project:      sampleKey.Identifier.Project,
				Domain:       sampleKey.Identifier.Domain,
				Version:      "version",
			},
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Name:    "wf",
					Project: "p1",
					Domain:  "d1",
				},
				NodeId: "unknown", // not set in Put request below --> defaults to "unknown"
			},
			RetryAttempt: 0,
		}

		newKey := sampleKey
		newKey.InputReader = ir
		or := ioutils.NewInMemoryOutputReader(sampleParameters, nil, nil)
		s, err := discovery.Update(ctx, newKey, or, catalog.Metadata{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Name:    taskID.NodeExecutionId.ExecutionId.Name,
				Domain:  taskID.NodeExecutionId.ExecutionId.Domain,
				Project: taskID.NodeExecutionId.ExecutionId.Project,
			},
			TaskExecutionIdentifier: &core.TaskExecutionIdentifier{
				TaskId:          &sampleKey.Identifier,
				NodeExecutionId: taskID.NodeExecutionId,
				RetryAttempt:    0,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_POPULATED, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
		assert.Equal(t, tagName, s.GetMetadata().ArtifactTag.Name)
		assert.Nil(t, s.GetMetadata().GetSourceTaskExecution())
		assert.True(t, createDatasetCalled)
		assert.True(t, updateArtifactCalled)
		assert.True(t, createArtifactCalled)
		assert.True(t, addTagCalled)
	})

	t.Run("Error while overwriting execution", func(t *testing.T) {
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

		genericErr := errors.New("generic error")
		mockClient.On("UpdateArtifact", ctx, mock.Anything).Return(nil, genericErr)

		newKey := sampleKey
		newKey.InputReader = ir
		or := ioutils.NewInMemoryOutputReader(sampleParameters, nil, nil)
		s, err := discovery.Update(ctx, newKey, or, catalog.Metadata{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Name: "test",
			},
			TaskExecutionIdentifier: nil,
		})
		assert.Error(t, err)
		assert.Equal(t, genericErr, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_DISABLED, s.GetCacheStatus())
		assert.Nil(t, s.GetMetadata())
	})
}

var tagName = "flyte_cached-BE6CZsMk6N3ExR_4X9EuwBgj2Jh2UwasXK3a_pM9xlY"
var reservationID = datacatalog.ReservationID{
	DatasetId: datasetID,
	TagName:   tagName,
}
var prevOwner = "prevOwner"
var currentOwner = "currentOwner"

func TestCatalog_GetOrExtendReservation(t *testing.T) {
	ctx := context.Background()

	heartbeatInterval := time.Second * 5
	prevReservation := datacatalog.Reservation{
		ReservationId: &reservationID,
		OwnerId:       prevOwner,
	}

	currentReservation := datacatalog.Reservation{
		ReservationId: &reservationID,
		OwnerId:       currentOwner,
	}

	t.Run("CreateOrUpdateReservation", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		catalogClient := &CatalogClient{
			client: mockClient,
		}

		mockClient.On("GetOrExtendReservation",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetOrExtendReservationRequest) bool {
				assert.EqualValues(t, datasetID.String(), o.ReservationId.DatasetId.String())
				assert.EqualValues(t, tagName, o.ReservationId.TagName)
				return true
			}),
		).Return(&datacatalog.GetOrExtendReservationResponse{Reservation: &currentReservation}, nil, "")

		newKey := sampleKey
		newKey.InputReader = ir
		reservation, err := catalogClient.GetOrExtendReservation(ctx, newKey, currentOwner, heartbeatInterval)

		assert.NoError(t, err)
		assert.Equal(t, reservation.OwnerId, currentOwner)
	})

	t.Run("ExistingReservation", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		catalogClient := &CatalogClient{
			client: mockClient,
		}

		mockClient.On("GetOrExtendReservation",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetOrExtendReservationRequest) bool {
				assert.EqualValues(t, datasetID.String(), o.ReservationId.DatasetId.String())
				assert.EqualValues(t, tagName, o.ReservationId.TagName)
				return true
			}),
		).Return(&datacatalog.GetOrExtendReservationResponse{Reservation: &prevReservation}, nil, "")

		newKey := sampleKey
		newKey.InputReader = ir
		reservation, err := catalogClient.GetOrExtendReservation(ctx, newKey, currentOwner, heartbeatInterval)

		assert.NoError(t, err)
		assert.Equal(t, reservation.OwnerId, prevOwner)
	})
}

func TestCatalog_ReleaseReservation(t *testing.T) {
	ctx := context.Background()

	t.Run("ReleaseReservation", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		catalogClient := &CatalogClient{
			client: mockClient,
		}

		mockClient.On("ReleaseReservation",
			ctx,
			mock.MatchedBy(func(o *datacatalog.ReleaseReservationRequest) bool {
				assert.EqualValues(t, datasetID.String(), o.ReservationId.DatasetId.String())
				assert.EqualValues(t, tagName, o.ReservationId.TagName)
				return true
			}),
		).Return(&datacatalog.ReleaseReservationResponse{}, nil, "")

		newKey := sampleKey
		newKey.InputReader = ir
		err := catalogClient.ReleaseReservation(ctx, newKey, currentOwner)

		assert.NoError(t, err)
	})

	t.Run("ReleaseReservationFailure", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		catalogClient := &CatalogClient{
			client: mockClient,
		}

		mockClient.On("ReleaseReservation",
			ctx,
			mock.MatchedBy(func(o *datacatalog.ReleaseReservationRequest) bool {
				assert.EqualValues(t, datasetID.String(), o.ReservationId.DatasetId.String())
				assert.EqualValues(t, tagName, o.ReservationId.TagName)
				return true
			}),
		).Return(nil, status.Error(codes.NotFound, "reservation not found"))

		newKey := sampleKey
		newKey.InputReader = ir
		err := catalogClient.ReleaseReservation(ctx, newKey, currentOwner)

		assertGrpcErr(t, err, codes.NotFound)
	})
}
