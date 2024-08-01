package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	mockScope "github.com/flyteorg/flyte/flytestdlib/promutils"
)

func TestRecoveryInterceptor(t *testing.T) {
	ctx := context.Background()
	testScope := mockScope.NewTestScope()
	recoveryInterceptor := NewRecoveryInterceptor(testScope)
	unaryInterceptor := recoveryInterceptor.UnaryServerInterceptor()
	streamInterceptor := recoveryInterceptor.StreamServerInterceptor()
	unaryInfo := &grpc.UnaryServerInfo{}
	streamInfo := &grpc.StreamServerInfo{}
	req := "test-request"

	t.Run("unary should recover from panic", func(t *testing.T) {
		_, err := unaryInterceptor(ctx, req, unaryInfo, func(ctx context.Context, req any) (any, error) {
			panic("synthetic")
		})
		expectedErr := status.Errorf(codes.Internal, "")
		require.Error(t, err)
		require.Equal(t, expectedErr, err)
	})

	t.Run("stream should recover from panic", func(t *testing.T) {
		stream := testStream{}
		err := streamInterceptor(nil, &stream, streamInfo, func(srv any, stream grpc.ServerStream) error {
			panic("synthetic")
		})
		expectedErr := status.Errorf(codes.Internal, "")
		require.Error(t, err)
		require.Equal(t, expectedErr, err)
	})

	t.Run("unary should plumb response without panic", func(t *testing.T) {
		mockedResponse := "test"
		resp, err := unaryInterceptor(ctx, req, unaryInfo, func(ctx context.Context, req any) (any, error) {
			return mockedResponse, nil
		})
		require.NoError(t, err)
		require.Equal(t, mockedResponse, resp)
	})

	t.Run("stream should plumb response without panic", func(t *testing.T) {
		stream := testStream{}
		handlerCalled := false
		err := streamInterceptor(nil, &stream, streamInfo, func(srv any, stream grpc.ServerStream) error {
			handlerCalled = true
			return nil
		})
		require.NoError(t, err)
		require.True(t, handlerCalled)
	})
}

// testStream is an implementation of grpc.ServerStream for testing.
type testStream struct {
}

func (s *testStream) SendMsg(m interface{}) error {
	return nil
}

func (s *testStream) RecvMsg(m interface{}) error {
	return nil
}

func (s *testStream) SetHeader(metadata.MD) error {
	return nil
}

func (s *testStream) SendHeader(metadata.MD) error {
	return nil
}

func (s *testStream) SetTrailer(metadata.MD) {}

func (s *testStream) Context() context.Context {
	return context.Background()
}
