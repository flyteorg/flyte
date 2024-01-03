package validators

import (
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/datacatalog/pkg/errors"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
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

func ValidateGetOrExtendReservationsRequest(request *datacatalog.GetOrExtendReservationsRequest) error {
	if request.GetReservations() == nil {
		return NewMissingArgumentError("reservations")
	}

	var errs []error
	for _, req := range request.GetReservations() {
		if err := ValidateGetOrExtendReservationRequest(req); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.NewCollectedErrors(codes.InvalidArgument, errs)
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

func ValidateReleaseReservationsRequest(request *datacatalog.ReleaseReservationsRequest) error {
	if request.GetReservations() == nil {
		return NewMissingArgumentError("reservations")
	}

	var errs []error
	for _, req := range request.GetReservations() {
		if err := ValidateReleaseReservationRequest(req); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.NewCollectedErrors(codes.InvalidArgument, errs)
	}

	return nil
}
