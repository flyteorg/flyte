package validation

import (
	"context"

	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/transformers"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	propellervalidators "github.com/flyteorg/flyte/flytepropeller/pkg/compiler/validators"
)

func ValidateSignalGetOrCreateRequest(ctx context.Context, request *admin.SignalGetOrCreateRequest) error {
	if request.GetId() == nil {
		return shared.GetMissingArgumentError("id")
	}
	if err := ValidateSignalIdentifier(request.GetId()); err != nil {
		return err
	}
	if request.GetType() == nil {
		return shared.GetMissingArgumentError("type")
	}

	return nil
}

func ValidateSignalIdentifier(identifier *core.SignalIdentifier) error {
	if identifier.GetExecutionId() == nil {
		return shared.GetMissingArgumentError(shared.ExecutionID)
	}
	if identifier.GetSignalId() == "" {
		return shared.GetMissingArgumentError("signal_id")
	}

	return ValidateWorkflowExecutionIdentifier(identifier.GetExecutionId())
}

func ValidateSignalListRequest(ctx context.Context, request *admin.SignalListRequest) error {
	if err := ValidateWorkflowExecutionIdentifier(request.GetWorkflowExecutionId()); err != nil {
		return shared.GetMissingArgumentError(shared.ExecutionID)
	}
	if err := ValidateLimit(request.GetLimit()); err != nil {
		return err
	}
	return nil
}

func ValidateSignalSetRequest(ctx context.Context, db repositoryInterfaces.Repository, request *admin.SignalSetRequest) error {
	if request.GetId() == nil {
		return shared.GetMissingArgumentError("id")
	}
	if err := ValidateSignalIdentifier(request.GetId()); err != nil {
		return err
	}
	if request.GetValue() == nil {
		return shared.GetMissingArgumentError("value")
	}

	// validate that signal value matches type of existing signal
	signalModel, err := transformers.CreateSignalModel(request.GetId(), nil, nil)
	if err != nil {
		return nil
	}
	lookupSignalModel, err := db.SignalRepo().Get(ctx, signalModel.SignalKey)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"failed to validate that signal [%v] exists, err: [%+v]",
			signalModel.SignalKey, err)
	}
	valueType := propellervalidators.LiteralTypeForLiteral(request.GetValue())
	lookupSignal, err := transformers.FromSignalModel(lookupSignalModel)
	if err != nil {
		return err
	}
	err = propellervalidators.ValidateLiteralType(valueType)
	if err != nil {
		return errors.NewInvalidLiteralTypeError("", err)
	}
	if !propellervalidators.AreTypesCastable(lookupSignal.GetType(), valueType) {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"requested signal value [%v] is not castable to existing signal type [%v]",
			request.GetValue(), lookupSignalModel.Type)
	}

	return nil
}
