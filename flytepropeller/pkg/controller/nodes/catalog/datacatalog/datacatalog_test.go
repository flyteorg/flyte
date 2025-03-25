package datacatalog

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteidl/clients/go/datacatalog/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	inputReaderMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	mocks2 "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	futureFileReaderMocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/interfaces/mocks"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
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
				assert.EqualValues(t, datasetID.String(), o.GetDataset().String())
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
				assert.EqualValues(t, datasetID.String(), o.GetDataset().String())
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
				Name:         sampleKey.Identifier.GetName(),
				Project:      sampleKey.Identifier.GetProject(),
				Domain:       sampleKey.Identifier.GetDomain(),
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
				assert.EqualValues(t, datasetID, o.GetDataset())
				return true
			}),
		).Return(&datacatalog.GetDatasetResponse{Dataset: sampleDataSet}, nil)

		sampleArtifact := &datacatalog.Artifact{
			Id:       "test-artifact",
			Dataset:  sampleDataSet.GetId(),
			Data:     []*datacatalog.ArtifactData{sampleArtifactData},
			Metadata: GetArtifactMetadataForSource(taskID),
			Tags: []*datacatalog.Tag{
				{
					Name:       "x",
					ArtifactId: "y",
				},
			},
		}

		assert.Equal(t, taskID.GetNodeExecutionId().GetExecutionId().GetName(), sampleArtifact.GetMetadata().GetKeyMap()[execNameKey])
		assert.Equal(t, taskID.GetNodeExecutionId().GetNodeId(), sampleArtifact.GetMetadata().GetKeyMap()[execNodeIDKey])
		assert.Equal(t, taskID.GetNodeExecutionId().GetExecutionId().GetProject(), sampleArtifact.GetMetadata().GetKeyMap()[execProjectKey])
		assert.Equal(t, taskID.GetNodeExecutionId().GetExecutionId().GetDomain(), sampleArtifact.GetMetadata().GetKeyMap()[execDomainKey])
		assert.Equal(t, strconv.Itoa(int(taskID.GetRetryAttempt())), sampleArtifact.GetMetadata().GetKeyMap()[execTaskAttemptKey])

		mockClient.On("GetArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetArtifactRequest) bool {
				assert.EqualValues(t, datasetID, o.GetDataset())
				assert.Equal(t, "flyte_cached-BE6CZsMk6N3ExR_4X9EuwBgj2Jh2UwasXK3a_pM9xlY", o.GetTagName())
				return true
			}),
		).Return(&datacatalog.GetArtifactResponse{Artifact: sampleArtifact}, nil)

		newKey := sampleKey
		newKey.InputReader = ir
		resp, err := catalogClient.Get(ctx, newKey)
		assert.NoError(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_HIT.String(), resp.GetStatus().GetCacheStatus().String())
		assert.NotNil(t, resp.GetStatus().GetMetadata().GetDatasetId())
		assert.Equal(t, core.ResourceType_DATASET, resp.GetStatus().GetMetadata().GetDatasetId().GetResourceType())
		assert.Equal(t, datasetID.GetName(), resp.GetStatus().GetMetadata().GetDatasetId().GetName())
		assert.Equal(t, datasetID.GetProject(), resp.GetStatus().GetMetadata().GetDatasetId().GetProject())
		assert.Equal(t, datasetID.GetDomain(), resp.GetStatus().GetMetadata().GetDatasetId().GetDomain())
		assert.Equal(t, datasetID.GetVersion(), resp.GetStatus().GetMetadata().GetDatasetId().GetVersion())
		assert.NotNil(t, resp.GetStatus().GetMetadata().GetArtifactTag())
		assert.NotNil(t, resp.GetStatus().GetMetadata().GetSourceExecution())
		sourceTID := resp.GetStatus().GetMetadata().GetSourceTaskExecution()
		assert.Equal(t, taskID.GetTaskId().String(), sourceTID.GetTaskId().String())
		assert.Equal(t, taskID.GetRetryAttempt(), sourceTID.GetRetryAttempt())
		assert.Equal(t, taskID.GetNodeExecutionId().String(), sourceTID.GetNodeExecutionId().String())
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
				assert.EqualValues(t, datasetID, o.GetDataset())
				return true
			}),
		).Return(&datacatalog.GetDatasetResponse{Dataset: sampleDataSet}, nil)
		createdAt, err := ptypes.TimestampProto(time.Now().Add(time.Minute * -61))
		assert.NoError(t, err)

		sampleArtifact := &datacatalog.Artifact{
			Id:        "test-artifact",
			Dataset:   sampleDataSet.GetId(),
			Data:      []*datacatalog.ArtifactData{sampleArtifactData},
			CreatedAt: createdAt,
		}
		mockClient.On("GetArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetArtifactRequest) bool {
				assert.EqualValues(t, datasetID, o.GetDataset())
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
				assert.EqualValues(t, datasetID, o.GetDataset())
				return true
			}),
		).Return(&datacatalog.GetDatasetResponse{Dataset: sampleDataSet}, nil)
		createdAt, err := ptypes.TimestampProto(time.Now().Add(time.Minute * -59))
		assert.NoError(t, err)

		sampleArtifact := &datacatalog.Artifact{
			Id:        "test-artifact",
			Dataset:   sampleDataSet.GetId(),
			Data:      []*datacatalog.ArtifactData{sampleArtifactData},
			CreatedAt: createdAt,
		}
		mockClient.On("GetArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetArtifactRequest) bool {
				assert.EqualValues(t, datasetID, o.GetDataset())
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
				assert.EqualValues(t, "1.0.0-GKw-c0Pw-GKw-c0Pw", o.GetDataset().GetVersion())
				return true
			}),
		).Return(&datacatalog.GetDatasetResponse{Dataset: sampleDataSet}, nil)

		sampleArtifact := &datacatalog.Artifact{
			Id:      "test-artifact",
			Dataset: sampleDataSet.GetId(),
			Data:    []*datacatalog.ArtifactData{},
		}
		mockClient.On("GetArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetArtifactRequest) bool {
				assert.EqualValues(t, "1.0.0-GKw-c0Pw-GKw-c0Pw", o.GetDataset().GetVersion())
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
		assert.Len(t, v.GetLiterals(), 0)
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
				assert.True(t, proto.Equal(o.GetDataset().GetId(), datasetID))
				return true
			}),
		).Return(&datacatalog.CreateDatasetResponse{}, nil)

		mockClient.On("CreateArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateArtifactRequest) bool {
				_, parseErr := uuid.Parse(o.GetArtifact().GetId())
				assert.NoError(t, parseErr)
				assert.EqualValues(t, 1, len(o.GetArtifact().GetData()))
				assert.EqualValues(t, "out1", o.GetArtifact().GetData()[0].GetName())
				assert.True(t, proto.Equal(newStringLiteral("output1-stringval"), o.GetArtifact().GetData()[0].GetValue()))
				return true
			}),
		).Return(&datacatalog.CreateArtifactResponse{}, nil)

		mockClient.On("AddTag",
			ctx,
			mock.MatchedBy(func(o *datacatalog.AddTagRequest) bool {
				assert.EqualValues(t, "flyte_cached-BE6CZsMk6N3ExR_4X9EuwBgj2Jh2UwasXK3a_pM9xlY", o.GetTag().GetName())
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
		assert.Equal(t, "flyte_cached-BE6CZsMk6N3ExR_4X9EuwBgj2Jh2UwasXK3a_pM9xlY", s.GetMetadata().GetArtifactTag().GetName())
	})

	t.Run("Create dataset fails", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
		}

		mockClient.On("CreateDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateDatasetRequest) bool {
				return true
			}),
		).Return(&datacatalog.CreateDatasetResponse{}, errors.New("generic error"))

		newKey := sampleKey
		newKey.InputReader = ir
		or := ioutils.NewInMemoryOutputReader(sampleParameters, nil, nil)
		s, err := discovery.Put(ctx, newKey, or, catalog.Metadata{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Name: "test",
			},
			TaskExecutionIdentifier: nil,
		})
		assert.Error(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_PUT_FAILURE, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
	})

	t.Run("Create artifact fails", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
		}

		mockClient.On("CreateDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateDatasetRequest) bool {
				return true
			}),
		).Return(&datacatalog.CreateDatasetResponse{}, nil)

		mockClient.On("CreateArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateArtifactRequest) bool {
				return true
			}),
		).Return(&datacatalog.CreateArtifactResponse{}, errors.New("generic error"))

		newKey := sampleKey
		newKey.InputReader = ir
		or := ioutils.NewInMemoryOutputReader(sampleParameters, nil, nil)
		s, err := discovery.Put(ctx, newKey, or, catalog.Metadata{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Name: "test",
			},
			TaskExecutionIdentifier: nil,
		})
		assert.Error(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_PUT_FAILURE, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
	})

	t.Run("Create new cached execution with no inputs/outputs", func(t *testing.T) {
		mockClient := &mocks.DataCatalogClient{}
		catalogClient := &CatalogClient{
			client: mockClient,
		}

		mockClient.On("CreateDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateDatasetRequest) bool {
				assert.Equal(t, "1.0.0-GKw-c0Pw-GKw-c0Pw", o.GetDataset().GetId().GetVersion())
				return true
			}),
		).Return(&datacatalog.CreateDatasetResponse{}, nil)

		mockClient.On("CreateArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateArtifactRequest) bool {
				assert.EqualValues(t, 0, len(o.GetArtifact().GetData()))
				return true
			}),
		).Return(&datacatalog.CreateArtifactResponse{}, nil)

		mockClient.On("AddTag",
			ctx,
			mock.MatchedBy(func(o *datacatalog.AddTagRequest) bool {
				assert.EqualValues(t, "flyte_cached-GKw-c0PwFokMUQ6T-TUmEWnZ4_VlQ2Qpgw-vCTT0-OQ", o.GetTag().GetName())
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
				_, parseErr := uuid.Parse(o.GetArtifact().GetId())
				assert.NoError(t, parseErr)
				assert.EqualValues(t, 1, len(o.GetArtifact().GetData()))
				assert.EqualValues(t, "out1", o.GetArtifact().GetData()[0].GetName())
				assert.True(t, proto.Equal(newStringLiteral("output1-stringval"), o.GetArtifact().GetData()[0].GetValue()))
				createArtifactCalled = true
				return true
			}),
		).Return(&datacatalog.CreateArtifactResponse{}, nil)

		addTagCalled := false
		mockClient.On("AddTag",
			ctx,
			mock.MatchedBy(func(o *datacatalog.AddTagRequest) bool {
				assert.EqualValues(t, "flyte_cached-BE6CZsMk6N3ExR_4X9EuwBgj2Jh2UwasXK3a_pM9xlY", o.GetTag().GetName())
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
				assert.True(t, proto.Equal(o.GetDataset().GetId(), datasetID))
				return true
			}),
		).Return(&datacatalog.CreateDatasetResponse{}, nil)

		mockClient.On("UpdateArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.UpdateArtifactRequest) bool {
				assert.True(t, proto.Equal(o.GetDataset(), datasetID))
				assert.IsType(t, &datacatalog.UpdateArtifactRequest_TagName{}, o.GetQueryHandle())
				assert.Equal(t, tagName, o.GetTagName())
				return true
			}),
		).Return(&datacatalog.UpdateArtifactResponse{ArtifactId: "test-artifact"}, nil)

		taskID := &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Name:         sampleKey.Identifier.GetName(),
				Project:      sampleKey.Identifier.GetProject(),
				Domain:       sampleKey.Identifier.GetDomain(),
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
				Name:    taskID.GetNodeExecutionId().GetExecutionId().GetName(),
				Domain:  taskID.GetNodeExecutionId().GetExecutionId().GetDomain(),
				Project: taskID.GetNodeExecutionId().GetExecutionId().GetProject(),
			},
			TaskExecutionIdentifier: &core.TaskExecutionIdentifier{
				TaskId:          &sampleKey.Identifier,
				NodeExecutionId: taskID.GetNodeExecutionId(),
				RetryAttempt:    0,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_POPULATED, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
		assert.Equal(t, tagName, s.GetMetadata().GetArtifactTag().GetName())
		sourceTID := s.GetMetadata().GetSourceTaskExecution()
		assert.Equal(t, taskID.GetTaskId().String(), sourceTID.GetTaskId().String())
		assert.Equal(t, taskID.GetRetryAttempt(), sourceTID.GetRetryAttempt())
		assert.Equal(t, taskID.GetNodeExecutionId().String(), sourceTID.GetNodeExecutionId().String())
	})

	t.Run("Overwrite non-existing execution", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
		}

		mockClient.On("CreateDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateDatasetRequest) bool {
				return true
			}),
		).Return(&datacatalog.CreateDatasetResponse{}, nil)

		mockClient.On("UpdateArtifact", ctx, mock.Anything).Return(nil, status.New(codes.NotFound, "missing entity of type Artifact with identifier id").Err())

		mockClient.On("CreateArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateArtifactRequest) bool {
				return true
			}),
		).Return(&datacatalog.CreateArtifactResponse{}, errors.New("generic error"))

		taskID := &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Name:         sampleKey.Identifier.GetName(),
				Project:      sampleKey.Identifier.GetProject(),
				Domain:       sampleKey.Identifier.GetDomain(),
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
				Name:    taskID.GetNodeExecutionId().GetExecutionId().GetName(),
				Domain:  taskID.GetNodeExecutionId().GetExecutionId().GetDomain(),
				Project: taskID.GetNodeExecutionId().GetExecutionId().GetProject(),
			},
			TaskExecutionIdentifier: &core.TaskExecutionIdentifier{
				TaskId:          &sampleKey.Identifier,
				NodeExecutionId: taskID.GetNodeExecutionId(),
				RetryAttempt:    0,
			},
		})
		assert.Error(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_PUT_FAILURE, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
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
				assert.True(t, proto.Equal(o.GetDataset().GetId(), datasetID))
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
				_, parseErr := uuid.Parse(o.GetArtifact().GetId())
				assert.NoError(t, parseErr)
				assert.True(t, proto.Equal(o.GetArtifact().GetDataset(), datasetID))
				createArtifactCalled = true
				return true
			}),
		).Return(&datacatalog.CreateArtifactResponse{}, nil)

		addTagCalled := false
		mockClient.On("AddTag",
			ctx,
			mock.MatchedBy(func(o *datacatalog.AddTagRequest) bool {
				assert.EqualValues(t, "flyte_cached-BE6CZsMk6N3ExR_4X9EuwBgj2Jh2UwasXK3a_pM9xlY", o.GetTag().GetName())
				addTagCalled = true
				return true
			}),
		).Return(&datacatalog.AddTagResponse{}, nil)

		taskID := &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Name:         sampleKey.Identifier.GetName(),
				Project:      sampleKey.Identifier.GetProject(),
				Domain:       sampleKey.Identifier.GetDomain(),
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
				Name:    taskID.GetNodeExecutionId().GetExecutionId().GetName(),
				Domain:  taskID.GetNodeExecutionId().GetExecutionId().GetDomain(),
				Project: taskID.GetNodeExecutionId().GetExecutionId().GetProject(),
			},
			TaskExecutionIdentifier: &core.TaskExecutionIdentifier{
				TaskId:          &sampleKey.Identifier,
				NodeExecutionId: taskID.GetNodeExecutionId(),
				RetryAttempt:    0,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_POPULATED, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
		assert.Equal(t, tagName, s.GetMetadata().GetArtifactTag().GetName())
		assert.Nil(t, s.GetMetadata().GetSourceTaskExecution())
		assert.True(t, createDatasetCalled)
		assert.True(t, updateArtifactCalled)
		assert.True(t, createArtifactCalled)
		assert.True(t, addTagCalled)
	})

	t.Run("Error while creating dataset", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
		}

		mockClient.On("CreateDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateDatasetRequest) bool {
				return true
			}),
		).Return(&datacatalog.CreateDatasetResponse{}, errors.New("generic error"))

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
		assert.Equal(t, core.CatalogCacheStatus_CACHE_PUT_FAILURE, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
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
		assert.Equal(t, core.CatalogCacheStatus_CACHE_PUT_FAILURE, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
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
				assert.EqualValues(t, datasetID.String(), o.GetReservationId().GetDatasetId().String())
				assert.EqualValues(t, tagName, o.GetReservationId().GetTagName())
				return true
			}),
		).Return(&datacatalog.GetOrExtendReservationResponse{Reservation: &currentReservation}, nil, "")

		newKey := sampleKey
		newKey.InputReader = ir
		reservation, err := catalogClient.GetOrExtendReservation(ctx, newKey, currentOwner, heartbeatInterval)

		assert.NoError(t, err)
		assert.Equal(t, reservation.GetOwnerId(), currentOwner)
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
				assert.EqualValues(t, datasetID.String(), o.GetReservationId().GetDatasetId().String())
				assert.EqualValues(t, tagName, o.GetReservationId().GetTagName())
				return true
			}),
		).Return(&datacatalog.GetOrExtendReservationResponse{Reservation: &prevReservation}, nil, "")

		newKey := sampleKey
		newKey.InputReader = ir
		reservation, err := catalogClient.GetOrExtendReservation(ctx, newKey, currentOwner, heartbeatInterval)

		assert.NoError(t, err)
		assert.Equal(t, reservation.GetOwnerId(), prevOwner)
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
				assert.EqualValues(t, datasetID.String(), o.GetReservationId().GetDatasetId().String())
				assert.EqualValues(t, tagName, o.GetReservationId().GetTagName())
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
				assert.EqualValues(t, datasetID.String(), o.GetReservationId().GetDatasetId().String())
				assert.EqualValues(t, tagName, o.GetReservationId().GetTagName())
				return true
			}),
		).Return(nil, status.Error(codes.NotFound, "reservation not found"))

		newKey := sampleKey
		newKey.InputReader = ir
		err := catalogClient.ReleaseReservation(ctx, newKey, currentOwner)

		assertGrpcErr(t, err, codes.NotFound)
	})
}

func TestGetFutureArtifactByTag(t *testing.T) {
	ctx := context.Background()
	mockClient := &mocks.DataCatalogClient{}

	dataset := &datacatalog.Dataset{
		Id: &datacatalog.DatasetID{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Version: "version",
		},
	}

	tagName := "test-tag"

	t.Run("Successfully retrieve non-expired artifact", func(t *testing.T) {
		createdAt, _ := ptypes.TimestampProto(time.Now().Add(-10 * time.Minute))
		expectedArtifact := &datacatalog.Artifact{
			Id:        "test-artifact",
			CreatedAt: createdAt,
		}

		mockClient.On("GetFutureArtifact", ctx, &datacatalog.GetArtifactRequest{
			Dataset: dataset.GetId(),
			QueryHandle: &datacatalog.GetArtifactRequest_TagName{
				TagName: tagName,
			},
		}).Return(&datacatalog.GetArtifactResponse{
			Artifact: expectedArtifact,
		}, nil).Once()

		catalogClient := &CatalogClient{
			client:      mockClient,
			maxCacheAge: 1 * time.Hour,
		}

		artifact, err := catalogClient.GetFutureArtifactByTag(ctx, tagName, dataset)

		assert.NoError(t, err)
		assert.Equal(t, expectedArtifact, artifact)
	})

	t.Run("Expired artifact should return NotFound error", func(t *testing.T) {
		// Create timestamp from 2 hours ago
		createdAt, _ := ptypes.TimestampProto(time.Now().Add(-2 * time.Hour))
		expiredArtifact := &datacatalog.Artifact{
			Id:        "expired-artifact",
			CreatedAt: createdAt,
		}

		mockClient.On("GetFutureArtifact", ctx, &datacatalog.GetArtifactRequest{
			Dataset: dataset.GetId(),
			QueryHandle: &datacatalog.GetArtifactRequest_TagName{
				TagName: tagName,
			},
		}).Return(&datacatalog.GetArtifactResponse{
			Artifact: expiredArtifact,
		}, nil).Once()

		catalogClient := &CatalogClient{
			client:      mockClient,
			maxCacheAge: 1 * time.Hour,
		}

		artifact, err := catalogClient.GetFutureArtifactByTag(ctx, tagName, dataset)

		assert.Error(t, err)
		assert.Nil(t, artifact)
		assert.Equal(t, codes.NotFound, status.Code(err))
		assert.Contains(t, err.Error(), "Artifact over age limit")
	})

	t.Run("Invalid createdAt timestamp should return error", func(t *testing.T) {
		invalidArtifact := &datacatalog.Artifact{
			Id: "invalid-artifact",
			CreatedAt: &timestamp.Timestamp{
				Seconds: -1,
				Nanos:   0,
			},
		}

		mockClient.On("GetFutureArtifact", ctx, &datacatalog.GetArtifactRequest{
			Dataset: dataset.GetId(),
			QueryHandle: &datacatalog.GetArtifactRequest_TagName{
				TagName: tagName,
			},
		}).Return(&datacatalog.GetArtifactResponse{
			Artifact: invalidArtifact,
		}, nil).Once()

		catalogClient := &CatalogClient{
			client:      mockClient,
			maxCacheAge: 1 * time.Hour,
		}

		artifact, err := catalogClient.GetFutureArtifactByTag(ctx, tagName, dataset)

		assert.Error(t, err)
		assert.Nil(t, artifact)
	})

	t.Run("Should return error when client returns error", func(t *testing.T) {
		expectedError := status.Error(codes.Internal, "internal error")

		mockClient.On("GetFutureArtifact", ctx, &datacatalog.GetArtifactRequest{
			Dataset: dataset.GetId(),
			QueryHandle: &datacatalog.GetArtifactRequest_TagName{
				TagName: tagName,
			},
		}).Return((*datacatalog.GetArtifactResponse)(nil), expectedError).Once()

		catalogClient := &CatalogClient{
			client:      mockClient,
			maxCacheAge: 1 * time.Hour,
		}

		artifact, err := catalogClient.GetFutureArtifactByTag(ctx, tagName, dataset)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, artifact)
	})
}

func TestCatalog_GetFuture(t *testing.T) {
	ctx := context.Background()

	sampleArtifactData := &datacatalog.ArtifactData{
		Name:  "future",
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
				assert.EqualValues(t, datasetID.String(), o.GetDataset().String())
				return true
			}),
		).Return(nil, status.Error(codes.NotFound, "test not found"))

		newKey := sampleKey
		newKey.InputReader = ir
		future, err := catalogClient.GetFuture(ctx, newKey)
		assert.NotNil(t, future)

		resp := future
		assert.Error(t, err)
		assertGrpcErr(t, err, codes.NotFound)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_DISABLED, resp.GetStatus().GetCacheStatus())
	})

	t.Run("Found with tag name", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		catalogClient := &CatalogClient{
			client: mockClient,
		}

		taskID := &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Name:         sampleKey.Identifier.GetName(),
				Project:      sampleKey.Identifier.GetProject(),
				Domain:       sampleKey.Identifier.GetDomain(),
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
				assert.EqualValues(t, datasetID, o.GetDataset())
				return true
			}),
		).Return(&datacatalog.GetDatasetResponse{Dataset: sampleDataSet}, nil)

		sampleArtifact := &datacatalog.Artifact{
			Id:       "test-artifact",
			Dataset:  sampleDataSet.GetId(),
			Data:     []*datacatalog.ArtifactData{sampleArtifactData},
			Metadata: GetArtifactMetadataForSource(taskID),
		}

		mockClient.On("GetFutureArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetArtifactRequest) bool {
				assert.EqualValues(t, datasetID, o.GetDataset())
				assert.Equal(t, "flyte_cached_future-BE6CZsMk6N3ExR_4X9EuwBgj2Jh2UwasXK3a_pM9xlY", o.GetTagName())
				return true
			}),
		).Return(&datacatalog.GetArtifactResponse{Artifact: sampleArtifact}, nil)

		newKey := sampleKey
		newKey.InputReader = ir
		future, err := catalogClient.GetFuture(ctx, newKey)
		assert.NoError(t, err)
		assert.NotNil(t, future)

		resp := future
		assert.Equal(t, core.CatalogCacheStatus_CACHE_HIT.String(), resp.GetStatus().GetCacheStatus().String())
		assert.NotNil(t, resp.GetStatus().GetMetadata().GetDatasetId())
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
				assert.EqualValues(t, datasetID, o.GetDataset())
				return true
			}),
		).Return(&datacatalog.GetDatasetResponse{Dataset: sampleDataSet}, nil)

		createdAt, err := ptypes.TimestampProto(time.Now().Add(time.Minute * -70))
		assert.NoError(t, err)

		sampleArtifact := &datacatalog.Artifact{
			Id:        "test-artifact",
			Dataset:   sampleDataSet.GetId(),
			Data:      []*datacatalog.ArtifactData{sampleArtifactData},
			CreatedAt: createdAt,
		}

		mockClient.On("GetFutureArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.GetArtifactRequest) bool {
				return true
			}),
		).Return(&datacatalog.GetArtifactResponse{Artifact: sampleArtifact}, nil)

		newKey := sampleKey
		newKey.InputReader = ir
		future, err := catalogClient.GetFuture(ctx, newKey)
		assert.NotNil(t, future)

		resp := future
		assert.Error(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_DISABLED, resp.GetStatus().GetCacheStatus())
	})
}

func TestCatalogClient_prepareInputsAndFuture(t *testing.T) {
	ctx := context.Background()

	t.Run("Success case - With input variables", func(t *testing.T) {
		// Prepare test data
		mockInputReader := &inputReaderMocks.InputReader{}
		mockFutureReader := &futureFileReaderMocks.FutureFileReaderInterface{}
		expectedInputs := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"test": {
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_Integer{
										Integer: 42,
									},
								},
							},
						},
					},
				},
			},
		}
		expectedFuture := &core.DynamicJobSpec{
			Tasks: []*core.TaskTemplate{
				{
					Id: &core.Identifier{
						Name: "test_task",
					},
				},
			},
		}

		mockInputReader.On("Get", mock.Anything).Return(expectedInputs, nil)
		mockFutureReader.On("Read", mock.Anything).Return(expectedFuture, nil)

		key := catalog.Key{
			TypedInterface: core.TypedInterface{
				Inputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"test": {
							Type: &core.LiteralType{
								Type: &core.LiteralType_Simple{
									Simple: core.SimpleType_INTEGER,
								},
							},
						},
					},
				},
			},
			InputReader: mockInputReader,
		}

		// Execute test
		inputs, future, err := (&CatalogClient{}).prepareInputsAndFuture(ctx, key, mockFutureReader)

		// Verify results
		assert.NoError(t, err)
		assert.Equal(t, expectedInputs, inputs)
		assert.Equal(t, expectedFuture, future)
		mockInputReader.AssertExpectations(t)
		mockFutureReader.AssertExpectations(t)
	})

	t.Run("Success case - No input variables", func(t *testing.T) {
		mockFutureReader := &futureFileReaderMocks.FutureFileReaderInterface{}
		expectedFuture := &core.DynamicJobSpec{
			Tasks: []*core.TaskTemplate{
				{
					Id: &core.Identifier{
						Name: "test_task",
					},
				},
			},
		}

		mockFutureReader.On("Read", mock.Anything).Return(expectedFuture, nil)

		key := catalog.Key{
			TypedInterface: core.TypedInterface{},
		}

		inputs, future, err := (&CatalogClient{}).prepareInputsAndFuture(ctx, key, mockFutureReader)

		assert.NoError(t, err)
		assert.Equal(t, &core.LiteralMap{}, inputs)
		assert.Equal(t, expectedFuture, future)
		mockFutureReader.AssertExpectations(t)
	})

	t.Run("Error case - InputReader.Get fails", func(t *testing.T) {
		mockInputReader := &inputReaderMocks.InputReader{}
		mockFutureReader := &futureFileReaderMocks.FutureFileReaderInterface{}
		expectedError := errors.New("input reader error")

		// Fix: Use On() instead of OnGetMatch()
		mockInputReader.On("Get", ctx).Return(nil, expectedError)

		key := catalog.Key{
			TypedInterface: core.TypedInterface{
				Inputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"test": {
							Type: &core.LiteralType{
								Type: &core.LiteralType_Simple{
									Simple: core.SimpleType_INTEGER,
								},
							},
						},
					},
				},
			},
			InputReader: mockInputReader,
		}

		inputs, future, err := (&CatalogClient{}).prepareInputsAndFuture(ctx, key, mockFutureReader)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, inputs)
		assert.Nil(t, future)
		mockInputReader.AssertExpectations(t)
	})
}

func TestCatalog_PutFuture(t *testing.T) {
	ctx := context.Background()

	t.Run("Create new cached future execution", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockFutureReader := &futureFileReaderMocks.FutureFileReaderInterface{}
		sampleDJSpec := &core.DynamicJobSpec{
			Tasks: []*core.TaskTemplate{
				{
					Id: &core.Identifier{
						Name: "task1",
					},
				},
			},
		}
		mockFutureReader.On("Read", mock.Anything).Return(sampleDJSpec, nil)

		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
		}

		mockClient.On("CreateDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateDatasetRequest) bool {
				assert.True(t, proto.Equal(o.GetDataset().GetId(), datasetID))
				return true
			}),
		).Return(&datacatalog.CreateDatasetResponse{}, nil)

		mockClient.On("CreateFutureArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateArtifactRequest) bool {
				_, parseErr := uuid.Parse(o.GetArtifact().GetId())
				assert.NoError(t, parseErr)
				assert.EqualValues(t, 1, len(o.GetArtifact().GetData()))
				assert.EqualValues(t, "future", o.GetArtifact().GetData()[0].GetName())
				// 验证序列化的DynamicJobSpec
				binary := o.GetArtifact().GetData()[0].GetValue().GetScalar().GetBinary().GetValue()
				var djSpec core.DynamicJobSpec
				err := proto.Unmarshal(binary, &djSpec)
				assert.NoError(t, err)
				assert.True(t, proto.Equal(sampleDJSpec, &djSpec))
				return true
			}),
		).Return(&datacatalog.CreateArtifactResponse{}, nil)

		mockClient.On("AddTag",
			ctx,
			mock.MatchedBy(func(o *datacatalog.AddTagRequest) bool {
				assert.EqualValues(t, "flyte_cached_future-BE6CZsMk6N3ExR_4X9EuwBgj2Jh2UwasXK3a_pM9xlY", o.GetTag().GetName())
				return true
			}),
		).Return(&datacatalog.AddTagResponse{}, nil)

		newKey := sampleKey
		newKey.InputReader = ir
		s, err := discovery.PutFuture(ctx, newKey, mockFutureReader, catalog.Metadata{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Name: "test",
			},
			TaskExecutionIdentifier: nil,
		})
		assert.NoError(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_POPULATED, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
		assert.Equal(t, "flyte_cached_future-BE6CZsMk6N3ExR_4X9EuwBgj2Jh2UwasXK3a_pM9xlY", s.GetMetadata().GetArtifactTag().GetName())
	})

	t.Run("Create dataset fails", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockFutureReader := &futureFileReaderMocks.FutureFileReaderInterface{}
		mockFutureReader.On("Read", mock.Anything).Return(&core.DynamicJobSpec{}, nil)

		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
		}

		mockClient.On("CreateDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateDatasetRequest) bool {
				return true
			}),
		).Return(&datacatalog.CreateDatasetResponse{}, errors.New("generic error"))

		newKey := sampleKey
		newKey.InputReader = ir
		s, err := discovery.PutFuture(ctx, newKey, mockFutureReader, catalog.Metadata{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Name: "test",
			},
			TaskExecutionIdentifier: nil,
		})
		assert.Error(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_PUT_FAILURE, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
	})

	t.Run("Create future artifact fails", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockFutureReader := &futureFileReaderMocks.FutureFileReaderInterface{}
		mockFutureReader.On("Read", mock.Anything).Return(&core.DynamicJobSpec{}, nil)

		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
		}

		mockClient.On("CreateDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateDatasetRequest) bool {
				return true
			}),
		).Return(&datacatalog.CreateDatasetResponse{}, nil)

		mockClient.On("CreateFutureArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateArtifactRequest) bool {
				return true
			}),
		).Return(&datacatalog.CreateArtifactResponse{}, errors.New("generic error"))

		newKey := sampleKey
		newKey.InputReader = ir
		s, err := discovery.PutFuture(ctx, newKey, mockFutureReader, catalog.Metadata{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Name: "test",
			},
			TaskExecutionIdentifier: nil,
		})
		assert.Error(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_PUT_FAILURE, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
	})

	t.Run("Create new cached future execution with existing dataset", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockFutureReader := &futureFileReaderMocks.FutureFileReaderInterface{}
		sampleDJSpec := &core.DynamicJobSpec{
			Tasks: []*core.TaskTemplate{
				{
					Id: &core.Identifier{
						Name: "task1",
					},
				},
			},
		}
		mockFutureReader.On("Read", mock.Anything).Return(sampleDJSpec, nil)

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

		createFutureArtifactCalled := false
		mockClient.On("CreateFutureArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateArtifactRequest) bool {
				_, parseErr := uuid.Parse(o.GetArtifact().GetId())
				assert.NoError(t, parseErr)
				createFutureArtifactCalled = true
				return true
			}),
		).Return(&datacatalog.CreateArtifactResponse{}, nil)

		addTagCalled := false
		mockClient.On("AddTag",
			ctx,
			mock.MatchedBy(func(o *datacatalog.AddTagRequest) bool {
				assert.EqualValues(t, "flyte_cached_future-BE6CZsMk6N3ExR_4X9EuwBgj2Jh2UwasXK3a_pM9xlY", o.GetTag().GetName())
				addTagCalled = true
				return true
			}),
		).Return(&datacatalog.AddTagResponse{}, nil)

		newKey := sampleKey
		newKey.InputReader = ir
		s, err := discovery.PutFuture(ctx, newKey, mockFutureReader, catalog.Metadata{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Name: "test",
			},
			TaskExecutionIdentifier: nil,
		})
		assert.NoError(t, err)
		assert.True(t, createDatasetCalled)
		assert.True(t, createFutureArtifactCalled)
		assert.True(t, addTagCalled)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_POPULATED, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
	})
}

func TestCatalog_UpdateFuture(t *testing.T) {
	ctx := context.Background()

	t.Run("Overwrite existing future cache", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
		}

		mockClient.On("CreateDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateDatasetRequest) bool {
				assert.True(t, proto.Equal(o.GetDataset().GetId(), datasetID))
				return true
			}),
		).Return(&datacatalog.CreateDatasetResponse{}, nil)

		mockClient.On("UpdateArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.UpdateArtifactRequest) bool {
				assert.True(t, proto.Equal(o.GetDataset(), datasetID))
				assert.IsType(t, &datacatalog.UpdateArtifactRequest_TagName{}, o.GetQueryHandle())
				return true
			}),
		).Return(&datacatalog.UpdateArtifactResponse{ArtifactId: "test-artifact"}, nil)

		taskID := &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Name:         sampleKey.Identifier.GetName(),
				Project:      sampleKey.Identifier.GetProject(),
				Domain:       sampleKey.Identifier.GetDomain(),
				Version:      "version",
			},
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Name:    "wf",
					Project: "p1",
					Domain:  "d1",
				},
				NodeId: "unknown",
			},
			RetryAttempt: 0,
		}

		newKey := sampleKey
		newKey.InputReader = ir
		futureReader := &futureFileReaderMocks.FutureFileReaderInterface{}
		futureReader.On("Read", mock.Anything).Return(&core.DynamicJobSpec{}, nil)

		s, err := discovery.UpdateFuture(ctx, newKey, futureReader, catalog.Metadata{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Name:    taskID.GetNodeExecutionId().GetExecutionId().GetName(),
				Domain:  taskID.GetNodeExecutionId().GetExecutionId().GetDomain(),
				Project: taskID.GetNodeExecutionId().GetExecutionId().GetProject(),
			},
			TaskExecutionIdentifier: &core.TaskExecutionIdentifier{
				TaskId:          &sampleKey.Identifier,
				NodeExecutionId: taskID.GetNodeExecutionId(),
				RetryAttempt:    0,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_POPULATED, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
		assert.Equal(t, "flyte_cached_future-BE6CZsMk6N3ExR_4X9EuwBgj2Jh2UwasXK3a_pM9xlY", s.GetMetadata().GetArtifactTag().GetName())
		sourceTID := s.GetMetadata().GetSourceTaskExecution()
		assert.Equal(t, taskID.GetTaskId().String(), sourceTID.GetTaskId().String())
		assert.Equal(t, taskID.GetRetryAttempt(), sourceTID.GetRetryAttempt())
		assert.Equal(t, taskID.GetNodeExecutionId().String(), sourceTID.GetNodeExecutionId().String())
	})

	t.Run("Overwrite non-existing future execution", func(t *testing.T) {
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
				assert.True(t, proto.Equal(o.GetDataset().GetId(), datasetID))
				createDatasetCalled = true
				return true
			}),
		).Return(&datacatalog.CreateDatasetResponse{}, nil)

		updateArtifactCalled := false
		mockClient.On("UpdateArtifact", ctx, mock.Anything).Run(func(args mock.Arguments) {
			updateArtifactCalled = true
		}).Return(nil, status.New(codes.NotFound, "missing entity of type Artifact with identifier id").Err())

		createArtifactCalled := false
		mockClient.On("CreateFutureArtifact",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateArtifactRequest) bool {
				_, parseErr := uuid.Parse(o.GetArtifact().GetId())
				assert.NoError(t, parseErr)
				assert.True(t, proto.Equal(o.GetArtifact().GetDataset(), datasetID))
				createArtifactCalled = true
				return true
			}),
		).Return(&datacatalog.CreateArtifactResponse{}, nil)

		addTagCalled := false
		mockClient.On("AddTag",
			ctx,
			mock.MatchedBy(func(o *datacatalog.AddTagRequest) bool {
				assert.EqualValues(t, "flyte_cached_future-BE6CZsMk6N3ExR_4X9EuwBgj2Jh2UwasXK3a_pM9xlY", o.GetTag().GetName())
				addTagCalled = true
				return true
			}),
		).Return(&datacatalog.AddTagResponse{}, nil)

		taskID := &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Name:         sampleKey.Identifier.GetName(),
				Project:      sampleKey.Identifier.GetProject(),
				Domain:       sampleKey.Identifier.GetDomain(),
				Version:      "version",
			},
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Name:    "wf",
					Project: "p1",
					Domain:  "d1",
				},
				NodeId: "unknown",
			},
			RetryAttempt: 0,
		}

		newKey := sampleKey
		newKey.InputReader = ir
		futureReader := &futureFileReaderMocks.FutureFileReaderInterface{}
		futureReader.On("Read", mock.Anything).Return(&core.DynamicJobSpec{}, nil)

		s, err := discovery.UpdateFuture(ctx, newKey, futureReader, catalog.Metadata{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Name:    taskID.GetNodeExecutionId().GetExecutionId().GetName(),
				Domain:  taskID.GetNodeExecutionId().GetExecutionId().GetDomain(),
				Project: taskID.GetNodeExecutionId().GetExecutionId().GetProject(),
			},
			TaskExecutionIdentifier: &core.TaskExecutionIdentifier{
				TaskId:          &sampleKey.Identifier,
				NodeExecutionId: taskID.GetNodeExecutionId(),
				RetryAttempt:    0,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_POPULATED, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
		assert.Equal(t, "flyte_cached_future-BE6CZsMk6N3ExR_4X9EuwBgj2Jh2UwasXK3a_pM9xlY", s.GetMetadata().GetArtifactTag().GetName())
		assert.Nil(t, s.GetMetadata().GetSourceTaskExecution())
		assert.True(t, createDatasetCalled)
		assert.True(t, updateArtifactCalled)
		assert.True(t, createArtifactCalled)
		assert.True(t, addTagCalled)
	})

	t.Run("Error when creating dataset", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
		}

		mockClient.On("CreateDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateDatasetRequest) bool {
				return true
			}),
		).Return(&datacatalog.CreateDatasetResponse{}, errors.New("generic error"))

		newKey := sampleKey
		newKey.InputReader = ir
		futureReader := &futureFileReaderMocks.FutureFileReaderInterface{}
		futureReader.On("Read", mock.Anything).Return(&core.DynamicJobSpec{}, nil)

		s, err := discovery.UpdateFuture(ctx, newKey, futureReader, catalog.Metadata{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Name: "test",
			},
			TaskExecutionIdentifier: nil,
		})
		assert.Error(t, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_PUT_FAILURE, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
	})

	t.Run("Error when overwriting execution", func(t *testing.T) {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(sampleParameters, nil, nil)

		mockClient := &mocks.DataCatalogClient{}
		discovery := &CatalogClient{
			client: mockClient,
		}

		mockClient.On("CreateDataset",
			ctx,
			mock.MatchedBy(func(o *datacatalog.CreateDatasetRequest) bool {
				return true
			}),
		).Return(&datacatalog.CreateDatasetResponse{}, nil)

		genericErr := errors.New("generic error")
		mockClient.On("UpdateArtifact", ctx, mock.Anything).Return(nil, genericErr)

		newKey := sampleKey
		newKey.InputReader = ir
		futureReader := &futureFileReaderMocks.FutureFileReaderInterface{}
		futureReader.On("Read", mock.Anything).Return(&core.DynamicJobSpec{}, nil)

		s, err := discovery.UpdateFuture(ctx, newKey, futureReader, catalog.Metadata{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Name: "test",
			},
			TaskExecutionIdentifier: nil,
		})
		assert.Error(t, err)
		assert.Equal(t, genericErr, err)
		assert.Equal(t, core.CatalogCacheStatus_CACHE_PUT_FAILURE, s.GetCacheStatus())
		assert.NotNil(t, s.GetMetadata())
	})
}
