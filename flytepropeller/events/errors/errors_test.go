package errors

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type testErrorWithReason struct {
	status *status.Status
}

func (e *testErrorWithReason) GRPCStatus() *status.Status {
	return e.status
}

func (e *testErrorWithReason) Error() string {
	return e.status.Message()
}

func createTestErrorWithReason() error {
	alreadyInTerminalPhase := &admin.EventErrorAlreadyInTerminalState{CurrentPhase: "some phase"}
	reason := &admin.EventFailureReason{Reason: &admin.EventFailureReason_AlreadyInTerminalState{AlreadyInTerminalState: alreadyInTerminalPhase}}

	s := status.New(codes.FailedPrecondition, "some test")
	s, _ = s.WithDetails(reason)
	return &testErrorWithReason{s}
}

func isEventError(err error) bool {
	_, ok := err.(*EventError)
	return ok
}

func isUnknownError(err error) bool {
	_, isGrpcErr := status.FromError(err)
	isEventErr := isEventError(err)
	return !isEventErr && !isGrpcErr
}

// Test that we wrap the gRPC error to our correct one
func TestWrapErrors(t *testing.T) {
	incompatibleClusterErr, _ := status.New(codes.FailedPrecondition, "incompat").WithDetails(&admin.EventFailureReason{
		Reason: &admin.EventFailureReason_IncompatibleCluster{
			IncompatibleCluster: &admin.EventErrorIncompatibleCluster{
				Cluster: "c1",
			},
		},
	})

	tests := []struct {
		name         string
		inputError   error
		expectedFunc func(error) bool
	}{
		{"alreadyExists", status.Error(codes.AlreadyExists, "Already exists"), IsAlreadyExists},
		{"invalidArgs", status.Error(codes.InvalidArgument, "Invalid Arguments"), IsInvalidArguments},
		{"resourceExhausted", status.Error(codes.ResourceExhausted, "Limit Exceeded"), IsResourceExhausted},
		{"uncaughtError", status.Error(codes.Unknown, "Unknown Err"), isEventError},
		{"uncaughtError", fmt.Errorf("Random err"), isUnknownError},
		{"errorWithReason", createTestErrorWithReason(), IsEventAlreadyInTerminalStateError},
		{"incompatibleCluster", incompatibleClusterErr.Err(), IsEventIncompatibleClusterError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wrappedErr := WrapError(test.inputError)
			assert.True(t, test.expectedFunc(wrappedErr))
		})
	}
}
