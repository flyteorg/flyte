// Defines the error messages used within FlyteAdmin categorized by common error codes.
package errors

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/wI2L/jsondiff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
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

func compareJsons(jsonArray1 jsondiff.Patch, jsonArray2 jsondiff.Patch) []string {
	var results []string
	map1 := make(map[string]jsondiff.Operation)
	for _, obj := range jsonArray1 {
		map1[obj.Path] = obj
	}

	for _, obj := range jsonArray2 {
		if val, ok := map1[obj.Path]; ok {
			result := fmt.Sprintf("%v:\n\t- %v \n\t+ %v", obj.Path, obj.Value, val.Value)
			results = append(results, result)
		}
	}
	return results
}

func NewTaskExistsDifferentStructureError(ctx context.Context, request *admin.TaskCreateRequest, oldSpec *core.CompiledTask, newSpec *core.CompiledTask) FlyteAdminError {
	errorMsg := fmt.Sprintf("%v task with different structure already exists. (Please register a new version of the task):\n", request.Id.Name)
	diff, _ := jsondiff.Compare(oldSpec, newSpec)
	rdiff, _ := jsondiff.Compare(newSpec, oldSpec)
	rs := compareJsons(diff, rdiff)

	errorMsg += strings.Join(rs, "\n")

	return NewFlyteAdminErrorf(codes.InvalidArgument, errorMsg)
}

func NewTaskExistsIdenticalStructureError(ctx context.Context, request *admin.TaskCreateRequest) FlyteAdminError {
	errorMsg := "task with identical structure already exists"
	return NewFlyteAdminErrorf(codes.AlreadyExists, errorMsg)
}

func NewWorkflowExistsDifferentStructureError(ctx context.Context, request *admin.WorkflowCreateRequest, oldSpec *core.CompiledWorkflowClosure, newSpec *core.CompiledWorkflowClosure) FlyteAdminError {
	errorMsg := fmt.Sprintf("%v workflow with different structure already exists. (Please register a new version of the workflow):\n", request.Id.Name)
	diff, _ := jsondiff.Compare(oldSpec, newSpec)
	rdiff, _ := jsondiff.Compare(newSpec, oldSpec)
	rs := compareJsons(diff, rdiff)

	errorMsg += strings.Join(rs, "\n")

	statusErr, transformationErr := NewFlyteAdminError(codes.InvalidArgument, errorMsg).WithDetails(&admin.CreateWorkflowFailureReason{
		Reason: &admin.CreateWorkflowFailureReason_ExistsDifferentStructure{
			ExistsDifferentStructure: &admin.WorkflowErrorExistsDifferentStructure{
				Id: request.Id,
			},
		},
	})
	if transformationErr != nil {
		logger.Errorf(ctx, "Failed to wrap grpc status in type 'Error': %v", transformationErr)
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
		logger.Errorf(ctx, "Failed to wrap grpc status in type 'Error': %v", transformationErr)
		return NewFlyteAdminErrorf(codes.AlreadyExists, errorMsg)
	}
	return statusErr
}

func NewLaunchPlanExistsDifferentStructureError(ctx context.Context, request *admin.LaunchPlanCreateRequest, oldSpec *admin.LaunchPlanSpec, newSpec *admin.LaunchPlanSpec) FlyteAdminError {
	errorMsg := fmt.Sprintf("%v launch plan with different structure already exists. (Please register a new version of the launch plan):\n", request.Id.Name)
	diff, _ := jsondiff.Compare(oldSpec, newSpec)
	rdiff, _ := jsondiff.Compare(newSpec, oldSpec)
	rs := compareJsons(diff, rdiff)

	errorMsg += strings.Join(rs, "\n")

	return NewFlyteAdminErrorf(codes.InvalidArgument, errorMsg)
}

func NewLaunchPlanExistsIdenticalStructureError(ctx context.Context, request *admin.LaunchPlanCreateRequest) FlyteAdminError {
	errorMsg := "launch plan with identical structure already exists"
	return NewFlyteAdminErrorf(codes.AlreadyExists, errorMsg)
}

func IsDoesNotExistError(err error) bool {
	adminError, ok := err.(FlyteAdminError)
	return ok && adminError.Code() == codes.NotFound
}

func NewInactiveProjectError(ctx context.Context, id string) FlyteAdminError {
	errMsg := fmt.Sprintf("project [%s] is not active", id)
	statusErr, transformationErr := NewFlyteAdminError(codes.InvalidArgument, errMsg).WithDetails(&admin.InactiveProject{
		Id: id,
	})
	if transformationErr != nil {
		logger.Errorf(ctx, "failed to wrap grpc status in type 'Error': %v", transformationErr)
		return NewFlyteAdminErrorf(codes.InvalidArgument, errMsg)
	}
	return statusErr
}
