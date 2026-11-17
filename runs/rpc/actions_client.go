package rpc

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect"
)

// AbortAction adapts the actions client for the abort reconciler. An action
// already removed from Kubernetes has reached the requested outcome.
func AbortAction(client actionsconnect.ActionsServiceClient) func(context.Context, *actions.AbortRequest) error {
	return func(ctx context.Context, req *actions.AbortRequest) error {
		_, err := client.Abort(ctx, connect.NewRequest(req))
		if connect.CodeOf(err) == connect.CodeNotFound {
			return nil
		}
		var rpcErr *connect.Error
		if errors.As(err, &rpcErr) && rpcErr.Code() == connect.CodeInternal && strings.Contains(rpcErr.Message(), "not found") {
			return nil
		}
		return err
	}
}
