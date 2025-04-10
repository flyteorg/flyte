package impl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/datacatalog/pkg/common"
	"github.com/flyteorg/flyte/datacatalog/pkg/errors"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/mocks"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/transformers"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	mockScope "github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

func init() {
	labeled.SetMetricKeys(contextutils.AppNameKey)
}

func getTestDataset() *datacatalog.Dataset {
	return &datacatalog.Dataset{
		Id: &datacatalog.DatasetID{
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "test-name",
			Version: "test-version",
			UUID:    "test-uuid",
		},
		Metadata: &datacatalog.Metadata{
			KeyMap: map[string]string{"key1": "value1"},
		},
		PartitionKeys: []string{"key1", "key2"},
	}
}

func getDataCatalogRepo() *mocks.DataCatalogRepo {
	return &mocks.DataCatalogRepo{
		MockDatasetRepo: &mocks.DatasetRepo{},
	}
}

func TestCreateDataset(t *testing.T) {

	expectedDataset := getTestDataset()

	t.Run("CreateDatasetWithPartitions", func(t *testing.T) {
		dcRepo := getDataCatalogRepo()
		datasetManager := NewDatasetManager(dcRepo, nil, mockScope.NewTestScope())
		dcRepo.MockDatasetRepo.On("Create",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(dataset models.Dataset) bool {

				return dataset.Name == expectedDataset.GetId().GetName() &&
					dataset.Project == expectedDataset.GetId().GetProject() &&
					dataset.Domain == expectedDataset.GetId().GetDomain() &&
					dataset.Version == expectedDataset.GetId().GetVersion() &&
					len(dataset.PartitionKeys) == len(expectedDataset.GetPartitionKeys()) &&
					dataset.PartitionKeys[0].Name == expectedDataset.GetPartitionKeys()[0] &&
					dataset.PartitionKeys[1].Name == expectedDataset.GetPartitionKeys()[1]
			})).Return(nil)
		request := &datacatalog.CreateDatasetRequest{Dataset: expectedDataset}
		datasetResponse, err := datasetManager.CreateDataset(context.Background(), request)
		assert.NoError(t, err)
		assert.NotNil(t, datasetResponse)
	})

	t.Run("CreateDatasetNoPartitions", func(t *testing.T) {
		dcRepo := getDataCatalogRepo()
		datasetManager := NewDatasetManager(dcRepo, nil, mockScope.NewTestScope())
		dcRepo.MockDatasetRepo.On("Create",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(dataset models.Dataset) bool {

				return dataset.Name == expectedDataset.GetId().GetName() &&
					dataset.Project == expectedDataset.GetId().GetProject() &&
					dataset.Domain == expectedDataset.GetId().GetDomain() &&
					dataset.Version == expectedDataset.GetId().GetVersion() &&
					len(dataset.PartitionKeys) == 0
			})).Return(nil)

		expectedDataset.PartitionKeys = nil
		request := &datacatalog.CreateDatasetRequest{Dataset: expectedDataset}
		datasetResponse, err := datasetManager.CreateDataset(context.Background(), request)
		assert.NoError(t, err)
		assert.NotNil(t, datasetResponse)
	})

	t.Run("MissingInput", func(t *testing.T) {
		dcRepo := getDataCatalogRepo()
		datasetManager := NewDatasetManager(dcRepo, nil, mockScope.NewTestScope())
		request := &datacatalog.CreateDatasetRequest{
			Dataset: &datacatalog.Dataset{
				Id: &datacatalog.DatasetID{
					Domain:  "missing-domain",
					Name:    "missing-name",
					Version: "missing-version",
				},
			},
		}

		_, err := datasetManager.CreateDataset(context.Background(), request)
		assert.Error(t, err)
		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})

	t.Run("AlreadyExists", func(t *testing.T) {
		dcRepo := getDataCatalogRepo()
		datasetManager := NewDatasetManager(dcRepo, nil, mockScope.NewTestScope())

		dcRepo.MockDatasetRepo.On("Create",
			mock.Anything,
			mock.Anything).Return(status.Error(codes.AlreadyExists, "test already exists"))
		request := &datacatalog.CreateDatasetRequest{
			Dataset: getTestDataset(),
		}

		_, err := datasetManager.CreateDataset(context.Background(), request)
		assert.Error(t, err)
		responseCode := status.Code(err)
		assert.Equal(t, codes.AlreadyExists, responseCode)
	})

	t.Run("DuplicatePartition", func(t *testing.T) {
		dcRepo := getDataCatalogRepo()
		badDataset := getTestDataset()
		badDataset.PartitionKeys = append(badDataset.PartitionKeys, badDataset.GetPartitionKeys()[0])
		datasetManager := NewDatasetManager(dcRepo, nil, mockScope.NewTestScope())

		dcRepo.MockDatasetRepo.On("Create",
			mock.Anything,
			mock.Anything).Return(status.Error(codes.AlreadyExists, "test already exists"))
		request := &datacatalog.CreateDatasetRequest{
			Dataset: badDataset,
		}
		_, err := datasetManager.CreateDataset(context.Background(), request)
		assert.Error(t, err)
		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})
}

func TestGetDataset(t *testing.T) {
	expectedDataset := getTestDataset()

	t.Run("HappyPath", func(t *testing.T) {
		dcRepo := getDataCatalogRepo()
		datasetManager := NewDatasetManager(dcRepo, nil, mockScope.NewTestScope())

		datasetModelResponse, err := transformers.CreateDatasetModel(expectedDataset)
		assert.NoError(t, err)

		dcRepo.MockDatasetRepo.On("Get",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(datasetKey models.DatasetKey) bool {

				return datasetKey.Name == expectedDataset.GetId().GetName() &&
					datasetKey.Project == expectedDataset.GetId().GetProject() &&
					datasetKey.Domain == expectedDataset.GetId().GetDomain() &&
					datasetKey.Version == expectedDataset.GetId().GetVersion()
			})).Return(*datasetModelResponse, nil)
		request := &datacatalog.GetDatasetRequest{Dataset: getTestDataset().GetId()}
		datasetResponse, err := datasetManager.GetDataset(context.Background(), request)
		assert.NoError(t, err)
		assert.NotNil(t, datasetResponse)
		assert.True(t, proto.Equal(datasetResponse.GetDataset(), expectedDataset))
		assert.EqualValues(t, datasetResponse.GetDataset().GetMetadata().GetKeyMap(), expectedDataset.GetMetadata().GetKeyMap())
	})

	t.Run("Does not exist", func(t *testing.T) {
		dcRepo := getDataCatalogRepo()
		datasetManager := NewDatasetManager(dcRepo, nil, mockScope.NewTestScope())

		dcRepo.MockDatasetRepo.On("Get",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(datasetKey models.DatasetKey) bool {

				return datasetKey.Name == expectedDataset.GetId().GetName() &&
					datasetKey.Project == expectedDataset.GetId().GetProject() &&
					datasetKey.Domain == expectedDataset.GetId().GetDomain() &&
					datasetKey.Version == expectedDataset.GetId().GetVersion()
			})).Return(models.Dataset{}, errors.NewDataCatalogError(codes.NotFound, "dataset does not exist"))
		request := &datacatalog.GetDatasetRequest{Dataset: getTestDataset().GetId()}
		_, err := datasetManager.GetDataset(context.Background(), request)
		assert.Error(t, err)
		responseCode := status.Code(err)
		assert.Equal(t, codes.NotFound, responseCode)
	})

}

func TestListDatasets(t *testing.T) {
	ctx := context.Background()
	expectedDataset := getTestDataset()
	dcRepo := getDataCatalogRepo()

	t.Run("List Datasets on invalid filter", func(t *testing.T) {
		datasetManager := NewDatasetManager(dcRepo, nil, mockScope.NewTestScope())
		filter := &datacatalog.FilterExpression{
			Filters: []*datacatalog.SinglePropertyFilter{
				{
					PropertyFilter: &datacatalog.SinglePropertyFilter_ArtifactFilter{
						ArtifactFilter: &datacatalog.ArtifactPropertyFilter{
							Property: &datacatalog.ArtifactPropertyFilter_ArtifactId{
								ArtifactId: "test",
							},
						},
					},
				},
			},
		}

		artifactResponse, err := datasetManager.ListDatasets(ctx, &datacatalog.ListDatasetsRequest{Filter: filter})
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)
		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})

	t.Run("List Datasets with Project and Name", func(t *testing.T) {
		datasetManager := NewDatasetManager(dcRepo, nil, mockScope.NewTestScope())
		filter := &datacatalog.FilterExpression{
			Filters: []*datacatalog.SinglePropertyFilter{
				{
					PropertyFilter: &datacatalog.SinglePropertyFilter_DatasetFilter{
						DatasetFilter: &datacatalog.DatasetPropertyFilter{
							Property: &datacatalog.DatasetPropertyFilter_Project{
								Project: "testProject",
							},
						},
					},
				},
				{
					PropertyFilter: &datacatalog.SinglePropertyFilter_DatasetFilter{
						DatasetFilter: &datacatalog.DatasetPropertyFilter{
							Property: &datacatalog.DatasetPropertyFilter_Domain{
								Domain: "testDomain",
							},
						},
					},
				},
			},
		}

		datasetModel, err := transformers.CreateDatasetModel(expectedDataset)
		assert.NoError(t, err)

		dcRepo.MockDatasetRepo.On("List", mock.Anything,
			mock.MatchedBy(func(listInput models.ListModelsInput) bool {
				return len(listInput.ModelFilters) == 2 &&
					listInput.ModelFilters[0].Entity == common.Dataset &&
					len(listInput.ModelFilters[0].ValueFilters) == 1 &&
					listInput.ModelFilters[1].Entity == common.Dataset &&
					len(listInput.ModelFilters[1].ValueFilters) == 1 &&
					listInput.Limit == 50 &&
					listInput.Offset == 0
			})).Return([]models.Dataset{*datasetModel}, nil)

		datasetResponse, err := datasetManager.ListDatasets(ctx, &datacatalog.ListDatasetsRequest{Filter: filter})
		assert.NoError(t, err)
		assert.NotEmpty(t, datasetResponse)
		assert.Len(t, datasetResponse.GetDatasets(), 1)
	})

	t.Run("List Datasets with no filtering", func(t *testing.T) {
		datasetManager := NewDatasetManager(dcRepo, nil, mockScope.NewTestScope())

		datasetModel, err := transformers.CreateDatasetModel(expectedDataset)
		assert.NoError(t, err)

		dcRepo.MockDatasetRepo.On("List", mock.Anything,
			mock.MatchedBy(func(listInput models.ListModelsInput) bool {
				return len(listInput.ModelFilters) == 0 &&
					listInput.Limit == 50 &&
					listInput.Offset == 0
			})).Return([]models.Dataset{*datasetModel}, nil)

		datasetResponse, err := datasetManager.ListDatasets(ctx, &datacatalog.ListDatasetsRequest{})
		assert.NoError(t, err)
		assert.NotEmpty(t, datasetResponse)
		assert.Len(t, datasetResponse.GetDatasets(), 1)
	})
}
