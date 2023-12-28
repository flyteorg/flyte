package interfaces

import (
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type DataCatalogRepo interface {
	DatasetRepo() DatasetRepo
	ArtifactRepo() ArtifactRepo
	TagRepo() TagRepo
	ReservationRepo() ReservationRepo
}

type NewRepositoryFunc = func(db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) DataCatalogRepo
