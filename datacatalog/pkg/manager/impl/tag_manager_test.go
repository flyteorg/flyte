package impl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/mocks"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	mockScope "github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

func init() {
	labeled.SetMetricKeys(contextutils.AppNameKey)
}

func getTestTag() models.Tag {
	return models.Tag{
		TagKey: models.TagKey{
			DatasetOrg:     "test-org",
			DatasetProject: "test-project",
			DatasetDomain:  "test-domain",
			DatasetVersion: "test-version",
			DatasetName:    "test-name",
			TagName:        "test-tag",
		},
		DatasetUUID: "test-uuid",
		ArtifactID:  "test-artifactID",
	}
}

func TestAddTag(t *testing.T) {
	dcRepo := &mocks.DataCatalogRepo{
		MockDatasetRepo:  &mocks.DatasetRepo{},
		MockArtifactRepo: &mocks.ArtifactRepo{},
		MockTagRepo:      &mocks.TagRepo{},
	}

	expectedTag := getTestTag()

	t.Run("HappyPath", func(t *testing.T) {
		dcRepo.MockTagRepo.On("Create", mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(tag models.Tag) bool {
				return tag.DatasetOrg == expectedTag.DatasetOrg && tag.DatasetProject == expectedTag.DatasetProject &&
					tag.DatasetDomain == expectedTag.DatasetDomain &&
					tag.DatasetName == expectedTag.DatasetName &&
					tag.DatasetVersion == expectedTag.DatasetVersion &&
					tag.ArtifactID == expectedTag.ArtifactID &&
					tag.TagName == expectedTag.TagName
			})).Return(nil)

		artifact := models.Artifact{
			ArtifactKey: models.ArtifactKey{
				DatasetOrg:     expectedTag.DatasetOrg,
				DatasetProject: expectedTag.DatasetProject,
				DatasetDomain:  expectedTag.DatasetDomain,
				DatasetName:    expectedTag.DatasetVersion,
				DatasetVersion: expectedTag.DatasetName,
				ArtifactID:     expectedTag.ArtifactID,
			},
			DatasetUUID: expectedTag.DatasetUUID,
		}

		dataset := models.Dataset{
			DatasetKey: models.DatasetKey{
				Org:     expectedTag.DatasetOrg,
				Project: expectedTag.DatasetProject,
				Domain:  expectedTag.DatasetDomain,
				Version: expectedTag.DatasetVersion,
				Name:    expectedTag.DatasetName,
			},
		}

		dcRepo.MockArtifactRepo.On("Get", mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(artifactKey models.ArtifactKey) bool {
				return artifactKey.DatasetOrg == expectedTag.DatasetOrg &&
					artifactKey.DatasetProject == expectedTag.DatasetProject &&
					artifactKey.DatasetDomain == expectedTag.DatasetDomain &&
					artifactKey.DatasetName == expectedTag.DatasetName &&
					artifactKey.DatasetVersion == expectedTag.DatasetVersion &&
					artifactKey.ArtifactID == expectedTag.ArtifactID
			})).Return(artifact, nil)

		dcRepo.MockDatasetRepo.On("Get", mock.MatchedBy(func(ctx context.Context) bool { return true }),
			mock.MatchedBy(func(datasetKey models.DatasetKey) bool {
				return datasetKey.Org == expectedTag.DatasetOrg &&
					datasetKey.Project == expectedTag.DatasetProject &&
					datasetKey.Domain == expectedTag.DatasetDomain &&
					datasetKey.Name == expectedTag.DatasetName &&
					datasetKey.Version == expectedTag.DatasetVersion
			})).Return(dataset, nil)

		tagManager := NewTagManager(dcRepo, nil, mockScope.NewTestScope())
		_, err := tagManager.AddTag(context.Background(), &datacatalog.AddTagRequest{
			Tag: &datacatalog.Tag{
				Name:       expectedTag.TagName,
				ArtifactId: expectedTag.ArtifactID,
				Dataset: &datacatalog.DatasetID{
					Org:     expectedTag.DatasetOrg,
					Project: expectedTag.DatasetProject,
					Domain:  expectedTag.DatasetDomain,
					Version: expectedTag.DatasetVersion,
					Name:    expectedTag.DatasetName,
					UUID:    expectedTag.DatasetUUID,
				},
			},
		})

		assert.NoError(t, err)
	})

	t.Run("NoDataset", func(t *testing.T) {
		tagManager := NewTagManager(dcRepo, nil, mockScope.NewTestScope())
		_, err := tagManager.AddTag(context.Background(), &datacatalog.AddTagRequest{
			Tag: &datacatalog.Tag{
				Name:       "noDataset",
				ArtifactId: "noDataset",
			},
		})

		assert.Error(t, err)
		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})

	t.Run("NoTagName", func(t *testing.T) {
		tagManager := NewTagManager(dcRepo, nil, mockScope.NewTestScope())
		_, err := tagManager.AddTag(context.Background(), &datacatalog.AddTagRequest{
			Tag: &datacatalog.Tag{
				ArtifactId: "noArtifact",
				Dataset:    getTestDataset().Id,
			},
		})

		assert.Error(t, err)
		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})

	t.Run("NoArtifactID", func(t *testing.T) {
		tagManager := NewTagManager(dcRepo, nil, mockScope.NewTestScope())
		_, err := tagManager.AddTag(context.Background(), &datacatalog.AddTagRequest{
			Tag: &datacatalog.Tag{
				Name:    "noArtifact",
				Dataset: getTestDataset().Id,
			},
		})

		assert.Error(t, err)
		responseCode := status.Code(err)
		assert.Equal(t, codes.InvalidArgument, responseCode)
	})
}
