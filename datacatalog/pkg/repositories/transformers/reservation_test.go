package transformers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

func TestFromReservationID(t *testing.T) {
	reservationID := datacatalog.ReservationID{
		DatasetId: &datacatalog.DatasetID{
			Project: "p",
			Name:    "n",
			Domain:  "d",
			Version: "v",
		},
		TagName: "t",
	}

	reservationKey := FromReservationID(&reservationID)
	assert.Equal(t, reservationKey.DatasetProject, reservationID.DatasetId.Project)
	assert.Equal(t, reservationKey.DatasetName, reservationID.DatasetId.Name)
	assert.Equal(t, reservationKey.DatasetDomain, reservationID.DatasetId.Domain)
	assert.Equal(t, reservationKey.DatasetVersion, reservationID.DatasetId.Version)
	assert.Equal(t, reservationKey.TagName, reservationID.TagName)
}

func TestCreateReservation(t *testing.T) {
	now := time.Now()
	heartbeatInterval := time.Second * 5
	modelReservation := models.Reservation{
		ReservationKey: models.ReservationKey{
			DatasetOrg:     "o",
			DatasetProject: "p",
			DatasetName:    "n",
			DatasetDomain:  "d",
			DatasetVersion: "v",
			TagName:        "t",
		},
		OwnerID:   "o",
		ExpiresAt: now,
	}

	reservation, err := CreateReservation(&modelReservation, heartbeatInterval)

	assert.Equal(t, err, nil)
	assert.Equal(t, reservation.ExpiresAt.AsTime(), modelReservation.ExpiresAt.UTC())
	assert.Equal(t, reservation.HeartbeatInterval.AsDuration(), heartbeatInterval)
	assert.Equal(t, reservation.OwnerId, modelReservation.OwnerID)

	reservationID := reservation.ReservationId
	assert.Equal(t, reservationID.TagName, modelReservation.TagName)

	datasetID := reservationID.DatasetId
	assert.Equal(t, datasetID.Org, modelReservation.DatasetOrg)
	assert.Equal(t, datasetID.Project, modelReservation.DatasetProject)
	assert.Equal(t, datasetID.Name, modelReservation.DatasetName)
	assert.Equal(t, datasetID.Domain, modelReservation.DatasetDomain)
	assert.Equal(t, datasetID.Version, modelReservation.DatasetVersion)
}
