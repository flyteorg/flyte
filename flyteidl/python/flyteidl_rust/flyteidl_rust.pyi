from flyteidl_rust import *

__all__  = [AdminStub, DataProxyStub]

class AdminStub:
    """Main entrypoint for programmatically accessing a Flyte remote backend.

    The term 'remote' is synonymous with 'backend' or 'deployment' and refers to a hosted instance of the
    Flyte platform, which comes with a Flyte Admin server on some known URI.
    """
    def __init__(self, dst: str) -> 'AdminStub':
        """Initialize a FlyteRemote object.

        Creates a AdminServiceClient based on destination endpoint that connects to Flyte gRPC Client.
        
        :param dst: destination endpoint of a AdminServiceClient to be connect to.
        :return: a AdminServiceClient instance with default platform configuration.
        """

class DataProxyStub:
    """Main entrypoint for programmatically accessing a Flyte remote backend.

    The term 'remote' is synonymous with 'backend' or 'deployment' and refers to a hosted instance of the
    Flyte platform, which comes with a Flyte Admin server on some known URI.
    """
    def __init__(self, dst: str) -> 'DataProxyStub':
        """Initialize a FlyteRemote object.

        Creates a DataProxyServiceClient based on destination endpoint that connects to Flyte gRPC Client.
        
        :param dst: destination endpoint of a DataProxyServiceClient to be connect to.
        :return: a DataProxyServiceClient instance with default platform configuration.
        """