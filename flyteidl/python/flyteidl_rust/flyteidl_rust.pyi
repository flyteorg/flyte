from flyteidl_rust import *

__all__  = [RawSynchronousFlyteClient]

class RawSynchronousFlyteClient:
    """Main entrypoint for programmatically accessing a Flyte remote backend.

    The term 'remote' is synonymous with 'backend' or 'deployment' and refers to a hosted instance of the
    Flyte platform, which comes with a Flyte Admin server on some known URI.
    """
    def __init__(self, endpoint: str) -> 'RawSynchronousFlyteClient':
        """Initialize a FlyteRemote object.

        Creates a RawSynchronousFlyteClient based on destination endpoint that connects to Flyte gRPC Server.
        
        :param endpoint: destination endpoint of a RawSynchronousFlyteClient to be connect to.
        :return: a RawSynchronousFlyteClient instance with default platform configuration.
        """