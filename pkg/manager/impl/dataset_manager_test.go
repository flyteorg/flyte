package impl

import (
	"testing"

	"context"

	"github.com/golang/protobuf/proto"
	"github.com/lyft/datacatalog/pkg/errors"
	"github.com/lyft/datacatalog/pkg/repositories/mocks"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/lyft/flytestdlib/contextutils"
	mockScope "github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		},
		Metadata: &datacatalog.Metadata{
			KeyMap: map[string]string{"key1": "value1"},
		},
	}
}

func getDataCatalogRepo() *mocks.DataCatalogRepo {
	return &mocks.DataCatalogRepo{
		MockDatasetRepo: &mocks.DatasetRepo{},
	}
}

func TestCreateDataset(t *testing.T) {

	expectedDataset := getTestDataset()

	t.Run("HappyPath", func(t *testing.T) {
		dcRepo := getDataCatalogRepo()
		datasetManager := NewDatasetManager(dcRepo, nil, mockScope.NewTestScope())
		dcRepo.MockDatasetRepo.On("Create",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(dataset models.Dataset) bool {

				return dataset.Name == expectedDataset.Id.Name &&
					dataset.Project == expectedDataset.Id.Project &&
					dataset.Domain == expectedDataset.Id.Domain &&
					dataset.Version == expectedDataset.Id.Version
			})).Return(nil)
		request := datacatalog.CreateDatasetRequest{Dataset: expectedDataset}
		datasetResponse, err := datasetManager.CreateDataset(context.Background(), request)
		assert.NoError(t, err)
		assert.NotNil(t, datasetResponse)
	})

	t.Run("MissingInput", func(t *testing.T) {
		dcRepo := getDataCatalogRepo()
		datasetManager := NewDatasetManager(dcRepo, nil, mockScope.NewTestScope())
		request := datacatalog.CreateDatasetRequest{
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
		request := datacatalog.CreateDatasetRequest{
			Dataset: getTestDataset(),
		}

		_, err := datasetManager.CreateDataset(context.Background(), request)
		assert.Error(t, err)
		responseCode := status.Code(err)
		assert.Equal(t, codes.AlreadyExists, responseCode)
	})
}

func TestGetDataset(t *testing.T) {
	expectedDataset := getTestDataset()

	t.Run("HappyPath", func(t *testing.T) {
		dcRepo := getDataCatalogRepo()
		datasetManager := NewDatasetManager(dcRepo, nil, mockScope.NewTestScope())

		serializedMetadata, _ := proto.Marshal(expectedDataset.Metadata)
		datasetModelResponse := models.Dataset{
			DatasetKey: models.DatasetKey{
				Project: expectedDataset.Id.Project,
				Domain:  expectedDataset.Id.Domain,
				Version: expectedDataset.Id.Version,
				Name:    expectedDataset.Id.Name,
			},
			SerializedMetadata: serializedMetadata,
		}

		dcRepo.MockDatasetRepo.On("Get",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(datasetKey models.DatasetKey) bool {

				return datasetKey.Name == expectedDataset.Id.Name &&
					datasetKey.Project == expectedDataset.Id.Project &&
					datasetKey.Domain == expectedDataset.Id.Domain &&
					datasetKey.Version == expectedDataset.Id.Version
			})).Return(datasetModelResponse, nil)
		request := datacatalog.GetDatasetRequest{Dataset: getTestDataset().Id}
		datasetResponse, err := datasetManager.GetDataset(context.Background(), request)
		assert.NoError(t, err)
		assert.NotNil(t, datasetResponse)
		assert.True(t, proto.Equal(datasetResponse.Dataset, expectedDataset))
		assert.EqualValues(t, datasetResponse.Dataset.Metadata.KeyMap, expectedDataset.Metadata.KeyMap)
	})

	t.Run("Does not exist", func(t *testing.T) {
		dcRepo := getDataCatalogRepo()
		datasetManager := NewDatasetManager(dcRepo, nil, mockScope.NewTestScope())

		dcRepo.MockDatasetRepo.On("Get",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(datasetKey models.DatasetKey) bool {

				return datasetKey.Name == expectedDataset.Id.Name &&
					datasetKey.Project == expectedDataset.Id.Project &&
					datasetKey.Domain == expectedDataset.Id.Domain &&
					datasetKey.Version == expectedDataset.Id.Version
			})).Return(models.Dataset{}, errors.NewDataCatalogError(codes.NotFound, "dataset does not exist"))
		request := datacatalog.GetDatasetRequest{Dataset: getTestDataset().Id}
		_, err := datasetManager.GetDataset(context.Background(), request)
		assert.Error(t, err)
		responseCode := status.Code(err)
		assert.Equal(t, codes.NotFound, responseCode)
	})

}
