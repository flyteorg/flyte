package repositories

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/cacheservice/pkg/errors"
	"github.com/flyteorg/flyte/cacheservice/pkg/repositories/clients/postgres"
	"github.com/flyteorg/flyte/cacheservice/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/cacheservice/pkg/repositories/models"
	"github.com/flyteorg/flyte/cacheservice/pkg/runtime/configs"
	"github.com/flyteorg/flyte/flytestdlib/database"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

var (
	_ interfaces.CachedOutputRepo = &memClient{}
)

type memClient struct {
	cacheMap *map[string]*models.CachedOutput
}

// Get returns the cached output for the given key. It returns an error if the key does not exist.
func (c *memClient) Get(ctx context.Context, key string) (*models.CachedOutput, error) {
	cache := *c.cacheMap
	if value, exists := cache[key]; exists {
		return value, nil
	}

	return nil, errors.NewNotFoundError("output", key)
}

// Put will always set the value for the given key. It will overwrite the existing value if it exists.
func (c *memClient) Put(ctx context.Context, key string, cachedOutput *models.CachedOutput) error {
	cache := *c.cacheMap
	cache[key] = cachedOutput
	return nil
}

// Delete is an idempotent operation. It will not return an error if the key does not exist.
func (c *memClient) Delete(ctx context.Context, key string) error {
	cache := *c.cacheMap
	if _, exists := cache[key]; exists {
		delete(cache, key)
		return nil
	}

	return errors.NewNotFoundError("output", key)
}

var (
	_ interfaces.ReservationRepo = &reservationMemClient{}
)

type reservationMemClient struct {
	keyValStore *map[string]*models.CacheReservation
}

func (r reservationMemClient) Create(ctx context.Context, reservation *models.CacheReservation, now time.Time) error {
	reservationStore := *r.keyValStore
	if _, exists := reservationStore[reservation.Key]; exists {
		return errors.NewCacheServiceErrorf(codes.AlreadyExists, "reservation with key %s already exists", reservation.Key)
	}
	reservationStore[reservation.Key] = reservation
	return nil
}

func (r reservationMemClient) Update(ctx context.Context, reservation *models.CacheReservation, now time.Time) error {
	reservationStore := *r.keyValStore
	reservationStore[reservation.Key] = reservation
	return nil
}

func (r reservationMemClient) Get(ctx context.Context, key string) (*models.CacheReservation, error) {
	reservationStore := *r.keyValStore
	if value, exists := reservationStore[key]; exists {
		return value, nil
	}

	return nil, errors.NewNotFoundError("reservation", key)
}

func (r reservationMemClient) Delete(ctx context.Context, key string, ownerID string) error {
	reservationStore := *r.keyValStore
	delete(reservationStore, key)
	return nil
}

type CacheRepo struct {
	cachedOutputRepo interfaces.CachedOutputRepo
	reservationRepo  interfaces.ReservationRepo
}

func (c CacheRepo) CachedOutputRepo() interfaces.CachedOutputRepo {
	return c.cachedOutputRepo
}

func (c CacheRepo) ReservationRepo() interfaces.ReservationRepo {
	return c.reservationRepo
}

func GetRepositories(ctx context.Context, serviceConfig configs.CacheServiceConfig, dbConfig database.DbConfig, scope promutils.Scope) interfaces.RepositoryInterface {
	outputClientType := serviceConfig.OutputDataStoreType
	reservationClientType := serviceConfig.ReservationDataStoreType

	var postgresConnection *gorm.DB
	//var postgresRepoMetrics postgres.GormMetrics
	if outputClientType == configs.Postgres || reservationClientType == configs.Postgres {
		var err error
		dbHandle, err := postgres.NewDBHandle(ctx, &dbConfig)
		if err != nil {
			panic(err)
		}
		postgresConnection = dbHandle.DB

		//postgresRepoMetrics = postgres.NewPostgresRepoMetrics(scope.NewSubScope("postgres"))
	}

	var outputRepo interfaces.CachedOutputRepo
	switch outputClientType {
	case configs.Mem:
		cacheMap := make(map[string]*models.CachedOutput)
		outputRepo = &memClient{cacheMap: &cacheMap}
	case configs.Postgres:
		outputRepo = postgres.NewPostgresOutputRepo(postgresConnection, scope.NewSubScope("repositories"))
	default:
		panic("Invalid output data store type")
	}

	var reservationRepo interfaces.ReservationRepo
	switch reservationClientType {
	case configs.Mem:
		keyValStore := make(map[string]*models.CacheReservation)
		reservationRepo = &reservationMemClient{keyValStore: &keyValStore}
	case configs.Postgres:
		reservationRepo = postgres.NewReservationRepo(postgresConnection, scope.NewSubScope("repositories"))
	default:
		panic("Invalid reservation data store type")
	}

	return &CacheRepo{
		cachedOutputRepo: outputRepo,
		reservationRepo:  reservationRepo,
	}
}
