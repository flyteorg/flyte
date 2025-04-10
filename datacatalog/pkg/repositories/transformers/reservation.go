package transformers

import (
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/ptypes"

	"github.com/flyteorg/flyte/datacatalog/pkg/errors"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

func FromReservationID(reservationID *datacatalog.ReservationID) models.ReservationKey {
	datasetID := reservationID.GetDatasetId()

	return models.ReservationKey{
		DatasetProject: datasetID.GetProject(),
		DatasetDomain:  datasetID.GetDomain(),
		DatasetName:    datasetID.GetName(),
		DatasetVersion: datasetID.GetVersion(),
		TagName:        reservationID.GetTagName(),
	}
}

func CreateReservation(reservation *models.Reservation, heartbeatInterval time.Duration) (datacatalog.Reservation, error) {
	expiresAtPb, err := timestamppb.New(reservation.ExpiresAt)
	if err != nil {
		return datacatalog.Reservation{}, errors.NewDataCatalogErrorf(codes.Internal, "failed to serialize expires at time")
	}

	heartbeatIntervalPb := durationpb.New(heartbeatInterval)
	return datacatalog.Reservation{
		ReservationId: &datacatalog.ReservationID{
			DatasetId: &datacatalog.DatasetID{
				Project: reservation.DatasetProject,
				Domain:  reservation.DatasetDomain,
				Name:    reservation.DatasetName,
				Version: reservation.DatasetVersion,
			},
			TagName: reservation.TagName,
		},
		OwnerId:           reservation.OwnerID,
		HeartbeatInterval: heartbeatIntervalPb,
		ExpiresAt:         expiresAtPb,
	}, nil
}
