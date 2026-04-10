package repository

import (
	"github.com/jmoiron/sqlx"

	"github.com/flyteorg/flyte/v2/cache_service/repository/impl"
	"github.com/flyteorg/flyte/v2/cache_service/repository/interfaces"
)

type repository struct {
	cachedOutputRepo interfaces.CachedOutputRepo
	reservationRepo  interfaces.ReservationRepo
}

func NewRepository(db *sqlx.DB) interfaces.Repository {
	return &repository{
		cachedOutputRepo: impl.NewCachedOutputRepo(db),
		reservationRepo:  impl.NewReservationRepo(db),
	}
}

func (r *repository) CachedOutputRepo() interfaces.CachedOutputRepo {
	return r.cachedOutputRepo
}

func (r *repository) ReservationRepo() interfaces.ReservationRepo {
	return r.reservationRepo
}
