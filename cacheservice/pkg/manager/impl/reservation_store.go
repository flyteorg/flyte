package impl

import (
	"context"
	"fmt"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
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

func (r reservationMemClient) Create(ctx context.Context, reservation *models.Reservation) error {
	reservationStore := *r.keyValStore
	if _, exists := reservationStore[reservation.Key]; exists {
		return errors.NewCacheServiceErrorf(codes.AlreadyExists, "reservation with key %s already exists", reservation.Key)
	}
	reservationStore[reservation.Key] = reservation
	return nil
}

func (r reservationMemClient) Update(ctx context.Context, reservation *models.Reservation, now time.Time) error {
	reservationStore := *r.keyValStore
	if existingReservation, exists := reservationStore[reservation.Key]; exists {
		if existingReservation.ExpiresAt.Before(now) || existingReservation.OwnerID == reservation.OwnerID {
			reservationStore[reservation.Key] = reservation
			return nil
		}
		return errors.NewCacheServiceErrorf(codes.AlreadyExists, "reservation with key %s already exists", reservation.Key)
	}
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
	if value, exists := reservationStore[key]; exists {
		if value.OwnerID != ownerID {
			return errors.NewNotFoundError("reservation", fmt.Sprintf("key: %s owner: %s", key, ownerID))
		}
		delete(reservationStore, key)
		return nil
	}

	return errors.NewNotFoundError("reservation", fmt.Sprintf("key: %s owner: %s", key, ownerID))
}

type dynamoReservationClient struct {
	DynamoDbClient *dynamodb.Client
	TableName      string
}

func (d dynamoReservationClient) Create(ctx context.Context, reservation *models.Reservation) error {
	return errors.NewCacheServiceErrorf(codes.Unimplemented, "GetOrExtendReservation endpoint not implemented")
}

func (d dynamoReservationClient) Update(ctx context.Context, reservation *models.Reservation, now time.Time) error {
	return errors.NewCacheServiceErrorf(codes.Unimplemented, "GetOrExtendReservation endpoint not implemented")
}

func (d dynamoReservationClient) Get(ctx context.Context, key string) (*models.Reservation, error) {
	return nil, errors.NewCacheServiceErrorf(codes.Unimplemented, "GetOrExtendReservation endpoint not implemented")
}

func (d dynamoReservationClient) Delete(ctx context.Context, key string, ownerID string) error {
	return errors.NewCacheServiceErrorf(codes.Unimplemented, "GetOrExtendReservation endpoint not implemented")
}

func NewReservationDataStore(serviceConfig configs.CacheServiceConfig) interfaces.ReservationDataStoreClient {
	clientType := serviceConfig.DataStoreType

	switch clientType {
	case configs.Mem:
		keyValStore := make(map[string]*models.Reservation)
		return &reservationMemClient{keyValStore: &keyValStore}
	case configs.DynamoDB:
		cfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
			awsConfig.WithRegion(serviceConfig.AwsRegion),
		)
		if err != nil {
			panic("unable to load AWS config to connect to Dynamo" + err.Error())
		}
		return &dynamoReservationClient{
			DynamoDbClient: dynamodb.NewFromConfig(cfg),
			TableName:      "cache",
		}
	}

	panic("unsupported reservation data store type")
}
