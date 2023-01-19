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
	"github.com/stretchr/testify/require"
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

func TestGetOrExtendReservations_CreateReservations(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	setUpTagRepoGetNotFound(&dcRepo)

	tagNames := map[string]bool{tagName: true, "tag2": true, "nonexistence": true}
	dcRepo.MockReservationRepo.On("Get",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(key models.ReservationKey) bool {
			return key.DatasetProject == datasetID.Project &&
				key.DatasetDomain == datasetID.Domain &&
				key.DatasetVersion == datasetID.Version &&
				key.DatasetName == datasetID.Name &&
				tagNames[key.TagName]
		})).Return(models.Reservation{}, errors2.NewDataCatalogErrorf(codes.NotFound, "entry not found"))

	now := time.Now()
	dcRepo.MockReservationRepo.On("Create",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(reservation models.Reservation) bool {
			return reservation.DatasetProject == datasetID.Project &&
				reservation.DatasetDomain == datasetID.Domain &&
				reservation.DatasetName == datasetID.Name &&
				reservation.DatasetVersion == datasetID.Version &&
				tagNames[reservation.TagName] &&
				reservation.OwnerID == currentOwner &&
				reservation.ExpiresAt == now.Add(heartbeatInterval*heartbeatGracePeriodMultiplier)
		}),
		mock.MatchedBy(func(now time.Time) bool { return true }),
	).Return(nil)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, maxHeartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := &datacatalog.GetOrExtendReservationsRequest{
		Reservations: []*datacatalog.GetOrExtendReservationRequest{
			{
				ReservationId:     &reservationID,
				OwnerId:           currentOwner,
				HeartbeatInterval: heartbeatIntervalPb,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "tag2",
				},
				OwnerId:           currentOwner,
				HeartbeatInterval: heartbeatIntervalPb,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "nonexistence",
				},
				OwnerId:           currentOwner,
				HeartbeatInterval: heartbeatIntervalPb,
			},
		},
	}

	resp, err := reservationManager.GetOrExtendReservations(context.Background(), req)

	assert.Nil(t, err)
	require.NotNil(t, resp.GetReservations())
	require.Len(t, resp.GetReservations(), 3)
	for _, reservation := range resp.GetReservations() {
		assert.Equal(t, currentOwner, reservation.OwnerId)
		assert.Equal(t, heartbeatIntervalPb, reservation.HeartbeatInterval)
	}
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

func TestGetOrExtendReservations_MaxHeartbeatInterval(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	setUpTagRepoGetNotFound(&dcRepo)

	tagNames := map[string]bool{tagName: true, "tag2": true, "nonexistence": true}
	dcRepo.MockReservationRepo.On("Get",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(key models.ReservationKey) bool {
			return key.DatasetProject == datasetID.Project &&
				key.DatasetDomain == datasetID.Domain &&
				key.DatasetVersion == datasetID.Version &&
				key.DatasetName == datasetID.Name &&
				tagNames[key.TagName]
		})).Return(models.Reservation{}, errors2.NewDataCatalogErrorf(codes.NotFound, "entry not found"))

	now := time.Now()

	dcRepo.MockReservationRepo.On("Create",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(reservation models.Reservation) bool {
			return reservation.DatasetProject == datasetID.Project &&
				reservation.DatasetDomain == datasetID.Domain &&
				reservation.DatasetName == datasetID.Name &&
				reservation.DatasetVersion == datasetID.Version &&
				tagNames[reservation.TagName] &&
				reservation.OwnerID == currentOwner &&
				reservation.ExpiresAt == now.Add(heartbeatInterval*heartbeatGracePeriodMultiplier)
		}),
		mock.MatchedBy(func(now time.Time) bool { return true }),
	).Return(nil)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, heartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := &datacatalog.GetOrExtendReservationsRequest{
		Reservations: []*datacatalog.GetOrExtendReservationRequest{
			{
				ReservationId:     &reservationID,
				OwnerId:           currentOwner,
				HeartbeatInterval: maxHeartbeatIntervalPb,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "tag2",
				},
				OwnerId:           currentOwner,
				HeartbeatInterval: maxHeartbeatIntervalPb,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "nonexistence",
				},
				OwnerId:           currentOwner,
				HeartbeatInterval: maxHeartbeatIntervalPb,
			},
		},
	}

	resp, err := reservationManager.GetOrExtendReservations(context.Background(), req)

	assert.Nil(t, err)
	require.NotNil(t, resp.GetReservations())
	require.Len(t, resp.GetReservations(), 3)
	for _, reservation := range resp.GetReservations() {
		assert.Equal(t, currentOwner, reservation.OwnerId)
		assert.Equal(t, heartbeatIntervalPb, reservation.HeartbeatInterval)
	}
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

func TestGetOrExtendReservations_ExtendReservations(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	setUpTagRepoGetNotFound(&dcRepo)

	now := time.Now()
	prevExpiresAt := now.Add(time.Second * 10)

	tagNames := map[string]bool{tagName: true, "tag2": true, "nonexistence": true}
	setUpReservationRepoGet(&dcRepo, prevExpiresAt, tagName, "tag2", "nonexistence")

	dcRepo.MockReservationRepo.On("Update",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(reservation models.Reservation) bool {
			return reservation.DatasetProject == datasetID.Project &&
				reservation.DatasetDomain == datasetID.Domain &&
				reservation.DatasetName == datasetID.Name &&
				reservation.DatasetVersion == datasetID.Version &&
				tagNames[reservation.TagName] &&
				reservation.OwnerID == prevOwner &&
				reservation.ExpiresAt == now.Add(heartbeatInterval*heartbeatGracePeriodMultiplier)
		}),
		mock.MatchedBy(func(now time.Time) bool { return true }),
	).Return(nil)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, maxHeartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := &datacatalog.GetOrExtendReservationsRequest{
		Reservations: []*datacatalog.GetOrExtendReservationRequest{
			{
				ReservationId:     &reservationID,
				OwnerId:           prevOwner,
				HeartbeatInterval: heartbeatIntervalPb,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "tag2",
				},
				OwnerId:           prevOwner,
				HeartbeatInterval: heartbeatIntervalPb,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "nonexistence",
				},
				OwnerId:           prevOwner,
				HeartbeatInterval: heartbeatIntervalPb,
			},
		},
	}

	resp, err := reservationManager.GetOrExtendReservations(context.Background(), req)

	assert.Nil(t, err)
	require.NotNil(t, resp.GetReservations())
	require.Len(t, resp.GetReservations(), 3)
	for _, reservation := range resp.GetReservations() {
		assert.Equal(t, prevOwner, reservation.OwnerId)
	}
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

func TestGetOrExtendReservations_TakeOverReservations(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	setUpTagRepoGetNotFound(&dcRepo)

	now := time.Now()
	prevExpiresAt := now.Add(time.Second * 10 * time.Duration(-1))

	tagNames := map[string]bool{tagName: true, "tag2": true, "nonexistence": true}
	setUpReservationRepoGet(&dcRepo, prevExpiresAt, tagName, "tag2", "nonexistence")

	dcRepo.MockReservationRepo.On("Update",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(reservation models.Reservation) bool {
			return reservation.DatasetProject == datasetID.Project &&
				reservation.DatasetDomain == datasetID.Domain &&
				reservation.DatasetName == datasetID.Name &&
				reservation.DatasetVersion == datasetID.Version &&
				tagNames[reservation.TagName] &&
				reservation.OwnerID == currentOwner &&
				reservation.ExpiresAt == now.Add(heartbeatInterval*heartbeatGracePeriodMultiplier)
		}),
		mock.MatchedBy(func(now time.Time) bool { return true }),
	).Return(nil)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, maxHeartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := &datacatalog.GetOrExtendReservationsRequest{
		Reservations: []*datacatalog.GetOrExtendReservationRequest{
			{
				ReservationId:     &reservationID,
				OwnerId:           currentOwner,
				HeartbeatInterval: heartbeatIntervalPb,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "tag2",
				},
				OwnerId:           currentOwner,
				HeartbeatInterval: heartbeatIntervalPb,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "nonexistence",
				},
				OwnerId:           currentOwner,
				HeartbeatInterval: heartbeatIntervalPb,
			},
		},
	}

	resp, err := reservationManager.GetOrExtendReservations(context.Background(), req)

	assert.Nil(t, err)
	require.NotNil(t, resp.GetReservations())
	require.Len(t, resp.GetReservations(), 3)
	for _, reservation := range resp.GetReservations() {
		assert.Equal(t, currentOwner, reservation.OwnerId)
	}
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

func TestGetOrExtendReservations_ReservationsExist(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	setUpTagRepoGetNotFound(&dcRepo)

	now := time.Now()
	prevExpiresAt := now.Add(time.Second * 10)

	setUpReservationRepoGet(&dcRepo, prevExpiresAt, tagName, "tag2", "nonexistence")

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, maxHeartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := &datacatalog.GetOrExtendReservationsRequest{
		Reservations: []*datacatalog.GetOrExtendReservationRequest{
			{
				ReservationId:     &reservationID,
				OwnerId:           currentOwner,
				HeartbeatInterval: heartbeatIntervalPb,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "tag2",
				},
				OwnerId:           currentOwner,
				HeartbeatInterval: heartbeatIntervalPb,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "nonexistence",
				},
				OwnerId:           currentOwner,
				HeartbeatInterval: heartbeatIntervalPb,
			},
		},
	}

	resp, err := reservationManager.GetOrExtendReservations(context.Background(), req)

	assert.Nil(t, err)
	require.NotNil(t, resp.GetReservations())
	require.Len(t, resp.GetReservations(), 3)
	for _, reservation := range resp.GetReservations() {
		assert.Equal(t, prevOwner, reservation.OwnerId)
	}
}

func TestGetOrExtendReservations_MultipleStates(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	setUpTagRepoGetNotFound(&dcRepo)

	now := time.Now()
	prevExpiresAt := now.Add(time.Second * 10)

	dcRepo.MockReservationRepo.On("Get",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(key models.ReservationKey) bool {
			return key.DatasetProject == datasetID.Project &&
				key.DatasetDomain == datasetID.Domain &&
				key.DatasetVersion == datasetID.Version &&
				key.DatasetName == datasetID.Name &&
				key.TagName == tagName
		})).Return(models.Reservation{}, errors2.NewDataCatalogErrorf(codes.NotFound, "entry not found"))

	dcRepo.MockReservationRepo.On("Get",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(key models.ReservationKey) bool {
			return key.DatasetProject == datasetID.Project &&
				key.DatasetDomain == datasetID.Domain &&
				key.DatasetVersion == datasetID.Version &&
				key.DatasetName == datasetID.Name &&
				key.TagName == "prevOwner"
		})).Return(
		models.Reservation{
			ReservationKey: models.ReservationKey{
				DatasetProject: project,
				DatasetName:    name,
				DatasetDomain:  domain,
				DatasetVersion: version,
				TagName:        "prevOwner",
			},
			OwnerID:   prevOwner,
			ExpiresAt: prevExpiresAt,
		}, nil,
	)

	dcRepo.MockReservationRepo.On("Get",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(key models.ReservationKey) bool {
			return key.DatasetProject == datasetID.Project &&
				key.DatasetDomain == datasetID.Domain &&
				key.DatasetVersion == datasetID.Version &&
				key.DatasetName == datasetID.Name &&
				key.TagName == "currentOwner"
		})).Return(
		models.Reservation{
			ReservationKey: getReservationKey(),
			OwnerID:        currentOwner,
			ExpiresAt:      prevExpiresAt,
		}, nil,
	)

	dcRepo.MockReservationRepo.On("Get",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(key models.ReservationKey) bool {
			return key.DatasetProject == datasetID.Project &&
				key.DatasetDomain == datasetID.Domain &&
				key.DatasetVersion == datasetID.Version &&
				key.DatasetName == datasetID.Name &&
				key.TagName == "currentOwnerExpired"
		})).Return(
		models.Reservation{
			ReservationKey: getReservationKey(),
			OwnerID:        currentOwner,
			ExpiresAt:      now.Add(-10 * time.Second),
		}, nil,
	)

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

	updateTagNames := map[string]bool{"currentOwner": true, "currentOwnerExpired": true}
	dcRepo.MockReservationRepo.On("Update",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(reservation models.Reservation) bool {
			return reservation.DatasetProject == datasetID.Project &&
				reservation.DatasetDomain == datasetID.Domain &&
				reservation.DatasetName == datasetID.Name &&
				reservation.DatasetVersion == datasetID.Version &&
				updateTagNames[reservation.TagName] &&
				reservation.OwnerID == currentOwner
		}),
		mock.MatchedBy(func(now time.Time) bool { return true }),
	).Return(nil)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, maxHeartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := &datacatalog.GetOrExtendReservationsRequest{
		Reservations: []*datacatalog.GetOrExtendReservationRequest{
			{
				ReservationId:     &reservationID,
				OwnerId:           currentOwner,
				HeartbeatInterval: heartbeatIntervalPb,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "prevOwner",
				},
				OwnerId:           currentOwner,
				HeartbeatInterval: heartbeatIntervalPb,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "currentOwner",
				},
				OwnerId:           currentOwner,
				HeartbeatInterval: heartbeatIntervalPb,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "currentOwnerExpired",
				},
				OwnerId:           currentOwner,
				HeartbeatInterval: heartbeatIntervalPb,
			},
		},
	}

	resp, err := reservationManager.GetOrExtendReservations(context.Background(), req)

	assert.Nil(t, err)
	require.NotNil(t, resp.GetReservations())
	require.Len(t, resp.GetReservations(), 4)
	for _, reservation := range resp.GetReservations() {
		if reservation.ReservationId.TagName == "prevOwner" {
			assert.Equal(t, prevOwner, reservation.OwnerId)
		} else {
			assert.Equal(t, currentOwner, reservation.OwnerId)
		}
		assert.Equal(t, heartbeatIntervalPb, reservation.HeartbeatInterval)
	}
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

func TestReleaseReservations(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	now := time.Now()

	tagNames := map[string]bool{tagName: true, "tag2": true, "nonexistence": true}
	dcRepo.MockReservationRepo.On("Delete",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(reservationKey models.ReservationKey) bool {
			return reservationKey.DatasetProject == datasetID.Project &&
				reservationKey.DatasetDomain == datasetID.Domain &&
				reservationKey.DatasetName == datasetID.Name &&
				reservationKey.DatasetVersion == datasetID.Version &&
				tagNames[reservationKey.TagName]
		}),
		mock.MatchedBy(func(ownerID string) bool {
			return ownerID == currentOwner
		}),
	).Return(nil)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, maxHeartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := &datacatalog.ReleaseReservationsRequest{
		Reservations: []*datacatalog.ReleaseReservationRequest{
			{
				ReservationId: &reservationID,
				OwnerId:       currentOwner,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "tag2",
				},
				OwnerId: currentOwner,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "nonexistence",
				},
				OwnerId: currentOwner,
			},
		},
	}

	_, err := reservationManager.ReleaseReservations(context.Background(), req)

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

func TestReleaseReservations_Failure(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	now := time.Now()
	reservationErr := fmt.Errorf("unknown error")

	tagNames := map[string]bool{tagName: true, "tag2": true, "nonexistence": true}
	dcRepo.MockReservationRepo.On("Delete",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(reservationKey models.ReservationKey) bool {
			return reservationKey.DatasetProject == datasetID.Project &&
				reservationKey.DatasetDomain == datasetID.Domain &&
				reservationKey.DatasetName == datasetID.Name &&
				reservationKey.DatasetVersion == datasetID.Version &&
				tagNames[reservationKey.TagName]
		}),
		mock.MatchedBy(func(ownerID string) bool {
			return ownerID == currentOwner
		}),
	).Return(reservationErr)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, maxHeartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := &datacatalog.ReleaseReservationsRequest{
		Reservations: []*datacatalog.ReleaseReservationRequest{
			{
				ReservationId: &reservationID,
				OwnerId:       currentOwner,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "tag2",
				},
				OwnerId: currentOwner,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "nonexistence",
				},
				OwnerId: currentOwner,
			},
		},
	}

	_, err := reservationManager.ReleaseReservations(context.Background(), req)

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

func TestReleaseReservations_GracefulFailure(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	now := time.Now()
	reservationErr := errors3.GetMissingEntityError("Reservation",
		&datacatalog.ReservationID{
			DatasetId: &datasetID,
			TagName:   tagName,
		})

	tagNames := map[string]bool{tagName: true, "tag2": true, "nonexistence": true}
	dcRepo.MockReservationRepo.On("Delete",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(reservationKey models.ReservationKey) bool {
			return reservationKey.DatasetProject == datasetID.Project &&
				reservationKey.DatasetDomain == datasetID.Domain &&
				reservationKey.DatasetName == datasetID.Name &&
				reservationKey.DatasetVersion == datasetID.Version &&
				tagNames[reservationKey.TagName]
		}),
		mock.MatchedBy(func(ownerID string) bool {
			return ownerID == currentOwner
		}),
	).Return(reservationErr)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, maxHeartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := &datacatalog.ReleaseReservationsRequest{
		Reservations: []*datacatalog.ReleaseReservationRequest{
			{
				ReservationId: &reservationID,
				OwnerId:       currentOwner,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "tag2",
				},
				OwnerId: currentOwner,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "nonexistence",
				},
				OwnerId: currentOwner,
			},
		},
	}

	_, err := reservationManager.ReleaseReservations(context.Background(), req)

	assert.Nil(t, err)
}

func TestReleaseReservation_MultipleStates(t *testing.T) {
	dcRepo := getDatacatalogRepo()

	now := time.Now()
	notFoundErr := errors3.GetMissingEntityError("Reservation",
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
				(reservationKey.TagName == tagName || reservationKey.TagName == prevOwner)
		}),
		mock.MatchedBy(func(ownerID string) bool {
			return ownerID == currentOwner
		}),
	).Return(notFoundErr)

	dcRepo.MockReservationRepo.On("Delete",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(reservationKey models.ReservationKey) bool {
			return reservationKey.DatasetProject == datasetID.Project &&
				reservationKey.DatasetDomain == datasetID.Domain &&
				reservationKey.DatasetName == datasetID.Name &&
				reservationKey.DatasetVersion == datasetID.Version &&
				reservationKey.TagName == "currentOwner"
		}),
		mock.MatchedBy(func(ownerID string) bool {
			return ownerID == currentOwner
		}),
	).Return(nil)

	reservationManager := NewReservationManager(&dcRepo,
		heartbeatGracePeriodMultiplier, maxHeartbeatInterval,
		func() time.Time { return now }, mockScope.NewTestScope())

	req := &datacatalog.ReleaseReservationsRequest{
		Reservations: []*datacatalog.ReleaseReservationRequest{
			{
				ReservationId: &reservationID,
				OwnerId:       currentOwner,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "prevOwner",
				},
				OwnerId: currentOwner,
			},
			{
				ReservationId: &datacatalog.ReservationID{
					DatasetId: &datasetID,
					TagName:   "currentOwner",
				},
				OwnerId: currentOwner,
			},
		},
	}

	_, err := reservationManager.ReleaseReservations(context.Background(), req)

	assert.Nil(t, err)
}

func getDatacatalogRepo() mocks.DataCatalogRepo {
	return mocks.DataCatalogRepo{
		MockReservationRepo: &mocks.ReservationRepo{},
		MockTagRepo:         &mocks.TagRepo{},
	}
}

func setUpReservationRepoGet(dcRepo *mocks.DataCatalogRepo, prevExpiresAt time.Time, tagNames ...string) {
	if len(tagNames) == 0 {
		tagNames = []string{tagName}
	}
	tags := make(map[string]bool)
	for _, tn := range tagNames {
		tags[tn] = true
	}

	dcRepo.MockReservationRepo.On("Get",
		mock.MatchedBy(func(ctx context.Context) bool { return true }),
		mock.MatchedBy(func(key models.ReservationKey) bool {
			return key.DatasetProject == datasetID.Project &&
				key.DatasetDomain == datasetID.Domain &&
				key.DatasetVersion == datasetID.Version &&
				key.DatasetName == datasetID.Name &&
				tags[key.TagName]
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
