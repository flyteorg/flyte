package errors

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

func TestGrpcStatusError(t *testing.T) {

	msg := "some error"
	curPhase := "some phase"
	statusErr := NewAlreadyInTerminalStateError(context.Background(), msg, curPhase)
	assert.NotNil(t, statusErr)
	s, ok := status.FromError(statusErr)
	assert.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, s.Code())
	assert.Equal(t, "some error", s.Message())

	details, ok := s.Details()[0].(*admin.EventFailureReason)
	assert.True(t, ok)
	_, ok = details.GetReason().(*admin.EventFailureReason_AlreadyInTerminalState)
	assert.True(t, ok)
}

func TestNewIncompatibleClusterError(t *testing.T) {
	errorMsg := "foo"
	cluster := "C1"
	statusErr := NewIncompatibleClusterError(context.Background(), errorMsg, cluster)
	assert.NotNil(t, statusErr)
	s, ok := status.FromError(statusErr)
	assert.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, s.Code())
	assert.Equal(t, errorMsg, s.Message())

	details, ok := s.Details()[0].(*admin.EventFailureReason)
	assert.True(t, ok)
	_, ok = details.GetReason().(*admin.EventFailureReason_IncompatibleCluster)
	assert.True(t, ok)
}
