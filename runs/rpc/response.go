package rpc

import (
	"errors"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/runs/service"
)

func rpcResponse[T any](resp *T, err error) (*connect.Response[T], error) {
	if err != nil {
		return nil, rpcError(err)
	}
	return connect.NewResponse(resp), nil
}

func rpcError(err error) error {
	var kind service.ErrorKind
	if !errors.As(err, &kind) {
		return err
	}
	switch kind {
	case service.ErrFailedPrecondition:
		return connect.NewError(connect.CodeFailedPrecondition, err)
	case service.ErrAlreadyExists:
		return connect.NewError(connect.CodeAlreadyExists, err)
	case service.ErrInternal:
		return connect.NewError(connect.CodeInternal, err)
	case service.ErrNotFound:
		return connect.NewError(connect.CodeNotFound, err)
	case service.ErrInvalidArgument:
		return connect.NewError(connect.CodeInvalidArgument, err)
	case service.ErrUnimplemented:
		return connect.NewError(connect.CodeUnimplemented, err)
	default:
		return err
	}
}
