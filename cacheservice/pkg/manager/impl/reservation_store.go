package impl

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/cacheservice/pkg/errors"
	"github.com/flyteorg/flyte/cacheservice/pkg/manager/interfaces"
	"github.com/flyteorg/flyte/cacheservice/pkg/runtime/configs"
	"github.com/flyteorg/flyte/cacheservice/repositories/models"
)

var (
	_ interfaces.ReservationDataStoreClient = &reservationMemClient{}
)

type reservationMemClient struct {
	keyValStore *map[string]*models.Reservation
}

func (r reservationMemClient) Create(ctx context.Context, reservation *models.Reservation, now time.Time) error {
	reservationStore := *r.keyValStore
	if _, exists := reservationStore[reservation.Key]; exists {
		return errors.NewCacheServiceErrorf(codes.AlreadyExists, "reservation with key %s already exists", reservation.Key)
	}
	reservationStore[reservation.Key] = reservation
	return nil
}

func (r reservationMemClient) Update(ctx context.Context, reservation *models.Reservation, now time.Time) error {
	reservationStore := *r.keyValStore
	reservationStore[reservation.Key] = reservation
	return nil
}

func (r reservationMemClient) Get(ctx context.Context, key string) (*models.Reservation, error) {
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

var (
	_ interfaces.ReservationDataStoreClient = &reservationRedisClient{}
)

type reservationRedisClient struct {
	RedisClient *redis.Client
}

// Create a new reservation in the data store. If the reservation already exists, return an error.
func (r *reservationRedisClient) Create(ctx context.Context, reservation *models.Reservation, now time.Time) error {
	ttl := reservation.ExpiresAt.Sub(now)

	marshaledReservation, err := json.Marshal(reservation)
	if err != nil {
		return errors.NewCacheServiceErrorf(codes.Internal, "Failed to marshal reservation with error: %v", err)
	}

	setCmd := r.RedisClient.SetNX(ctx, reservation.Key, marshaledReservation, ttl)
	if err := setCmd.Err(); err != nil {
		return errors.NewCacheServiceErrorf(codes.Internal, "Failed to set reservation with key %s due to error: %v", reservation.Key, err)
	}

	wasSet, err := setCmd.Result()
	if err != nil {
		return errors.NewCacheServiceErrorf(codes.Internal, "Failed to confirm setting of reservation with key %s", reservation.Key)
	}
	if !wasSet {
		return errors.NewCacheServiceErrorf(codes.AlreadyExists, "Reservation with key %s already exists", reservation.Key)
	}

	return nil
}

// Update acts as an upsert operation. Since we already check in cache_manager for collision handling, we reduce the
// number of calls to Redis by upsert-ing as opposed to inserting
func (r *reservationRedisClient) Update(ctx context.Context, reservation *models.Reservation, now time.Time) error {
	ttl := reservation.ExpiresAt.Sub(now)

	marshaledReservation, err := json.Marshal(reservation)
	if err != nil {
		return errors.NewCacheServiceErrorf(codes.Internal, "Failed to marshal reservation with error: %v", err)
	}

	_, err = r.RedisClient.Set(ctx, reservation.Key, marshaledReservation, ttl).Result()
	if err != nil {
		return errors.NewCacheServiceErrorf(codes.Internal, "Failed to upsert reservation with key %s: %v", reservation.Key, err)
	}

	return nil
}

// Get a reservation from the data store. If the reservation does not exist, return an error.
func (r *reservationRedisClient) Get(ctx context.Context, key string) (*models.Reservation, error) {
	val, err := r.RedisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, errors.NewNotFoundError("reservation", key)
	} else if err != nil {
		return nil, errors.NewCacheServiceErrorf(codes.Internal, "Failed to get reservation with key %s", key)
	}

	var reservation models.Reservation
	err = json.Unmarshal([]byte(val), &reservation)
	if err != nil {
		return nil, errors.NewCacheServiceErrorf(codes.Internal, "Failed to unmarshal reservation with error %v", err)
	}

	return &reservation, nil
}

// Delete a reservation from the data store.
func (r *reservationRedisClient) Delete(ctx context.Context, key string, ownerID string) error {
	_, err := r.RedisClient.Del(ctx, key).Result()
	if err != nil {
		return errors.NewCacheServiceErrorf(codes.Internal, "Failed deleting redis key %v with error %v", key, err)
	}
	return nil
}

func NewReservationDataStore(ctx context.Context, serviceConfig configs.CacheServiceConfig) interfaces.ReservationDataStoreClient {
	clientType := serviceConfig.ReservationDataStoreType

	switch clientType {
	case configs.Mem:
		keyValStore := make(map[string]*models.Reservation)
		return &reservationMemClient{keyValStore: &keyValStore}
	case configs.DynamoDB:
		panic("dynamodb not supported for reservation data store")
	case configs.Redis:
		rdb := redis.NewClient(
			&redis.Options{
				Addr:     serviceConfig.RedisAddress,
				Username: serviceConfig.RedisUsername,
				Password: serviceConfig.RedisPassword,
			})

		_, err := rdb.Ping(ctx).Result()
		if err != nil {
			panic("failed to connect to redis " + err.Error())
		}

		return &reservationRedisClient{
			RedisClient: rdb,
		}
	}

	panic("unsupported reservation data store type")
}