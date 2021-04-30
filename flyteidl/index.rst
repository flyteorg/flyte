.. flyteidl documentation master file, created by

``FlyteIDL``: Flyte's Core Language Specification
===================================================

``FlyteIDL`` contains the core language specification of Flyte, using `Google's Protocol Buffers <https://developers.google.com/protocol-buffers>`_. The Specification contains

#. The core specification for Flyte Workflows, tasks and the type system
#. The specification for FlyteAdmin's `gRPC <https://grpc.io/>`_ and ``REST`` endpoints
#. Some of the core plugin API's like - Spark, Sagemaker, etc

This specification is used to generate client stubs for `FlyteKit <https://flyte.readthedocs.io/projects/flytekit>`_, `FlyteKit Java <https://github.com/spotify/flytekit-java>`_, `Flytectl <https://github.com/flyteorg/flytectl>`_ and the `FlyteAdmin Service <https://pkg.go.dev/github.com/lyft/flyteadmin>`_.


.. toctree::
   :maxdepth: 1
   :hidden:

   Getting Started <https://docs.flyte.org/en/latest/getting_started.html>
   User Guide <https://docs.flyte.org/projects/cookbook/en/latest/user_guide.html>
   Tutorials <https://docs.flyte.org/projects/cookbook/en/latest/tutorials.html>
   Concepts <https://docs.flyte.org/en/latest/concepts/basics.html>
   API Reference <https://docs.flyte.org/en/latest/reference/index.html>
   Community <https://docs.flyte.org/en/latest/community/index.html>

.. toctree::
   :maxdepth: -1
   :caption: FlyteIDL
   :hidden:

   Overview <self>
   gen/pb-protodoc/flyteidl/index
   developing
