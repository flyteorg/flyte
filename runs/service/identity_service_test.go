package service

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
)

func TestIdentityService_UserInfo(t *testing.T) {
	svc := NewIdentityService()

	resp, err := svc.UserInfo(context.Background(), connect.NewRequest(&auth.UserInfoRequest{}))
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Msg)
}
