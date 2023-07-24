package impl

import (
	"context"
	"fmt"

	mockScope "github.com/flyteorg/flytestdlib/promutils"

	"testing"
	"time"

	errors2 "github.com/flyteorg/datacatalog/pkg/errors"
	errors3 "github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/mocks"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
)

var tagName = "tag"
var project = "p"
var name = "n"
var domain = "d"
var version = "v"
var datasetID = datacatalog.DatasetID{
	Project: project,
	Name:    name,
	Domain:  domain,
	Version: version,
}
var reservationID = datacatalog.ReservationID{
	DatasetId: &datasetID,
	TagName:   tagName,
}
var heartbeatInterval = time.Second * 5
var heartbeatIntervalPb = ptypes.DurationProto(heartbeatInterval)
var maxHeartbeatInterval = time.Second * 10
var maxHeartbeatIntervalPb = ptypes.DurationProto(maxHeartbeatInterval)
var heartbeatGracePeriodMultiplier = time.Second * 3
var prevOwner = "prevOwner"
var currentOwner = "currentOwner"

func TestGetOrExtendReservation_CreateReservation(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	setUpTagRepoGetNotFound(&dcRepo)

	dcRepo.MockReservationRepo.On("Get",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(key models.ReservationKey) bool {
			return key.DatasetProject == datasetID.Project &&
				key.DatasetDomain == datasetID.Domain &&
				key.DatasetVersion == datasetID.Version &&
				key.DatasetName == datasetID.Name &&
				key.TagName == tagName
		})).Return(models.Reservation{}, errors2.NewDataCatalogErrorf(codes.NotFound, "entry not found"))

	now := time.Now()

	dcRepo.MockReservationRepo.On("Create",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(reservation models.Reservation) bool {
			return reservation.DatasetProject == datasetID.Project &&
				reservation.DatasetDomain == datasetID.Domain &&
				reservation.DatasetName == datasetID.Name &&
				reservation.DatasetVersion == datasetID.Version &&
				reservation.TagName == tagName &&
				reservation.OwnerID == currentOwner &&
				reservation.ExpiresAt == now.Add(heartbeatInterval*heartbeatGracePeriodMultiplier)
		}),
		mock.MatchedBy(func(now time.Time) bool { return true }),
	).Return(nil)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, maxHeartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := datacatalog.GetOrExtendReservationRequest{
		ReservationId:     &reservationID,
		OwnerId:           currentOwner,
		HeartbeatInterval: heartbeatIntervalPb,
	}

	resp, err := reservationManager.GetOrExtendReservation(context.Background(), &req)

	assert.Nil(t, err)
	assert.Equal(t, currentOwner, resp.GetReservation().OwnerId)
	assert.Equal(t, heartbeatIntervalPb, resp.GetReservation().HeartbeatInterval)
}

func TestGetOrExtendReservation_MaxHeartbeatInterval(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	setUpTagRepoGetNotFound(&dcRepo)

	dcRepo.MockReservationRepo.On("Get",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(key models.ReservationKey) bool {
			return key.DatasetProject == datasetID.Project &&
				key.DatasetDomain == datasetID.Domain &&
				key.DatasetVersion == datasetID.Version &&
				key.DatasetName == datasetID.Name &&
				key.TagName == tagName
		})).Return(models.Reservation{}, errors2.NewDataCatalogErrorf(codes.NotFound, "entry not found"))

	now := time.Now()

	dcRepo.MockReservationRepo.On("Create",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(reservation models.Reservation) bool {
			return reservation.DatasetProject == datasetID.Project &&
				reservation.DatasetDomain == datasetID.Domain &&
				reservation.DatasetName == datasetID.Name &&
				reservation.DatasetVersion == datasetID.Version &&
				reservation.TagName == tagName &&
				reservation.OwnerID == currentOwner &&
				reservation.ExpiresAt == now.Add(heartbeatInterval*heartbeatGracePeriodMultiplier)
		}),
		mock.MatchedBy(func(now time.Time) bool { return true }),
	).Return(nil)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, heartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := datacatalog.GetOrExtendReservationRequest{
		ReservationId:     &reservationID,
		OwnerId:           currentOwner,
		HeartbeatInterval: maxHeartbeatIntervalPb,
	}

	resp, err := reservationManager.GetOrExtendReservation(context.Background(), &req)

	assert.Nil(t, err)
	assert.Equal(t, currentOwner, resp.GetReservation().OwnerId)
	assert.Equal(t, heartbeatIntervalPb, resp.GetReservation().HeartbeatInterval)
}

func TestGetOrExtendReservation_ExtendReservation(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	setUpTagRepoGetNotFound(&dcRepo)

	now := time.Now()
	prevExpiresAt := now.Add(time.Second * 10)

	setUpReservationRepoGet(&dcRepo, prevExpiresAt)

	dcRepo.MockReservationRepo.On("Update",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(reservation models.Reservation) bool {
			return reservation.DatasetProject == datasetID.Project &&
				reservation.DatasetDomain == datasetID.Domain &&
				reservation.DatasetName == datasetID.Name &&
				reservation.DatasetVersion == datasetID.Version &&
				reservation.TagName == tagName &&
				reservation.OwnerID == prevOwner &&
				reservation.ExpiresAt == now.Add(heartbeatInterval*heartbeatGracePeriodMultiplier)
		}),
		mock.MatchedBy(func(now time.Time) bool { return true }),
	).Return(nil)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, maxHeartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := datacatalog.GetOrExtendReservationRequest{
		ReservationId:     &reservationID,
		OwnerId:           prevOwner,
		HeartbeatInterval: heartbeatIntervalPb,
	}

	resp, err := reservationManager.GetOrExtendReservation(context.Background(), &req)

	assert.Nil(t, err)
	assert.Equal(t, prevOwner, resp.GetReservation().OwnerId)
}

func TestGetOrExtendReservation_TakeOverReservation(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	setUpTagRepoGetNotFound(&dcRepo)

	now := time.Now()
	prevExpiresAt := now.Add(time.Second * 10 * time.Duration(-1))

	setUpReservationRepoGet(&dcRepo, prevExpiresAt)

	dcRepo.MockReservationRepo.On("Update",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(reservation models.Reservation) bool {
			return reservation.DatasetProject == datasetID.Project &&
				reservation.DatasetDomain == datasetID.Domain &&
				reservation.DatasetName == datasetID.Name &&
				reservation.DatasetVersion == datasetID.Version &&
				reservation.TagName == tagName &&
				reservation.OwnerID == currentOwner &&
				reservation.ExpiresAt == now.Add(heartbeatInterval*heartbeatGracePeriodMultiplier)
		}),
		mock.MatchedBy(func(now time.Time) bool { return true }),
	).Return(nil)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, maxHeartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := datacatalog.GetOrExtendReservationRequest{
		ReservationId:     &reservationID,
		OwnerId:           currentOwner,
		HeartbeatInterval: heartbeatIntervalPb,
	}

	resp, err := reservationManager.GetOrExtendReservation(context.Background(), &req)

	assert.Nil(t, err)
	assert.Equal(t, currentOwner, resp.GetReservation().OwnerId)
}

func TestGetOrExtendReservation_ReservationExists(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	setUpTagRepoGetNotFound(&dcRepo)

	now := time.Now()
	prevExpiresAt := now.Add(time.Second * 10)

	setUpReservationRepoGet(&dcRepo, prevExpiresAt)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, maxHeartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := datacatalog.GetOrExtendReservationRequest{
		ReservationId:     &reservationID,
		OwnerId:           currentOwner,
		HeartbeatInterval: heartbeatIntervalPb,
	}

	resp, err := reservationManager.GetOrExtendReservation(context.Background(), &req)

	assert.Nil(t, err)
	assert.Equal(t, prevOwner, resp.GetReservation().OwnerId)
}

func TestReleaseReservation(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	now := time.Now()

	dcRepo.MockReservationRepo.On("Delete",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(reservationKey models.ReservationKey) bool {
			return reservationKey.DatasetProject == datasetID.Project &&
				reservationKey.DatasetDomain == datasetID.Domain &&
				reservationKey.DatasetName == datasetID.Name &&
				reservationKey.DatasetVersion == datasetID.Version &&
				reservationKey.TagName == tagName
		}),
		mock.MatchedBy(func(ownerID string) bool {
			return ownerID == currentOwner
		}),
	).Return(nil)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, maxHeartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := datacatalog.ReleaseReservationRequest{
		ReservationId: &reservationID,
		OwnerId:       currentOwner,
	}

	_, err := reservationManager.ReleaseReservation(context.Background(), &req)

	assert.Nil(t, err)
}

func TestReleaseReservation_Failure(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	now := time.Now()
	reservationErr := fmt.Errorf("unknown error")

	dcRepo.MockReservationRepo.On("Delete",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(reservationKey models.ReservationKey) bool {
			return reservationKey.DatasetProject == datasetID.Project &&
				reservationKey.DatasetDomain == datasetID.Domain &&
				reservationKey.DatasetName == datasetID.Name &&
				reservationKey.DatasetVersion == datasetID.Version &&
				reservationKey.TagName == tagName
		}),
		mock.MatchedBy(func(ownerID string) bool {
			return ownerID == currentOwner
		}),
	).Return(reservationErr)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, maxHeartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := datacatalog.ReleaseReservationRequest{
		ReservationId: &reservationID,
		OwnerId:       currentOwner,
	}

	_, err := reservationManager.ReleaseReservation(context.Background(), &req)

	assert.Equal(t, reservationErr, err)
}

func TestReleaseReservation_GracefulFailure(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	now := time.Now()
	reservationErr := errors3.GetMissingEntityError("Reservation",
		&datacatalog.ReservationID{
			DatasetId: &datasetID,
			TagName:   tagName,
		})

	dcRepo.MockReservationRepo.On("Delete",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(reservationKey models.ReservationKey) bool {
			return reservationKey.DatasetProject == datasetID.Project &&
				reservationKey.DatasetDomain == datasetID.Domain &&
				reservationKey.DatasetName == datasetID.Name &&
				reservationKey.DatasetVersion == datasetID.Version &&
				reservationKey.TagName == tagName
		}),
		mock.MatchedBy(func(ownerID string) bool {
			return ownerID == currentOwner
		}),
	).Return(reservationErr)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, maxHeartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := datacatalog.ReleaseReservationRequest{
		ReservationId: &reservationID,
		OwnerId:       currentOwner,
	}

	_, err := reservationManager.ReleaseReservation(context.Background(), &req)

	assert.Nil(t, err)
}

func getDatacatalogRepo() mocks.DataCatalogRepo {
	return mocks.DataCatalogRepo{
		MockReservationRepo: &mocks.ReservationRepo{},
		MockTagRepo:         &mocks.TagRepo{},
	}
}

func setUpReservationRepoGet(dcRepo *mocks.DataCatalogRepo, prevExpiresAt time.Time) {
	dcRepo.MockReservationRepo.On("Get",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(key models.ReservationKey) bool {
			return key.DatasetProject == datasetID.Project &&
				key.DatasetDomain == datasetID.Domain &&
				key.DatasetVersion == datasetID.Version &&
				key.DatasetName == datasetID.Name &&
				key.TagName == tagName
		})).Return(
		models.Reservation{
			ReservationKey: getReservationKey(),
			OwnerID:        prevOwner,
			ExpiresAt:      prevExpiresAt,
		}, nil,
	)
}

func setUpTagRepoGetNotFound(dcRepo *mocks.DataCatalogRepo) {
	dcRepo.MockTagRepo.On("Get",
		mock.Anything,
		mock.Anything,
	).Return(models.Tag{}, errors2.NewDataCatalogErrorf(codes.NotFound, "entry not found"))
}

func getReservationKey() models.ReservationKey {
	return models.ReservationKey{
		DatasetProject: project,
		DatasetName:    name,
		DatasetDomain:  domain,
		DatasetVersion: version,
		TagName:        tagName,
	}
}
