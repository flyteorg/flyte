package interfaces

type DataCatalogRepo interface {
	DatasetRepo() DatasetRepo
	ArtifactRepo() ArtifactRepo
	TagRepo() TagRepo
}
