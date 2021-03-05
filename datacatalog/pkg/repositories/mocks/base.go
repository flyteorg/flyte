package mocks

import "github.com/flyteorg/datacatalog/pkg/repositories/interfaces"

type DataCatalogRepo struct {
	MockDatasetRepo  *DatasetRepo
	MockArtifactRepo *ArtifactRepo
	MockTagRepo      *TagRepo
}

func (m *DataCatalogRepo) DatasetRepo() interfaces.DatasetRepo {
	return m.MockDatasetRepo
}

func (m *DataCatalogRepo) ArtifactRepo() interfaces.ArtifactRepo {
	return m.MockArtifactRepo
}

func (m *DataCatalogRepo) TagRepo() interfaces.TagRepo {
	return m.MockTagRepo
}
