import typing

import grpc

from flytekit.clients.grpc_utils.auth_interceptor import _ClientCallDetails


class DefaultMetadataInterceptor(grpc.UnaryUnaryClientInterceptor, grpc.UnaryStreamClientInterceptor):
    def _inject_default_metadata(self, call_details: grpc.ClientCallDetails):
        metadata = [("accept", "application/grpc")]
        if call_details.metadata:
            metadata.extend(list(call_details.metadata))
        new_details = _ClientCallDetails(
            call_details.method,
            call_details.timeout,
            metadata,
            call_details.credentials,
        )
        return new_details

    def intercept_unary_unary(
        self,
        continuation: typing.Callable,
        client_call_details: grpc.ClientCallDetails,
        request: typing.Any,
    ):
        """
        Intercepts unary calls and inject default metadata
        """
        updated_call_details = self._inject_default_metadata(client_call_details)
        return continuation(updated_call_details, request)

    def intercept_unary_stream(
        self,
        continuation: typing.Callable,
        client_call_details: grpc.ClientCallDetails,
        request: typing.Any,
    ):
        """
        Handles a stream call and inject default metadata
        """
        updated_call_details = self._inject_default_metadata(client_call_details)
        return continuation(updated_call_details, request)
