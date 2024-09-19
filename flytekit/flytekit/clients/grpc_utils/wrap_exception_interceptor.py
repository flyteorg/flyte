import typing
from typing import Union

import grpc

from flytekit.exceptions.base import FlyteException
from flytekit.exceptions.system import FlyteSystemException
from flytekit.exceptions.user import (
    FlyteAuthenticationException,
    FlyteEntityAlreadyExistsException,
    FlyteEntityNotExistException,
    FlyteInvalidInputException,
)


class RetryExceptionWrapperInterceptor(grpc.UnaryUnaryClientInterceptor, grpc.UnaryStreamClientInterceptor):
    def __init__(self, max_retries: int = 3):
        self._max_retries = 3

    @staticmethod
    def _raise_if_exc(request: typing.Any, e: Union[grpc.Call, grpc.Future]):
        if isinstance(e, grpc.RpcError):
            if e.code() == grpc.StatusCode.UNAUTHENTICATED:
                raise FlyteAuthenticationException() from e
            elif e.code() == grpc.StatusCode.ALREADY_EXISTS:
                raise FlyteEntityAlreadyExistsException() from e
            elif e.code() == grpc.StatusCode.NOT_FOUND:
                raise FlyteEntityNotExistException() from e
            elif e.code() == grpc.StatusCode.INVALID_ARGUMENT:
                raise FlyteInvalidInputException(request) from e
        raise FlyteSystemException() from e

    def intercept_unary_unary(self, continuation, client_call_details, request):
        retries = 0
        while True:
            fut: grpc.Future = continuation(client_call_details, request)
            e = fut.exception()
            try:
                if e:
                    self._raise_if_exc(request, e)
                return fut
            except FlyteException as e:
                if retries == self._max_retries:
                    raise e
                retries = retries + 1

    def intercept_unary_stream(self, continuation, client_call_details, request):
        c: grpc.Call = continuation(client_call_details, request)
        return c
