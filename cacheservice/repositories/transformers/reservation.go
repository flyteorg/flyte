package transformers

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/cacheservice/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/cacheservice"
)

func FromReservationModel(ctx context.Context, reservationModel *models.Reservation) *cacheservice.Reservation {
	return &cacheservice.Reservation{
		Key:       reservationModel.Key,
		OwnerId:   reservationModel.OwnerID,
		ExpiresAt: timestamppb.New(reservationModel.ExpiresAt),
	}
}
