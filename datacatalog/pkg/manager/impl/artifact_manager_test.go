package impl

import (
	"context"
	stdErrors "errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/ptypes"

	"github.com/flyteorg/flyte/datacatalog/pkg/common"
	"github.com/flyteorg/flyte/datacatalog/pkg/errors"
	repoErrors "github.com/flyteorg/flyte/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/mocks"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/transformers"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	mockScope "github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func init() {
	labeled.SetMetricKeys(contextutils.AppNameKey)
}

func createInmemoryDataStore(t testing.TB, scope mockScope.Scope) *storage.DataStore {
	cfg := storage.Config{
		Type: storage.TypeMemory,
	}
	d, err := storage.NewDataStore(&cfg, scope)
	assert.NoError(t, err)
	return d
}

func getTestStringLiteral() *core.Literal {
	return getTestStringLiteralWithValue("value1")
}

func getTestStringLiteralWithValue(val string) *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: val}},
				},
			},
		},
	}
}

func getTestTimestamp() time.Time {
	timestamp, _ := time.Parse(time.RFC3339, "2019-12-26T00:00:00+00:00")
	return timestamp
}

func getTestArtifact() *datacatalog.Artifact {
	datasetID := &datacatalog.DatasetID{
		Project: "test-project",
		Domain:  "test-domain",
		Name:    "test-name",
		Version: "test-version",
		UUID:    "test-uuid",
	}
	createdAt, _ := timestamppb.New(getTestTimestamp())

	return &datacatalog.Artifact{
		Id:      "test-id",
		Dataset: datasetID,
		Metadata: &datacatalog.Metadata{
			KeyMap: map[string]string{"key1": "value1"},
		},
		Data: []*datacatalog.ArtifactData{
			{
				Name:  "data1",
				Value: getTestStringLiteral(),
			},
		},
		Partitions: []*datacatalog.Partition{
			{Key: "key1", Value: "value1"},
			{Key: "key2", Value: "value2"},
		},
		Tags: []*datacatalog.Tag{
			{Name: "test-tag", Dataset: datasetID, ArtifactId: "test-id"},
		},
		CreatedAt: createdAt,
	}
}

func newMockDataCatalogRepo() *mocks.DataCatalogRepo {
	return &mocks.DataCatalogRepo{
		MockDatasetRepo:     &mocks.DatasetRepo{},
		MockArtifactRepo:    &mocks.ArtifactRepo{},
		MockReservationRepo: &mocks.ReservationRepo{},
		MockTagRepo:         &mocks.TagRepo{},
	}
}

func getExpectedDatastoreLocation(ctx context.Context, store *storage.DataStore, prefix storage.DataReference, artifact *datacatalog.Artifact, idx int) (storage.DataReference, error) {
	return getExpectedDatastoreLocationFromName(ctx, store, prefix, artifact, artifact.GetData()[idx].GetName())
}

func getExpectedDatastoreLocationFromName(ctx context.Context, store *storage.DataStore, prefix storage.DataReference, artifact *datacatalog.Artifact, artifactDataName string) (storage.DataReference, error) {
	dataset := artifact.GetDataset()
	return store.ConstructReference(ctx, prefix, dataset.GetProject(), dataset.GetDomain(), dataset.GetName(), dataset.GetVersion(), artifact.GetId(), artifactDataName, artifactDataFile)
}

func getExpectedArtifactModel(ctx context.Context, t *testing.T, datastore *storage.DataStore, artifact *datacatalog.Artifact) models.Artifact {
	expectedDataset := artifact.GetDataset()

	artifactData := make([]models.ArtifactData, len(artifact.GetData()))
	// Write sample artifact data to the expected location and see if the retrieved data matches
	for i := range artifact.GetData() {
		testStoragePrefix, err := datastore.ConstructReference(ctx, datastore.GetBaseContainerFQN(ctx), "test")
		assert.NoError(t, err)
		dataLocation, err := getExpectedDatastoreLocation(ctx, datastore, testStoragePrefix, artifact, i)
		assert.NoError(t, err)
		err = datastore.WriteProtobuf(ctx, dataLocation, storage.Options{}, artifact.GetData()[i].GetValue())
		assert.NoError(t, err)

		artifactData[i].Name = artifact.GetData()[i].GetName()
		artifactData[i].Location = dataLocation.String()
	}

	// construct the artifact model we will return on the queries
	serializedMetadata, err := proto.Marshal(artifact.GetMetadata())
	assert.NoError(t, err)
	datasetKey := models.DatasetKey{
		Project: expectedDataset.GetProject(),
		Domain:  expectedDataset.GetDomain(),
		Version: expectedDataset.GetVersion(),
		Name:    expectedDataset.GetName(),
		UUID:    expectedDataset.GetUUID(),
	}
	return models.Artifact{
		ArtifactKey: models.ArtifactKey{
			DatasetProject: expectedDataset.GetProject(),
			DatasetDomain:  expectedDataset.GetDomain(),
			DatasetVersion: expectedDataset.GetVersion(),
			DatasetName:    expectedDataset.GetName(),
			ArtifactID:     artifact.GetId(),
		},
		DatasetUUID:  expectedDataset.GetUUID(),
		ArtifactData: artifactData,
		Dataset: models.Dataset{
			DatasetKey:         datasetKey,
			SerializedMetadata: serializedMetadata,
		},
		SerializedMetadata: serializedMetadata,
		Partitions: []models.Partition{
			{Key: "key1", Value: "value1"},
			{Key: "key2", Value: "value2"},
		},
		Tags: []models.Tag{
			{TagKey: models.TagKey{TagName: "test-tag"}, DatasetUUID: expectedDataset.GetUUID(), ArtifactID: artifact.GetId()},
		},
		BaseModel: models.BaseModel{
			CreatedAt: getTestTimestamp(),
		},
	}
}

func TestCreateArtifact(t *testing.T) {
	ctx := context.Background()
	datastore := createInmemoryDataStore(t, mockScope.NewTestScope())
	testStoragePrefix, err := datastore.ConstructReference(ctx, datastore.GetBaseContainerFQN(ctx), "test")
	assert.NoError(t, err)

	// Mock dataset to return for artifact lookups
	expectedDataset := getTestDataset()
	mockDatasetModel := models.Dataset{
		DatasetKey: models.DatasetKey{
			Project: expectedDataset.GetId().GetProject(),
			Domain:  expectedDataset.GetId().GetDomain(),
			Name:    expectedDataset.GetId().GetName(),
			Version: expectedDataset.GetId().GetVersion(),
			UUID:    expectedDataset.GetId().GetUUID(),
		},
		PartitionKeys: []models.PartitionKey{
			{Name: expectedDataset.GetPartitionKeys()[0]},
			{Name: expectedDataset.GetPartitionKeys()[1]},
		},
	}

	t.Run("HappyPath", func(t *testing.T) {
		datastore := createInmemoryDataStore(t, mockScope.NewTestScope())
		expectedDataset := getTestDataset()

		ctx := context.Background()
		dcRepo := newMockDataCatalogRepo()
		dcRepo.MockDatasetRepo.On("Get", mock.Anything,
			mock.MatchedBy(func(dataset models.DatasetKey) bool {
				return dataset.Project == expectedDataset.GetId().GetProject() &&
					dataset.Domain == expectedDataset.GetId().GetDomain() &&
					dataset.Name == expectedDataset.GetId().GetName() &&
					dataset.Version == expectedDataset.GetId().GetVersion()
			})).Return(mockDatasetModel, nil)

		dcRepo.MockArtifactRepo.On("Create",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(artifact models.Artifact) bool {
				expectedArtifact := getTestArtifact()
				return artifact.ArtifactID == expectedArtifact.GetId() &&
					artifact.SerializedMetadata != nil &&
					len(artifact.ArtifactData) == len(expectedArtifact.GetData()) &&
					artifact.ArtifactKey.DatasetProject == expectedArtifact.GetDataset().GetProject() &&
					artifact.ArtifactKey.DatasetDomain == expectedArtifact.GetDataset().GetDomain() &&
					artifact.ArtifactKey.DatasetName == expectedArtifact.GetDataset().GetName() &&
					artifact.ArtifactKey.DatasetVersion == expectedArtifact.GetDataset().GetVersion() &&
					artifact.DatasetUUID == expectedArtifact.GetDataset().GetUUID() &&
					artifact.Partitions[0].Key == expectedArtifact.GetPartitions()[0].GetKey() &&
					artifact.Partitions[0].Value == expectedArtifact.GetPartitions()[0].GetValue() &&
					artifact.Partitions[0].DatasetUUID == expectedDataset.GetId().GetUUID() &&
					artifact.Partitions[1].Key == expectedArtifact.GetPartitions()[1].GetKey() &&
					artifact.Partitions[1].Value == expectedArtifact.GetPartitions()[1].GetValue() &&
					artifact.Partitions[1].DatasetUUID == expectedDataset.GetId().GetUUID()
			})).Return(nil)

		request := &datacatalog.CreateArtifactRequest{Artifact: getTestArtifact()}
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.CreateArtifact(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, artifactResponse)

		// check that the datastore has the artifactData
		dataRef, err := getExpectedDatastoreLocation(ctx, datastore, testStoragePrefix, getTestArtifact(), 0)
		assert.NoError(t, err)
		var value core.Literal
		err = datastore.ReadProtobuf(ctx, dataRef, &value)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(&value, getTestArtifact().GetData()[0].GetValue()))
	})

	t.Run("Dataset does not exist", func(t *testing.T) {
		dcRepo := newMockDataCatalogRepo()
		dcRepo.MockDatasetRepo.On("Get", mock.Anything, mock.Anything).Return(models.Dataset{}, status.Error(codes.NotFound, "not found"))

		request := &datacatalog.CreateArtifactRequest{Artifact: getTestArtifact()}
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.CreateArtifact(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)
		responseCode := status.Code(err)
		assert.Equal(t, codes.NotFound, responseCode)
	})

	t.Run("Artifact missing ID", func(t *testing.T) {
		request := &datacatalog.CreateArtifactRequest{
			Artifact: &datacatalog.Artifact{
				// missing artifact id
				Dataset: getTestDataset().GetId(),
			},
		}

		artifactManager := NewArtifactManager(&mocks.DataCatalogRepo{}, createInmemoryDataStore(t, mockScope.NewTestScope()), testStoragePrefix, mockScope.NewTestScope())
		_, err := artifactManager.CreateArtifact(ctx, request)
		assert.Error(t, err)
		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})

	t.Run("Artifact missing artifact data", func(t *testing.T) {
		request := &datacatalog.CreateArtifactRequest{
			Artifact: &datacatalog.Artifact{
				Id:      "test",
				Dataset: getTestDataset().GetId(),
				// missing artifactData
			},
		}

		artifactManager := NewArtifactManager(&mocks.DataCatalogRepo{}, datastore, testStoragePrefix, mockScope.NewTestScope())
		_, err := artifactManager.CreateArtifact(ctx, request)
		assert.Error(t, err)
		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})

	t.Run("Already exists", func(t *testing.T) {
		dcRepo := newMockDataCatalogRepo()

		dcRepo.MockDatasetRepo.On("Get", mock.Anything, mock.Anything).Return(mockDatasetModel, nil)

		dcRepo.MockArtifactRepo.On("Create",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(artifact models.Artifact) bool {
				expectedArtifact := getTestArtifact()
				return artifact.ArtifactID == expectedArtifact.GetId() &&
					artifact.SerializedMetadata != nil &&
					len(artifact.ArtifactData) == len(expectedArtifact.GetData()) &&
					artifact.ArtifactKey.DatasetProject == expectedArtifact.GetDataset().GetProject() &&
					artifact.ArtifactKey.DatasetDomain == expectedArtifact.GetDataset().GetDomain() &&
					artifact.ArtifactKey.DatasetName == expectedArtifact.GetDataset().GetName() &&
					artifact.ArtifactKey.DatasetVersion == expectedArtifact.GetDataset().GetVersion()
			})).Return(status.Error(codes.AlreadyExists, "test already exists"))

		request := &datacatalog.CreateArtifactRequest{Artifact: getTestArtifact()}
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.CreateArtifact(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)

		responseCode := status.Code(err)
		assert.Equal(t, codes.AlreadyExists, responseCode)
	})

	t.Run("Missing Partitions", func(t *testing.T) {
		dcRepo := newMockDataCatalogRepo()
		dcRepo.MockDatasetRepo.On("Get", mock.Anything, mock.Anything).Return(mockDatasetModel, nil)
		artifact := getTestArtifact()
		artifact.Partitions = nil
		dcRepo.MockArtifactRepo.On("Create",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(artifact models.Artifact) bool {
				return false
			})).Return(fmt.Errorf("Validation should happen before this happens"))

		request := &datacatalog.CreateArtifactRequest{Artifact: artifact}
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.CreateArtifact(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)

		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})

	t.Run("No Partitions", func(t *testing.T) {
		dcRepo := newMockDataCatalogRepo()
		mockDatasetModel := models.Dataset{
			DatasetKey: models.DatasetKey{
				Project: expectedDataset.GetId().GetProject(),
				Domain:  expectedDataset.GetId().GetDomain(),
				Name:    expectedDataset.GetId().GetName(),
				Version: expectedDataset.GetId().GetVersion(),
			},
		}
		dcRepo.MockDatasetRepo.On("Get", mock.Anything, mock.Anything).Return(mockDatasetModel, nil)
		artifact := getTestArtifact()
		artifact.Partitions = []*datacatalog.Partition{}
		dcRepo.MockArtifactRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		request := &datacatalog.CreateArtifactRequest{Artifact: artifact}
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		_, err := artifactManager.CreateArtifact(ctx, request)
		assert.NoError(t, err)
	})

	t.Run("Invalid Partition", func(t *testing.T) {
		dcRepo := newMockDataCatalogRepo()
		dcRepo.MockDatasetRepo.On("Get", mock.Anything, mock.Anything).Return(mockDatasetModel, nil)
		artifact := getTestArtifact()
		artifact.Partitions = append(artifact.Partitions, &datacatalog.Partition{Key: "invalidKey", Value: "invalid"})
		dcRepo.MockArtifactRepo.On("Create",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(artifact models.Artifact) bool {
				return false
			})).Return(fmt.Errorf("Validation should happen before this happens"))

		request := &datacatalog.CreateArtifactRequest{Artifact: artifact}
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.CreateArtifact(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)

		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})

}

func TestGetArtifact(t *testing.T) {
	ctx := context.Background()
	datastore := createInmemoryDataStore(t, mockScope.NewTestScope())
	testStoragePrefix, err := datastore.ConstructReference(ctx, datastore.GetBaseContainerFQN(ctx), "test")
	assert.NoError(t, err)

	dcRepo := newMockDataCatalogRepo()

	expectedArtifact := getTestArtifact()
	mockArtifactModel := getExpectedArtifactModel(ctx, t, datastore, expectedArtifact)

	t.Run("Get by Id", func(t *testing.T) {
		dcRepo.MockArtifactRepo.On("Get", mock.Anything,
			mock.MatchedBy(func(artifactKey models.ArtifactKey) bool {
				return artifactKey.ArtifactID == expectedArtifact.GetId() &&
					artifactKey.DatasetProject == expectedArtifact.GetDataset().GetProject() &&
					artifactKey.DatasetDomain == expectedArtifact.GetDataset().GetDomain() &&
					artifactKey.DatasetVersion == expectedArtifact.GetDataset().GetVersion() &&
					artifactKey.DatasetName == expectedArtifact.GetDataset().GetName()
			})).Return(mockArtifactModel, nil)

		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.GetArtifact(ctx, &datacatalog.GetArtifactRequest{
			Dataset:     getTestDataset().GetId(),
			QueryHandle: &datacatalog.GetArtifactRequest_ArtifactId{ArtifactId: expectedArtifact.GetId()},
		})
		assert.NoError(t, err)

		assert.True(t, proto.Equal(expectedArtifact, artifactResponse.GetArtifact()))
	})

	t.Run("Get by Artifact Tag", func(t *testing.T) {
		expectedTag := getTestTag()

		dcRepo.MockTagRepo.On("Get", mock.Anything,
			mock.MatchedBy(func(tag models.TagKey) bool {
				return tag.TagName == expectedTag.TagName &&
					tag.DatasetProject == expectedTag.DatasetProject &&
					tag.DatasetDomain == expectedTag.DatasetDomain &&
					tag.DatasetVersion == expectedTag.DatasetVersion &&
					tag.DatasetName == expectedTag.DatasetName
			})).Return(models.Tag{
			TagKey: models.TagKey{
				DatasetProject: expectedTag.DatasetProject,
				DatasetDomain:  expectedTag.DatasetDomain,
				DatasetName:    expectedTag.DatasetName,
				DatasetVersion: expectedTag.DatasetVersion,
				TagName:        expectedTag.TagName,
			},
			DatasetUUID: expectedTag.DatasetUUID,
			Artifact:    mockArtifactModel,
			ArtifactID:  mockArtifactModel.ArtifactID,
		}, nil)

		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.GetArtifact(ctx, &datacatalog.GetArtifactRequest{
			Dataset:     getTestDataset().GetId(),
			QueryHandle: &datacatalog.GetArtifactRequest_TagName{TagName: expectedTag.TagName},
		})
		assert.NoError(t, err)
		assert.True(t, proto.Equal(expectedArtifact, artifactResponse.GetArtifact()))
	})

	t.Run("Get missing input", func(t *testing.T) {
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.GetArtifact(ctx, &datacatalog.GetArtifactRequest{Dataset: getTestDataset().GetId()})
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)
		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})

	t.Run("Get does not exist", func(t *testing.T) {
		dcRepo.MockTagRepo.On("Get", mock.Anything, mock.Anything).Return(
			models.Tag{}, errors.NewDataCatalogError(codes.NotFound, "tag with artifact does not exist"))
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.GetArtifact(ctx, &datacatalog.GetArtifactRequest{Dataset: getTestDataset().GetId(), QueryHandle: &datacatalog.GetArtifactRequest_TagName{TagName: "test"}})
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)
		responseCode := status.Code(err)
		assert.Equal(t, codes.NotFound, responseCode)
	})
}

func TestListArtifact(t *testing.T) {
	ctx := context.Background()
	datastore := createInmemoryDataStore(t, mockScope.NewTestScope())
	testStoragePrefix, err := datastore.ConstructReference(ctx, datastore.GetBaseContainerFQN(ctx), "test")
	assert.NoError(t, err)

	dcRepo := newMockDataCatalogRepo()

	expectedDataset := getTestDataset()
	mockDatasetModel := models.Dataset{
		DatasetKey: models.DatasetKey{
			Project: expectedDataset.GetId().GetProject(),
			Domain:  expectedDataset.GetId().GetDomain(),
			Name:    expectedDataset.GetId().GetName(),
			Version: expectedDataset.GetId().GetVersion(),
			UUID:    expectedDataset.GetId().GetUUID(),
		},
	}

	expectedArtifact := getTestArtifact()
	mockArtifactModel := getExpectedArtifactModel(ctx, t, datastore, expectedArtifact)

	t.Run("List Artifact on invalid filter", func(t *testing.T) {
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		filter := &datacatalog.FilterExpression{
			Filters: []*datacatalog.SinglePropertyFilter{
				{
					PropertyFilter: &datacatalog.SinglePropertyFilter_DatasetFilter{
						DatasetFilter: &datacatalog.DatasetPropertyFilter{
							Property: &datacatalog.DatasetPropertyFilter_Project{
								Project: "test",
							},
						},
					},
				},
			},
		}

		artifactResponse, err := artifactManager.ListArtifacts(ctx, &datacatalog.ListArtifactsRequest{Dataset: getTestDataset().GetId(), Filter: filter})
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)
		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})

	t.Run("List Artifacts with Partition and Tag", func(t *testing.T) {
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		filter := &datacatalog.FilterExpression{
			Filters: []*datacatalog.SinglePropertyFilter{
				{
					PropertyFilter: &datacatalog.SinglePropertyFilter_PartitionFilter{
						PartitionFilter: &datacatalog.PartitionPropertyFilter{
							Property: &datacatalog.PartitionPropertyFilter_KeyVal{
								KeyVal: &datacatalog.KeyValuePair{Key: "key1", Value: "val1"},
							},
						},
					},
				},
				{
					PropertyFilter: &datacatalog.SinglePropertyFilter_PartitionFilter{
						PartitionFilter: &datacatalog.PartitionPropertyFilter{
							Property: &datacatalog.PartitionPropertyFilter_KeyVal{
								KeyVal: &datacatalog.KeyValuePair{Key: "key2", Value: "val2"},
							},
						},
					},
				},
				{
					PropertyFilter: &datacatalog.SinglePropertyFilter_TagFilter{
						TagFilter: &datacatalog.TagPropertyFilter{
							Property: &datacatalog.TagPropertyFilter_TagName{
								TagName: "special",
							},
						},
					},
				},
			},
		}

		dcRepo.MockDatasetRepo.On("Get", mock.Anything,
			mock.MatchedBy(func(dataset models.DatasetKey) bool {
				return dataset.Project == expectedDataset.GetId().GetProject() &&
					dataset.Domain == expectedDataset.GetId().GetDomain() &&
					dataset.Name == expectedDataset.GetId().GetName() &&
					dataset.Version == expectedDataset.GetId().GetVersion()
			})).Return(mockDatasetModel, nil)

		mockArtifacts := []models.Artifact{
			mockArtifactModel,
			mockArtifactModel,
		}

		dcRepo.MockArtifactRepo.On("List", mock.Anything,
			mock.MatchedBy(func(dataset models.DatasetKey) bool {
				return dataset.Project == expectedDataset.GetId().GetProject() &&
					dataset.Domain == expectedDataset.GetId().GetDomain() &&
					dataset.Name == expectedDataset.GetId().GetName() &&
					dataset.Version == expectedDataset.GetId().GetVersion()
			}),
			mock.MatchedBy(func(listInput models.ListModelsInput) bool {
				return len(listInput.ModelFilters) == 3 &&
					listInput.ModelFilters[0].Entity == common.Partition &&
					len(listInput.ModelFilters[0].ValueFilters) == 2 &&
					listInput.ModelFilters[1].Entity == common.Partition &&
					len(listInput.ModelFilters[1].ValueFilters) == 2 &&
					listInput.ModelFilters[2].Entity == common.Tag &&
					len(listInput.ModelFilters[2].ValueFilters) == 1 &&
					listInput.Limit == 50 &&
					listInput.Offset == 0
			})).Return(mockArtifacts, nil)

		artifactResponse, err := artifactManager.ListArtifacts(ctx, &datacatalog.ListArtifactsRequest{Dataset: expectedDataset.GetId(), Filter: filter})
		assert.NoError(t, err)
		assert.NotEmpty(t, artifactResponse)
	})

	t.Run("List Artifacts with No Partition", func(t *testing.T) {
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		filter := &datacatalog.FilterExpression{Filters: nil}

		dcRepo.MockDatasetRepo.On("Get", mock.Anything,
			mock.MatchedBy(func(dataset models.DatasetKey) bool {
				return dataset.Project == expectedDataset.GetId().GetProject() &&
					dataset.Domain == expectedDataset.GetId().GetDomain() &&
					dataset.Name == expectedDataset.GetId().GetName() &&
					dataset.Version == expectedDataset.GetId().GetVersion()
			})).Return(mockDatasetModel, nil)

		mockArtifacts := []models.Artifact{
			mockArtifactModel,
			mockArtifactModel,
		}
		dcRepo.MockArtifactRepo.On("List", mock.Anything,
			mock.MatchedBy(func(dataset models.DatasetKey) bool {
				return dataset.Project == expectedDataset.GetId().GetProject() &&
					dataset.Domain == expectedDataset.GetId().GetDomain() &&
					dataset.Name == expectedDataset.GetId().GetName() &&
					dataset.Version == expectedDataset.GetId().GetVersion()
			}),
			mock.MatchedBy(func(listInput models.ListModelsInput) bool {
				return len(listInput.ModelFilters) == 0
			})).Return(mockArtifacts, nil)

		artifactResponse, err := artifactManager.ListArtifacts(ctx, &datacatalog.ListArtifactsRequest{Dataset: expectedDataset.GetId(), Filter: filter})
		assert.NoError(t, err)
		assert.NotEmpty(t, artifactResponse)
	})
}

func TestUpdateArtifact(t *testing.T) {
	ctx := context.Background()
	datastore := createInmemoryDataStore(t, mockScope.NewTestScope())
	testStoragePrefix, err := datastore.ConstructReference(ctx, datastore.GetBaseContainerFQN(ctx), "test")
	assert.NoError(t, err)

	expectedDataset := getTestDataset()
	expectedArtifact := getTestArtifact()
	expectedArtifact.Data = append(expectedArtifact.Data, &datacatalog.ArtifactData{
		Name:  "data2",
		Value: getTestStringLiteralWithValue("value2"),
	})
	expectedTag := getTestTag()

	t.Run("Update by ID", func(t *testing.T) {
		ctx := context.Background()
		datastore := createInmemoryDataStore(t, mockScope.NewTestScope())
		mockArtifactModel := getExpectedArtifactModel(ctx, t, datastore, expectedArtifact)

		dcRepo := newMockDataCatalogRepo()
		dcRepo.MockArtifactRepo.On("Get",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(artifactKey models.ArtifactKey) bool {
				return artifactKey.ArtifactID == expectedArtifact.GetId() &&
					artifactKey.DatasetProject == expectedArtifact.GetDataset().GetProject() &&
					artifactKey.DatasetDomain == expectedArtifact.GetDataset().GetDomain() &&
					artifactKey.DatasetName == expectedArtifact.GetDataset().GetName() &&
					artifactKey.DatasetVersion == expectedArtifact.GetDataset().GetVersion()
			})).Return(mockArtifactModel, nil)

		metaData := &datacatalog.Metadata{
			KeyMap: map[string]string{"key2": "value2"},
		}
		serializedMetadata, err := transformers.SerializedMetadata(metaData)
		assert.NoError(t, err)

		dcRepo.MockArtifactRepo.On("Update",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(artifact models.Artifact) bool {
				return artifact.ArtifactID == expectedArtifact.GetId() &&
					artifact.ArtifactKey.DatasetProject == expectedArtifact.GetDataset().GetProject() &&
					artifact.ArtifactKey.DatasetDomain == expectedArtifact.GetDataset().GetDomain() &&
					artifact.ArtifactKey.DatasetName == expectedArtifact.GetDataset().GetName() &&
					artifact.ArtifactKey.DatasetVersion == expectedArtifact.GetDataset().GetVersion() &&
					reflect.DeepEqual(artifact.SerializedMetadata, serializedMetadata)
			})).Return(nil)

		request := &datacatalog.UpdateArtifactRequest{
			Dataset: expectedDataset.GetId(),
			QueryHandle: &datacatalog.UpdateArtifactRequest_ArtifactId{
				ArtifactId: expectedArtifact.GetId(),
			},
			Data: []*datacatalog.ArtifactData{
				{
					Name:  "data1",
					Value: getTestStringLiteralWithValue("value11"),
				},
				{
					Name:  "data3",
					Value: getTestStringLiteralWithValue("value3"),
				},
			},
			Metadata: &datacatalog.Metadata{
				KeyMap: map[string]string{"key2": "value2"},
			},
		}

		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.UpdateArtifact(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, artifactResponse)
		assert.Equal(t, expectedArtifact.GetId(), artifactResponse.GetArtifactId())
		dcRepo.MockArtifactRepo.AssertExpectations(t)

		// check that the datastore has the updated artifactData available
		// data1 should contain updated value
		dataRef, err := getExpectedDatastoreLocationFromName(ctx, datastore, testStoragePrefix, expectedArtifact, "data1")
		assert.NoError(t, err)
		var value core.Literal
		err = datastore.ReadProtobuf(ctx, dataRef, &value)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(&value, getTestStringLiteralWithValue("value11")))

		// data2 was not included in update payload, should be removed
		dataRef, err = getExpectedDatastoreLocationFromName(ctx, datastore, testStoragePrefix, expectedArtifact, "data2")
		assert.NoError(t, err)
		err = datastore.ReadProtobuf(ctx, dataRef, &value)
		assert.Error(t, err)
		assert.True(t, stdErrors.Is(err, os.ErrNotExist))

		// data3 did not exist before, should be present after update
		dataRef, err = getExpectedDatastoreLocationFromName(ctx, datastore, testStoragePrefix, expectedArtifact, "data3")
		assert.NoError(t, err)
		err = datastore.ReadProtobuf(ctx, dataRef, &value)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(&value, getTestStringLiteralWithValue("value3")))
	})

	t.Run("Update by artifact tag", func(t *testing.T) {
		ctx := context.Background()
		datastore := createInmemoryDataStore(t, mockScope.NewTestScope())
		mockArtifactModel := getExpectedArtifactModel(ctx, t, datastore, expectedArtifact)

		metaData := &datacatalog.Metadata{
			KeyMap: map[string]string{"key2": "value2"},
		}
		serializedMetadata, err := transformers.SerializedMetadata(metaData)
		assert.NoError(t, err)

		dcRepo := newMockDataCatalogRepo()
		dcRepo.MockArtifactRepo.On("Update",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(artifact models.Artifact) bool {
				return artifact.ArtifactID == expectedArtifact.GetId() &&
					artifact.ArtifactKey.DatasetProject == expectedArtifact.GetDataset().GetProject() &&
					artifact.ArtifactKey.DatasetDomain == expectedArtifact.GetDataset().GetDomain() &&
					artifact.ArtifactKey.DatasetName == expectedArtifact.GetDataset().GetName() &&
					artifact.ArtifactKey.DatasetVersion == expectedArtifact.GetDataset().GetVersion() &&
					reflect.DeepEqual(artifact.SerializedMetadata, serializedMetadata)
			})).Return(nil)

		dcRepo.MockTagRepo.On("Get", mock.Anything,
			mock.MatchedBy(func(tag models.TagKey) bool {
				return tag.TagName == expectedTag.TagName &&
					tag.DatasetProject == expectedTag.DatasetProject &&
					tag.DatasetDomain == expectedTag.DatasetDomain &&
					tag.DatasetVersion == expectedTag.DatasetVersion &&
					tag.DatasetName == expectedTag.DatasetName
			})).Return(models.Tag{
			TagKey: models.TagKey{
				DatasetProject: expectedTag.DatasetProject,
				DatasetDomain:  expectedTag.DatasetDomain,
				DatasetName:    expectedTag.DatasetName,
				DatasetVersion: expectedTag.DatasetVersion,
				TagName:        expectedTag.TagName,
			},
			DatasetUUID: expectedTag.DatasetUUID,
			Artifact:    mockArtifactModel,
			ArtifactID:  mockArtifactModel.ArtifactID,
		}, nil)

		request := &datacatalog.UpdateArtifactRequest{
			Dataset: expectedDataset.GetId(),
			QueryHandle: &datacatalog.UpdateArtifactRequest_TagName{
				TagName: expectedTag.TagName,
			},
			Data: []*datacatalog.ArtifactData{
				{
					Name:  "data1",
					Value: getTestStringLiteralWithValue("value11"),
				},
				{
					Name:  "data3",
					Value: getTestStringLiteralWithValue("value3"),
				},
			},
			Metadata: &datacatalog.Metadata{
				KeyMap: map[string]string{"key2": "value2"},
			},
		}

		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.UpdateArtifact(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, artifactResponse)
		assert.Equal(t, expectedArtifact.GetId(), artifactResponse.GetArtifactId())
		dcRepo.MockArtifactRepo.AssertExpectations(t)

		// check that the datastore has the updated artifactData available
		// data1 should contain updated value
		dataRef, err := getExpectedDatastoreLocationFromName(ctx, datastore, testStoragePrefix, expectedArtifact, "data1")
		assert.NoError(t, err)
		var value core.Literal
		err = datastore.ReadProtobuf(ctx, dataRef, &value)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(&value, getTestStringLiteralWithValue("value11")))

		// data2 was not included in update payload, should be removed
		dataRef, err = getExpectedDatastoreLocationFromName(ctx, datastore, testStoragePrefix, expectedArtifact, "data2")
		assert.NoError(t, err)
		err = datastore.ReadProtobuf(ctx, dataRef, &value)
		assert.Error(t, err)
		assert.True(t, stdErrors.Is(err, os.ErrNotExist))

		// data3 did not exist before, should be present after update
		dataRef, err = getExpectedDatastoreLocationFromName(ctx, datastore, testStoragePrefix, expectedArtifact, "data3")
		assert.NoError(t, err)
		err = datastore.ReadProtobuf(ctx, dataRef, &value)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(&value, getTestStringLiteralWithValue("value3")))
	})

	t.Run("Artifact not found", func(t *testing.T) {
		ctx := context.Background()
		datastore := createInmemoryDataStore(t, mockScope.NewTestScope())

		dcRepo := newMockDataCatalogRepo()
		dcRepo.MockArtifactRepo.On("Get", mock.Anything, mock.Anything).Return(models.Artifact{}, repoErrors.GetMissingEntityError("Artifact", &datacatalog.Artifact{
			Dataset: expectedDataset.GetId(),
			Id:      expectedArtifact.GetId(),
		}))

		request := &datacatalog.UpdateArtifactRequest{
			Dataset: expectedDataset.GetId(),
			QueryHandle: &datacatalog.UpdateArtifactRequest_ArtifactId{
				ArtifactId: expectedArtifact.GetId(),
			},
			Data: []*datacatalog.ArtifactData{
				{
					Name:  "data1",
					Value: getTestStringLiteralWithValue("value11"),
				},
				{
					Name:  "data3",
					Value: getTestStringLiteralWithValue("value3"),
				},
			},
		}

		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.UpdateArtifact(ctx, request)
		assert.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
		assert.Nil(t, artifactResponse)
	})

	t.Run("Missing artifact ID", func(t *testing.T) {
		ctx := context.Background()
		datastore := createInmemoryDataStore(t, mockScope.NewTestScope())

		dcRepo := newMockDataCatalogRepo()

		request := &datacatalog.UpdateArtifactRequest{
			Dataset:     expectedDataset.GetId(),
			QueryHandle: &datacatalog.UpdateArtifactRequest_ArtifactId{},
			Data: []*datacatalog.ArtifactData{
				{
					Name:  "data1",
					Value: getTestStringLiteralWithValue("value11"),
				},
				{
					Name:  "data3",
					Value: getTestStringLiteralWithValue("value3"),
				},
			},
		}

		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.UpdateArtifact(ctx, request)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Nil(t, artifactResponse)
	})

	t.Run("Missing artifact tag", func(t *testing.T) {
		ctx := context.Background()
		datastore := createInmemoryDataStore(t, mockScope.NewTestScope())

		dcRepo := newMockDataCatalogRepo()

		request := &datacatalog.UpdateArtifactRequest{
			Dataset:     expectedDataset.GetId(),
			QueryHandle: &datacatalog.UpdateArtifactRequest_TagName{},
			Data: []*datacatalog.ArtifactData{
				{
					Name:  "data1",
					Value: getTestStringLiteralWithValue("value11"),
				},
				{
					Name:  "data3",
					Value: getTestStringLiteralWithValue("value3"),
				},
			},
		}

		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.UpdateArtifact(ctx, request)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Nil(t, artifactResponse)
	})

	t.Run("Missing artifact data", func(t *testing.T) {
		ctx := context.Background()
		datastore := createInmemoryDataStore(t, mockScope.NewTestScope())

		dcRepo := newMockDataCatalogRepo()

		request := &datacatalog.UpdateArtifactRequest{
			Dataset: expectedDataset.GetId(),
			QueryHandle: &datacatalog.UpdateArtifactRequest_ArtifactId{
				ArtifactId: expectedArtifact.GetId(),
			},
			Data: nil,
		}

		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.UpdateArtifact(ctx, request)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Nil(t, artifactResponse)
	})

	t.Run("Empty artifact data", func(t *testing.T) {
		ctx := context.Background()
		datastore := createInmemoryDataStore(t, mockScope.NewTestScope())

		dcRepo := newMockDataCatalogRepo()

		request := &datacatalog.UpdateArtifactRequest{
			Dataset: expectedDataset.GetId(),
			QueryHandle: &datacatalog.UpdateArtifactRequest_ArtifactId{
				ArtifactId: expectedArtifact.GetId(),
			},
			Data: []*datacatalog.ArtifactData{},
		}

		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.UpdateArtifact(ctx, request)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Nil(t, artifactResponse)
	})
}
