// Defines the error messages used within FlyteAdmin categorized by common error codes.
package errors

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FlyteAdminError interface {
	Error() string
	Code() codes.Code
	GRPCStatus() *status.Status
	WithDetails(details proto.Message) (FlyteAdminError, error)
	String() string
}
type flyteAdminErrorImpl struct {
	status *status.Status
}

func (e *flyteAdminErrorImpl) Error() string {
	return e.status.Message()
}

func (e *flyteAdminErrorImpl) Code() codes.Code {
	return e.status.Code()
}

func (e *flyteAdminErrorImpl) GRPCStatus() *status.Status {
	return e.status
}

func (e *flyteAdminErrorImpl) String() string {
	return fmt.Sprintf("status: %v", e.status)
}

// enclose the error in the format that grpc server expect from golang:
//
//	https://github.com/grpc/grpc-go/blob/master/status/status.go#L133
func (e *flyteAdminErrorImpl) WithDetails(details proto.Message) (FlyteAdminError, error) {
	s, err := e.status.WithDetails(details)
	if err != nil {
		return nil, err
	}
	return NewFlyteAdminErrorFromStatus(s), nil
}

func NewFlyteAdminErrorFromStatus(status *status.Status) FlyteAdminError {
	return &flyteAdminErrorImpl{
		status: status,
	}
}

func NewFlyteAdminError(code codes.Code, message string) FlyteAdminError {
	return &flyteAdminErrorImpl{
		status: status.New(code, message),
	}
}

func NewFlyteAdminErrorf(code codes.Code, format string, a ...interface{}) FlyteAdminError {
	return NewFlyteAdminError(code, fmt.Sprintf(format, a...))
}

func toStringSlice(errors []error) []string {
	errSlice := make([]string, len(errors))
	for idx, err := range errors {
		errSlice[idx] = err.Error()
	}
	return errSlice
}

func NewCollectedFlyteAdminError(code codes.Code, errors []error) FlyteAdminError {
	return NewFlyteAdminError(code, strings.Join(toStringSlice(errors), ", "))
}

func NewAlreadyInTerminalStateError(ctx context.Context, errorMsg string, curPhase string) FlyteAdminError {
	logger.Warn(ctx, errorMsg)
	alreadyInTerminalPhase := &admin.EventErrorAlreadyInTerminalState{CurrentPhase: curPhase}
	reason := &admin.EventFailureReason{
		Reason: &admin.EventFailureReason_AlreadyInTerminalState{AlreadyInTerminalState: alreadyInTerminalPhase},
	}
	statusErr, transformationErr := NewFlyteAdminError(codes.FailedPrecondition, errorMsg).WithDetails(reason)
	if transformationErr != nil {
		logger.Panicf(ctx, "Failed to wrap grpc status in type 'Error': %v", transformationErr)
		return NewFlyteAdminErrorf(codes.FailedPrecondition, errorMsg)
	}
	return statusErr
}

func NewIncompatibleClusterError(ctx context.Context, errorMsg, curCluster string) FlyteAdminError {
	statusErr, transformationErr := NewFlyteAdminError(codes.FailedPrecondition, errorMsg).WithDetails(&admin.EventFailureReason{
		Reason: &admin.EventFailureReason_IncompatibleCluster{
			IncompatibleCluster: &admin.EventErrorIncompatibleCluster{
				Cluster: curCluster,
			},
		},
	})
	if transformationErr != nil {
		logger.Panicf(ctx, "Failed to wrap grpc status in type 'Error': %v", transformationErr)
		return NewFlyteAdminErrorf(codes.FailedPrecondition, errorMsg)
	}
	return statusErr
}

func NewWorkflowExistsDifferentStructureError(ctx context.Context, request *admin.WorkflowCreateRequest) FlyteAdminError {
	errorMsg := "workflow with different structure already exists"
	statusErr, transformationErr := NewFlyteAdminError(codes.InvalidArgument, errorMsg).WithDetails(&admin.CreateWorkflowFailureReason{
		Reason: &admin.CreateWorkflowFailureReason_ExistsDifferentStructure{
			ExistsDifferentStructure: &admin.WorkflowErrorExistsDifferentStructure{
				Id: request.Id,
			},
		},
	})
	if transformationErr != nil {
		logger.Panicf(ctx, "Failed to wrap grpc status in type 'Error': %v", transformationErr)
		return NewFlyteAdminErrorf(codes.InvalidArgument, errorMsg)
	}
	return statusErr
}

func NewWorkflowExistsIdenticalStructureError(ctx context.Context, request *admin.WorkflowCreateRequest) FlyteAdminError {
	errorMsg := "workflow with identical structure already exists"
	statusErr, transformationErr := NewFlyteAdminError(codes.AlreadyExists, errorMsg).WithDetails(&admin.CreateWorkflowFailureReason{
		Reason: &admin.CreateWorkflowFailureReason_ExistsIdenticalStructure{
			ExistsIdenticalStructure: &admin.WorkflowErrorExistsIdenticalStructure{
				Id: request.Id,
			},
		},
	})
	if transformationErr != nil {
		logger.Panicf(ctx, "Failed to wrap grpc status in type 'Error': %v", transformationErr)
		return NewFlyteAdminErrorf(codes.AlreadyExists, errorMsg)
	}
	return statusErr
}
