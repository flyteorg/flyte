# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc

from flyteidl.cacheservice import cacheservice_pb2 as flyteidl_dot_cacheservice_dot_cacheservice__pb2


class CacheServiceStub(object):
    """
    CacheService defines operations for cache management including retrieval, storage, and deletion of cached task/workflow outputs.
    """

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.Get = channel.unary_unary(
                '/flyteidl.cacheservice.CacheService/Get',
                request_serializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.GetCacheRequest.SerializeToString,
                response_deserializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.GetCacheResponse.FromString,
                )
        self.Put = channel.unary_unary(
                '/flyteidl.cacheservice.CacheService/Put',
                request_serializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.PutCacheRequest.SerializeToString,
                response_deserializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.PutCacheResponse.FromString,
                )
        self.Delete = channel.unary_unary(
                '/flyteidl.cacheservice.CacheService/Delete',
                request_serializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.DeleteCacheRequest.SerializeToString,
                response_deserializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.DeleteCacheResponse.FromString,
                )
        self.GetOrExtendReservation = channel.unary_unary(
                '/flyteidl.cacheservice.CacheService/GetOrExtendReservation',
                request_serializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.GetOrExtendReservationRequest.SerializeToString,
                response_deserializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.GetOrExtendReservationResponse.FromString,
                )
        self.ReleaseReservation = channel.unary_unary(
                '/flyteidl.cacheservice.CacheService/ReleaseReservation',
                request_serializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.ReleaseReservationRequest.SerializeToString,
                response_deserializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.ReleaseReservationResponse.FromString,
                )


class CacheServiceServicer(object):
    """
    CacheService defines operations for cache management including retrieval, storage, and deletion of cached task/workflow outputs.
    """

    def Get(self, request, context):
        """Retrieves cached data by key.
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Put(self, request, context):
        """Stores or updates cached data by key.
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Delete(self, request, context):
        """Deletes cached data by key.
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetOrExtendReservation(self, request, context):
        """Get or extend a reservation for a cache key
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def ReleaseReservation(self, request, context):
        """Release the reservation for a cache key
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_CacheServiceServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'Get': grpc.unary_unary_rpc_method_handler(
                    servicer.Get,
                    request_deserializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.GetCacheRequest.FromString,
                    response_serializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.GetCacheResponse.SerializeToString,
            ),
            'Put': grpc.unary_unary_rpc_method_handler(
                    servicer.Put,
                    request_deserializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.PutCacheRequest.FromString,
                    response_serializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.PutCacheResponse.SerializeToString,
            ),
            'Delete': grpc.unary_unary_rpc_method_handler(
                    servicer.Delete,
                    request_deserializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.DeleteCacheRequest.FromString,
                    response_serializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.DeleteCacheResponse.SerializeToString,
            ),
            'GetOrExtendReservation': grpc.unary_unary_rpc_method_handler(
                    servicer.GetOrExtendReservation,
                    request_deserializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.GetOrExtendReservationRequest.FromString,
                    response_serializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.GetOrExtendReservationResponse.SerializeToString,
            ),
            'ReleaseReservation': grpc.unary_unary_rpc_method_handler(
                    servicer.ReleaseReservation,
                    request_deserializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.ReleaseReservationRequest.FromString,
                    response_serializer=flyteidl_dot_cacheservice_dot_cacheservice__pb2.ReleaseReservationResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'flyteidl.cacheservice.CacheService', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class CacheService(object):
    """
    CacheService defines operations for cache management including retrieval, storage, and deletion of cached task/workflow outputs.
    """

    @staticmethod
    def Get(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/flyteidl.cacheservice.CacheService/Get',
            flyteidl_dot_cacheservice_dot_cacheservice__pb2.GetCacheRequest.SerializeToString,
            flyteidl_dot_cacheservice_dot_cacheservice__pb2.GetCacheResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Put(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/flyteidl.cacheservice.CacheService/Put',
            flyteidl_dot_cacheservice_dot_cacheservice__pb2.PutCacheRequest.SerializeToString,
            flyteidl_dot_cacheservice_dot_cacheservice__pb2.PutCacheResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Delete(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/flyteidl.cacheservice.CacheService/Delete',
            flyteidl_dot_cacheservice_dot_cacheservice__pb2.DeleteCacheRequest.SerializeToString,
            flyteidl_dot_cacheservice_dot_cacheservice__pb2.DeleteCacheResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def GetOrExtendReservation(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/flyteidl.cacheservice.CacheService/GetOrExtendReservation',
            flyteidl_dot_cacheservice_dot_cacheservice__pb2.GetOrExtendReservationRequest.SerializeToString,
            flyteidl_dot_cacheservice_dot_cacheservice__pb2.GetOrExtendReservationResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def ReleaseReservation(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/flyteidl.cacheservice.CacheService/ReleaseReservation',
            flyteidl_dot_cacheservice_dot_cacheservice__pb2.ReleaseReservationRequest.SerializeToString,
            flyteidl_dot_cacheservice_dot_cacheservice__pb2.ReleaseReservationResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)
