package impl

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/lyft/datacatalog/pkg/errors"
	"github.com/lyft/datacatalog/pkg/repositories/mocks"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/contextutils"
	mockScope "github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "value1"}},
				},
			},
		},
	}
}

func getTestArtifact() *datacatalog.Artifact {

	return &datacatalog.Artifact{
		Id: "test-id",
		Dataset: &datacatalog.DatasetID{
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "test-name",
			Version: "test-version",
		},
		Metadata: &datacatalog.Metadata{
			KeyMap: map[string]string{"key1": "value1"},
		},
		Data: []*datacatalog.ArtifactData{
			{
				Name:  "data1",
				Value: getTestStringLiteral(),
			},
		},
	}
}

func newMockDataCatalogRepo() *mocks.DataCatalogRepo {
	return &mocks.DataCatalogRepo{
		MockDatasetRepo:  &mocks.DatasetRepo{},
		MockArtifactRepo: &mocks.ArtifactRepo{},
	}
}

func getExpectedDatastoreLocation(ctx context.Context, store *storage.DataStore, prefix storage.DataReference, artifact *datacatalog.Artifact, idx int) (storage.DataReference, error) {
	dataset := artifact.Dataset
	return store.ConstructReference(ctx, prefix, dataset.Project, dataset.Domain, dataset.Name, dataset.Version, artifact.Id, artifact.Data[idx].Name, artifactDataFile)
}

func TestCreateArtifact(t *testing.T) {
	ctx := context.Background()
	datastore := createInmemoryDataStore(t, mockScope.NewTestScope())
	testStoragePrefix, err := datastore.ConstructReference(ctx, datastore.GetBaseContainerFQN(ctx), "test")
	assert.NoError(t, err)

	t.Run("HappyPath", func(t *testing.T) {
		datastore := createInmemoryDataStore(t, mockScope.NewTestScope())
		ctx := context.Background()
		dcRepo := newMockDataCatalogRepo()
		dcRepo.MockDatasetRepo.On("Get", mock.Anything,
			mock.MatchedBy(func(dataset models.DatasetKey) bool {
				expectedDataset := getTestDataset()
				return dataset.Project == expectedDataset.Id.Project &&
					dataset.Domain == expectedDataset.Id.Domain &&
					dataset.Name == expectedDataset.Id.Name &&
					dataset.Version == expectedDataset.Id.Version
			})).Return(models.Dataset{
			DatasetKey: models.DatasetKey{
				Project: getTestDataset().Id.Project,
				Domain:  getTestDataset().Id.Domain,
				Name:    getTestDataset().Id.Name,
				Version: getTestDataset().Id.Version,
			},
		}, nil)

		dcRepo.MockArtifactRepo.On("Create",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(artifact models.Artifact) bool {
				expectedArtifact := getTestArtifact()
				return artifact.ArtifactID == expectedArtifact.Id &&
					artifact.SerializedMetadata != nil &&
					len(artifact.ArtifactData) == len(expectedArtifact.Data) &&
					artifact.ArtifactKey.DatasetProject == expectedArtifact.Dataset.Project &&
					artifact.ArtifactKey.DatasetDomain == expectedArtifact.Dataset.Domain &&
					artifact.ArtifactKey.DatasetName == expectedArtifact.Dataset.Name &&
					artifact.ArtifactKey.DatasetVersion == expectedArtifact.Dataset.Version
			})).Return(nil)

		request := datacatalog.CreateArtifactRequest{Artifact: getTestArtifact()}
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
		assert.Equal(t, value, *getTestArtifact().Data[0].Value)
	})

	t.Run("Dataset does not exist", func(t *testing.T) {
		dcRepo := newMockDataCatalogRepo()
		dcRepo.MockDatasetRepo.On("Get", mock.Anything, mock.Anything).Return(models.Dataset{}, status.Error(codes.NotFound, "not found"))

		request := datacatalog.CreateArtifactRequest{Artifact: getTestArtifact()}
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.CreateArtifact(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)
		responseCode := status.Code(err)
		assert.Equal(t, codes.NotFound, responseCode)
	})

	t.Run("Artifact missing ID", func(t *testing.T) {
		request := datacatalog.CreateArtifactRequest{
			Artifact: &datacatalog.Artifact{
				// missing artifact id
				Dataset: getTestDataset().Id,
			},
		}

		artifactManager := NewArtifactManager(&mocks.DataCatalogRepo{}, createInmemoryDataStore(t, mockScope.NewTestScope()), testStoragePrefix, mockScope.NewTestScope())
		_, err := artifactManager.CreateArtifact(ctx, request)
		assert.Error(t, err)
		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})

	t.Run("Artifact missing artifact data", func(t *testing.T) {
		request := datacatalog.CreateArtifactRequest{
			Artifact: &datacatalog.Artifact{
				Id:      "test",
				Dataset: getTestDataset().Id,
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
		dcRepo := &mocks.DataCatalogRepo{
			MockDatasetRepo:  &mocks.DatasetRepo{},
			MockArtifactRepo: &mocks.ArtifactRepo{},
		}
		dcRepo.MockDatasetRepo.On("Get", mock.Anything,
			mock.MatchedBy(func(dataset models.DatasetKey) bool {
				expectedDataset := getTestDataset()
				return dataset.Project == expectedDataset.Id.Project &&
					dataset.Domain == expectedDataset.Id.Domain &&
					dataset.Name == expectedDataset.Id.Name &&
					dataset.Version == expectedDataset.Id.Version
			})).Return(models.Dataset{
			DatasetKey: models.DatasetKey{
				Project: getTestDataset().Id.Project,
				Domain:  getTestDataset().Id.Domain,
				Name:    getTestDataset().Id.Name,
				Version: getTestDataset().Id.Version,
			},
		}, nil)

		dcRepo.MockArtifactRepo.On("Create",
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(artifact models.Artifact) bool {
				expectedArtifact := getTestArtifact()
				return artifact.ArtifactID == expectedArtifact.Id &&
					artifact.SerializedMetadata != nil &&
					len(artifact.ArtifactData) == len(expectedArtifact.Data) &&
					artifact.ArtifactKey.DatasetProject == expectedArtifact.Dataset.Project &&
					artifact.ArtifactKey.DatasetDomain == expectedArtifact.Dataset.Domain &&
					artifact.ArtifactKey.DatasetName == expectedArtifact.Dataset.Name &&
					artifact.ArtifactKey.DatasetVersion == expectedArtifact.Dataset.Version
			})).Return(status.Error(codes.AlreadyExists, "test already exists"))

		request := datacatalog.CreateArtifactRequest{Artifact: getTestArtifact()}
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.CreateArtifact(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)

		responseCode := status.Code(err)
		assert.Equal(t, codes.AlreadyExists, responseCode)
	})
}

func TestGetArtifact(t *testing.T) {
	ctx := context.Background()
	datastore := createInmemoryDataStore(t, mockScope.NewTestScope())
	testStoragePrefix, err := datastore.ConstructReference(ctx, datastore.GetBaseContainerFQN(ctx), "test")
	assert.NoError(t, err)

	dcRepo := &mocks.DataCatalogRepo{
		MockDatasetRepo:  &mocks.DatasetRepo{},
		MockArtifactRepo: &mocks.ArtifactRepo{},
		MockTagRepo:      &mocks.TagRepo{},
	}

	expectedArtifact := getTestArtifact()
	expectedDataset := expectedArtifact.Dataset

	// Write the artifact data to the expected location and see if the retrieved data matches
	dataLocation, err := getExpectedDatastoreLocation(ctx, datastore, testStoragePrefix, expectedArtifact, 0)
	assert.NoError(t, err)
	err = datastore.WriteProtobuf(ctx, dataLocation, storage.Options{}, getTestStringLiteral())
	assert.NoError(t, err)

	// construct the artifact model we will return on the queries
	serializedMetadata, err := proto.Marshal(expectedArtifact.Metadata)
	assert.NoError(t, err)
	datasetKey := models.DatasetKey{
		Project: expectedDataset.Project,
		Domain:  expectedDataset.Domain,
		Version: expectedDataset.Version,
		Name:    expectedDataset.Name,
	}
	testArtifactModel := models.Artifact{
		ArtifactKey: models.ArtifactKey{
			DatasetProject: expectedDataset.Project,
			DatasetDomain:  expectedDataset.Domain,
			DatasetVersion: expectedDataset.Version,
			DatasetName:    expectedDataset.Name,
			ArtifactID:     expectedArtifact.Id,
		},
		ArtifactData: []models.ArtifactData{
			{Name: "data1", Location: dataLocation.String()},
		},
		Dataset: models.Dataset{
			DatasetKey:         datasetKey,
			SerializedMetadata: serializedMetadata,
		},
		SerializedMetadata: serializedMetadata,
	}

	t.Run("Get by Id", func(t *testing.T) {

		dcRepo.MockArtifactRepo.On("Get", mock.Anything,
			mock.MatchedBy(func(artifactKey models.ArtifactKey) bool {
				return artifactKey.ArtifactID == expectedArtifact.Id &&
					artifactKey.DatasetProject == expectedArtifact.Dataset.Project &&
					artifactKey.DatasetDomain == expectedArtifact.Dataset.Domain &&
					artifactKey.DatasetVersion == expectedArtifact.Dataset.Version &&
					artifactKey.DatasetName == expectedArtifact.Dataset.Name
			})).Return(testArtifactModel, nil)

		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.GetArtifact(ctx, datacatalog.GetArtifactRequest{
			Dataset:     getTestDataset().Id,
			QueryHandle: &datacatalog.GetArtifactRequest_ArtifactId{ArtifactId: expectedArtifact.Id},
		})
		assert.NoError(t, err)

		assert.True(t, proto.Equal(expectedArtifact, artifactResponse.Artifact))
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
			Artifact:   testArtifactModel,
			ArtifactID: testArtifactModel.ArtifactID,
		}, nil)

		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.GetArtifact(ctx, datacatalog.GetArtifactRequest{
			Dataset:     getTestDataset().Id,
			QueryHandle: &datacatalog.GetArtifactRequest_TagName{TagName: expectedTag.TagName},
		})
		assert.NoError(t, err)

		assert.True(t, proto.Equal(expectedArtifact, artifactResponse.Artifact))
	})

	t.Run("Get missing input", func(t *testing.T) {
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.GetArtifact(ctx, datacatalog.GetArtifactRequest{Dataset: getTestDataset().Id})
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)
		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})

	t.Run("Get does not exist", func(t *testing.T) {
		dcRepo.MockTagRepo.On("Get", mock.Anything, mock.Anything).Return(
			models.Tag{}, errors.NewDataCatalogError(codes.NotFound, "tag with artifact does not exist"))
		artifactManager := NewArtifactManager(dcRepo, datastore, testStoragePrefix, mockScope.NewTestScope())
		artifactResponse, err := artifactManager.GetArtifact(ctx, datacatalog.GetArtifactRequest{Dataset: getTestDataset().Id, QueryHandle: &datacatalog.GetArtifactRequest_TagName{TagName: "test"}})
		assert.Error(t, err)
		assert.Nil(t, artifactResponse)
		responseCode := status.Code(err)
		assert.Equal(t, codes.NotFound, responseCode)
	})

}
