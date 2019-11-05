package transformers

import (
	"testing"

	"github.com/lyft/datacatalog/pkg/repositories/models"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func getTestArtifactData() []*datacatalog.ArtifactData {
	testInteger := &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: &core.Primitive{Value: &core.Primitive_Integer{Integer: 1}},
				},
			},
		},
	}
	return []*datacatalog.ArtifactData{
		{Name: "data1", Value: testInteger},
		{Name: "data2", Value: testInteger},
	}
}

func getTestPartitions() []models.Partition {
	return []models.Partition{
		{DatasetUUID: "test-uuid", Key: "key1", Value: "value1"},
		{DatasetUUID: "test-uuid", Key: "key2", Value: "value2"},
	}
}

func getDatasetModel() models.Dataset {
	return models.Dataset{
		DatasetKey: models.DatasetKey{
			Project: datasetID.Project,
			Domain:  datasetID.Domain,
			Name:    datasetID.Name,
			Version: datasetID.Version,
			UUID:    datasetID.UUID,
		},
	}
}

func TestCreateArtifactModel(t *testing.T) {

	createArtifactRequest := datacatalog.CreateArtifactRequest{
		Artifact: &datacatalog.Artifact{
			Id:       "artifactID-1",
			Dataset:  &datasetID,
			Data:     getTestArtifactData(),
			Metadata: &metadata,
			Partitions: []*datacatalog.Partition{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
		},
	}

	testArtifactData := []models.ArtifactData{
		{Name: "data1", Location: "s3://test1"},
		{Name: "data3", Location: "s3://test2"},
	}

	artifactModel, err := CreateArtifactModel(createArtifactRequest, testArtifactData, getDatasetModel())
	assert.NoError(t, err)
	assert.Equal(t, artifactModel.ArtifactID, createArtifactRequest.Artifact.Id)
	assert.Equal(t, artifactModel.ArtifactKey.DatasetProject, datasetID.Project)
	assert.Equal(t, artifactModel.ArtifactKey.DatasetDomain, datasetID.Domain)
	assert.Equal(t, artifactModel.ArtifactKey.DatasetName, datasetID.Name)
	assert.Equal(t, artifactModel.ArtifactKey.DatasetVersion, datasetID.Version)
	assert.EqualValues(t, testArtifactData, artifactModel.ArtifactData)
	assert.EqualValues(t, getTestPartitions(), artifactModel.Partitions)
}

func TestCreateArtifactModelNoMetdata(t *testing.T) {
	createArtifactRequest := datacatalog.CreateArtifactRequest{
		Artifact: &datacatalog.Artifact{
			Id:      "artifactID-1",
			Dataset: &datasetID,
			Data:    getTestArtifactData(),
		},
	}

	testArtifactData := []models.ArtifactData{
		{Name: "data1", Location: "s3://test1"},
		{Name: "data3", Location: "s3://test2"},
	}
	artifactModel, err := CreateArtifactModel(createArtifactRequest, testArtifactData, getDatasetModel())
	assert.NoError(t, err)
	assert.Equal(t, []byte{}, artifactModel.SerializedMetadata)
	assert.Len(t, artifactModel.Partitions, 0)
}

func TestFromArtifactModel(t *testing.T) {
	artifactModel := models.Artifact{
		ArtifactKey: models.ArtifactKey{
			DatasetProject: "project1",
			DatasetDomain:  "domain1",
			DatasetName:    "name1",
			DatasetVersion: "version1",
			ArtifactID:     "id1",
		},
		SerializedMetadata: []byte{},
		Partitions: []models.Partition{
			{DatasetUUID: "test-uuid", Key: "key1", Value: "value1"},
			{DatasetUUID: "test-uuid", Key: "key2", Value: "value2"},
		},
	}

	actual, err := FromArtifactModel(artifactModel)
	assert.NoError(t, err)
	assert.Equal(t, artifactModel.ArtifactID, actual.Id)
	assert.Equal(t, artifactModel.DatasetProject, actual.Dataset.Project)
	assert.Equal(t, artifactModel.DatasetDomain, actual.Dataset.Domain)
	assert.Equal(t, artifactModel.DatasetName, actual.Dataset.Name)
	assert.Equal(t, artifactModel.DatasetVersion, actual.Dataset.Version)

	assert.Len(t, actual.Partitions, 2)
	assert.EqualValues(t, artifactModel.Partitions, getTestPartitions())
}

func TestToArtifactKey(t *testing.T) {
	artifactKey := ToArtifactKey(datasetID, "artifactID-1")
	assert.Equal(t, datasetID.Project, artifactKey.DatasetProject)
	assert.Equal(t, datasetID.Domain, artifactKey.DatasetDomain)
	assert.Equal(t, datasetID.Name, artifactKey.DatasetName)
	assert.Equal(t, datasetID.Version, artifactKey.DatasetVersion)
	assert.Equal(t, artifactKey.ArtifactID, "artifactID-1")
}
