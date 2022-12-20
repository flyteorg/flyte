package validation

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	repositoryInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	propellervalidators "github.com/flyteorg/flytepropeller/pkg/compiler/validators"

	"google.golang.org/grpc/codes"
)

func ValidateSignalGetOrCreateRequest(ctx context.Context, request admin.SignalGetOrCreateRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError("id")
	}
	if err := ValidateSignalIdentifier(*request.Id); err != nil {
		return err
	}
	if request.Type == nil {
		return shared.GetMissingArgumentError("type")
	}

	return nil
}

func ValidateSignalIdentifier(identifier core.SignalIdentifier) error {
	if identifier.ExecutionId == nil {
		return shared.GetMissingArgumentError(shared.ExecutionID)
	}
	if identifier.SignalId == "" {
		return shared.GetMissingArgumentError("signal_id")
	}

	return ValidateWorkflowExecutionIdentifier(identifier.ExecutionId)
}

func ValidateSignalListRequest(ctx context.Context, request admin.SignalListRequest) error {
	if err := ValidateWorkflowExecutionIdentifier(request.WorkflowExecutionId); err != nil {
		return shared.GetMissingArgumentError(shared.ExecutionID)
	}
	if err := ValidateLimit(request.Limit); err != nil {
		return err
	}
	return nil
}

func ValidateSignalSetRequest(ctx context.Context, db repositoryInterfaces.Repository, request admin.SignalSetRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError("id")
	}
	if err := ValidateSignalIdentifier(*request.Id); err != nil {
		return err
	}
	if request.Value == nil {
		return shared.GetMissingArgumentError("value")
	}

	// validate that signal value matches type of existing signal
	signalModel, err := transformers.CreateSignalModel(request.Id, nil, nil)
	if err != nil {
		return nil
	}
	lookupSignalModel, err := db.SignalRepo().Get(ctx, signalModel.SignalKey)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"failed to validate that signal [%v] exists, err: [%+v]",
			signalModel.SignalKey, err)
	}
	valueType := propellervalidators.LiteralTypeForLiteral(request.Value)
	lookupSignal, err := transformers.FromSignalModel(lookupSignalModel)
	if err != nil {
		return err
	}
	if !propellervalidators.AreTypesCastable(lookupSignal.Type, valueType) {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"requested signal value [%v] is not castable to existing signal type [%v]",
			request.Value, lookupSignalModel.Type)
	}

	return nil
}
