package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mockScope "github.com/flyteorg/flyte/flytestdlib/promutils"
)

func TestRecoveryInterceptor(t *testing.T) {
	ctx := context.Background()
	testScope := mockScope.NewTestScope()
	recoveryInterceptor := NewRecoveryInterceptor(testScope)
	unaryInterceptor := recoveryInterceptor.UnaryServerInterceptor()
	info := &grpc.UnaryServerInfo{}
	req := "test-request"

	t.Run("should recover from panic", func(t *testing.T) {
		_, err := unaryInterceptor(ctx, req, info, func(ctx context.Context, req any) (any, error) {
			panic("synthetic")
		})
		expectedErr := status.Errorf(codes.Internal, "")
		require.Error(t, err)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should plumb response without panic", func(t *testing.T) {
		mockedResponse := "test"
		resp, err := unaryInterceptor(ctx, req, info, func(ctx context.Context, req any) (any, error) {
			return mockedResponse, nil
		})
		require.NoError(t, err)
		require.Equal(t, mockedResponse, resp)
	})
}
