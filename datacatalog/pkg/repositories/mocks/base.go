package mocks

import "github.com/flyteorg/datacatalog/pkg/repositories/interfaces"

type DataCatalogRepo struct {
	MockDatasetRepo     *DatasetRepo
	MockArtifactRepo    *ArtifactRepo
	MockTagRepo         *TagRepo
	MockReservationRepo *ReservationRepo
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

func (m *DataCatalogRepo) ReservationRepo() interfaces.ReservationRepo {
	return m.MockReservationRepo
}
