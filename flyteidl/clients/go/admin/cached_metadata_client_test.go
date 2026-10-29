package admin

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

type countingMetadataClient struct {
	service.AuthMetadataServiceClient
	calls atomic.Int32
	errs  []error
	resp  *service.PublicClientAuthConfigResponse
}

func (c *countingMetadataClient) GetPublicClientConfig(
	_ context.Context,
	_ *service.PublicClientAuthConfigRequest,
	_ ...grpc.CallOption,
) (*service.PublicClientAuthConfigResponse, error) {
	n := int(c.calls.Add(1))
	if n <= len(c.errs) {
		return nil, c.errs[n-1]
	}
	return c.resp, nil
}

func TestCachedMetadataClient_CallsInnerOnceForConcurrentCallers(t *testing.T) {
	inner := &countingMetadataClient{resp: &service.PublicClientAuthConfigResponse{ClientId: "flytectl"}}
	client := NewCachedMetadataClient(inner)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := client.GetPublicClientConfig(context.Background(), &service.PublicClientAuthConfigRequest{})
			assert.NoError(t, err)
			assert.Equal(t, "flytectl", resp.GetClientId())
		}()
	}
	wg.Wait()

	assert.Equal(t, int32(1), inner.calls.Load())
}

func TestCachedMetadataClient_RetriesAfterError(t *testing.T) {
	inner := &countingMetadataClient{
		errs: []error{errors.New("unavailable"), context.DeadlineExceeded},
		resp: &service.PublicClientAuthConfigResponse{ClientId: "flytectl"},
	}
	client := NewCachedMetadataClient(inner)
	ctx := context.Background()
	req := &service.PublicClientAuthConfigRequest{}

	_, err := client.GetPublicClientConfig(ctx, req)
	assert.Error(t, err)
	_, err = client.GetPublicClientConfig(ctx, req)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	resp, err := client.GetPublicClientConfig(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, "flytectl", resp.GetClientId())

	resp, err = client.GetPublicClientConfig(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, "flytectl", resp.GetClientId())
	assert.Equal(t, int32(3), inner.calls.Load())
}
