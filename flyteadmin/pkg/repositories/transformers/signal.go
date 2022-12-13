package transformers

import (
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/golang/protobuf/proto"

	"google.golang.org/grpc/codes"
)

func CreateSignalModel(signalID *core.SignalIdentifier, signalType *core.LiteralType, signalValue *core.Literal) (models.Signal, error) {
	signalModel := models.Signal{}
	if signalID != nil {
		signalKey := &signalModel.SignalKey
		if signalID.ExecutionId != nil {
			executionKey := &signalKey.ExecutionKey
			if len(signalID.ExecutionId.Project) > 0 {
				executionKey.Project = signalID.ExecutionId.Project
			}
			if len(signalID.ExecutionId.Domain) > 0 {
				executionKey.Domain = signalID.ExecutionId.Domain
			}
			if len(signalID.ExecutionId.Name) > 0 {
				executionKey.Name = signalID.ExecutionId.Name
			}
		}

		if len(signalID.SignalId) > 0 {
			signalKey.SignalID = signalID.SignalId
		}
	}

	if signalType != nil {
		typeBytes, err := proto.Marshal(signalType)
		if err != nil {
			return models.Signal{}, errors.NewFlyteAdminError(codes.Internal, "Failed to serialize signal type")
		}

		signalModel.Type = typeBytes
	}

	if signalValue != nil {
		valueBytes, err := proto.Marshal(signalValue)
		if err != nil {
			return models.Signal{}, errors.NewFlyteAdminError(codes.Internal, "Failed to serialize signal value")
		}

		signalModel.Value = valueBytes
	}

	return signalModel, nil
}

func initSignalIdentifier(id *core.SignalIdentifier) *core.SignalIdentifier {
	if id == nil {
		id = &core.SignalIdentifier{}
	}
	return id
}

func initWorkflowExecutionIdentifier(id *core.WorkflowExecutionIdentifier) *core.WorkflowExecutionIdentifier {
	if id == nil {
		return &core.WorkflowExecutionIdentifier{}
	}
	return id
}

func FromSignalModel(signalModel models.Signal) (admin.Signal, error) {
	signal := admin.Signal{}

	var executionID *core.WorkflowExecutionIdentifier
	if len(signalModel.SignalKey.ExecutionKey.Project) > 0 {
		executionID = initWorkflowExecutionIdentifier(executionID)
		executionID.Project = signalModel.SignalKey.ExecutionKey.Project
	}
	if len(signalModel.SignalKey.ExecutionKey.Domain) > 0 {
		executionID = initWorkflowExecutionIdentifier(executionID)
		executionID.Domain = signalModel.SignalKey.ExecutionKey.Domain
	}
	if len(signalModel.SignalKey.ExecutionKey.Name) > 0 {
		executionID = initWorkflowExecutionIdentifier(executionID)
		executionID.Name = signalModel.SignalKey.ExecutionKey.Name
	}

	var signalID *core.SignalIdentifier
	if executionID != nil {
		signalID = initSignalIdentifier(signalID)
		signalID.ExecutionId = executionID
	}
	if len(signalModel.SignalKey.SignalID) > 0 {
		signalID = initSignalIdentifier(signalID)
		signalID.SignalId = signalModel.SignalKey.SignalID
	}

	if signalID != nil {
		signal.Id = signalID
	}

	if len(signalModel.Type) > 0 {
		var typeDeserialized core.LiteralType
		err := proto.Unmarshal(signalModel.Type, &typeDeserialized)
		if err != nil {
			return admin.Signal{}, errors.NewFlyteAdminError(codes.Internal, "failed to unmarshal signal type")
		}
		signal.Type = &typeDeserialized
	}

	if len(signalModel.Value) > 0 {
		var valueDeserialized core.Literal
		err := proto.Unmarshal(signalModel.Value, &valueDeserialized)
		if err != nil {
			return admin.Signal{}, errors.NewFlyteAdminError(codes.Internal, "failed to unmarshal signal value")
		}
		signal.Value = &valueDeserialized
	}

	return signal, nil
}

func FromSignalModels(signalModels []models.Signal) ([]*admin.Signal, error) {
	signals := make([]*admin.Signal, len(signalModels))
	for idx, signalModel := range signalModels {
		signal, err := FromSignalModel(signalModel)
		if err != nil {
			return nil, err
		}
		signals[idx] = &signal
	}
	return signals, nil
}
