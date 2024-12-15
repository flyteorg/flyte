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
	assert.Equal(t, reservationKey.DatasetProject, reservationID.GetDatasetId().GetProject())
	assert.Equal(t, reservationKey.DatasetName, reservationID.GetDatasetId().GetName())
	assert.Equal(t, reservationKey.DatasetDomain, reservationID.GetDatasetId().GetDomain())
	assert.Equal(t, reservationKey.DatasetVersion, reservationID.GetDatasetId().GetVersion())
	assert.Equal(t, reservationKey.TagName, reservationID.GetTagName())
}

func TestCreateReservation(t *testing.T) {
	now := time.Now()
	heartbeatInterval := time.Second * 5
	modelReservation := models.Reservation{
		ReservationKey: models.ReservationKey{
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
	assert.Equal(t, reservation.GetExpiresAt().AsTime(), modelReservation.ExpiresAt.UTC())
	assert.Equal(t, reservation.GetHeartbeatInterval().AsDuration(), heartbeatInterval)
	assert.Equal(t, reservation.GetOwnerId(), modelReservation.OwnerID)

	reservationID := reservation.GetReservationId()
	assert.Equal(t, reservationID.GetTagName(), modelReservation.TagName)

	datasetID := reservationID.GetDatasetId()
	assert.Equal(t, datasetID.GetProject(), modelReservation.DatasetProject)
	assert.Equal(t, datasetID.GetName(), modelReservation.DatasetName)
	assert.Equal(t, datasetID.GetDomain(), modelReservation.DatasetDomain)
	assert.Equal(t, datasetID.GetVersion(), modelReservation.DatasetVersion)
}
