package validators

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

func ValidateGetOrExtendReservationRequest(request *datacatalog.GetOrExtendReservationRequest) error {
	if request.GetReservationId() == nil {
		return NewMissingArgumentError("reservationID")
	}
	if err := ValidateDatasetID(request.GetReservationId().GetDatasetId()); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.GetReservationId().GetTagName(), tagName); err != nil {
		return err
	}

	if err := ValidateEmptyStringField(request.GetOwnerId(), "ownerID"); err != nil {
		return err
	}

	if request.GetHeartbeatInterval() == nil {
		return NewMissingArgumentError("heartbeatInterval")
	}

	return nil
}

func ValidateReleaseReservationRequest(request *datacatalog.ReleaseReservationRequest) error {
	if request.GetReservationId() == nil {
		return NewMissingArgumentError("reservationID")
	}
	if err := ValidateDatasetID(request.GetReservationId().GetDatasetId()); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.GetReservationId().GetTagName(), tagName); err != nil {
		return err
	}

	if err := ValidateEmptyStringField(request.GetOwnerId(), "ownerID"); err != nil {
		return err
	}

	return nil
}
