package repositories

import (
	"gorm.io/gorm"
	"k8s.io/utils/clock"

	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/gormimpl"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type PostgresRepo struct {
	datasetRepo     interfaces.DatasetRepo
	artifactRepo    interfaces.ArtifactRepo
	tagRepo         interfaces.TagRepo
	reservationRepo interfaces.ReservationRepo
}

func (dc *PostgresRepo) DatasetRepo() interfaces.DatasetRepo {
	return dc.datasetRepo
}

func (dc *PostgresRepo) ArtifactRepo() interfaces.ArtifactRepo {
	return dc.artifactRepo
}

func (dc *PostgresRepo) TagRepo() interfaces.TagRepo {
	return dc.tagRepo
}

func (dc *PostgresRepo) ReservationRepo() interfaces.ReservationRepo {
	return dc.reservationRepo
}

func NewPostgresRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) interfaces.DataCatalogRepo {
	return &PostgresRepo{
		datasetRepo:     gormimpl.NewDatasetRepo(db, errorTransformer, scope.NewSubScope("dataset")),
		artifactRepo:    gormimpl.NewArtifactRepo(db, errorTransformer, scope.NewSubScope("artifact"), clock.RealClock{}),
		tagRepo:         gormimpl.NewTagRepo(db, errorTransformer, scope.NewSubScope("tag")),
		reservationRepo: gormimpl.NewReservationRepo(db, errorTransformer, scope.NewSubScope("reservation")),
	}
}
