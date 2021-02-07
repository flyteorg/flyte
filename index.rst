.. flyteidl documentation master file, created by

Welcome to ``FlyteIDL - Flytes Languages Specification`` documentation!
=========================================================================
FlyteIDL contains the core specification of Flyte, using `Googles Protocol Buffers <https://developers.google.com/protocol-buffers>`_. The Specification contains

#. The core specification for Flyte Workflows, tasks and the type system
#. The specification for FlyteAdmin's `gRPC <https://grpc.io/>`_ and ``REST`` endpoints
#. Some of the core plugin API's like - Spark, Sagemaker, etc

This specification is used to generate client stubs for `FlyteKit <https://flyte.readthedocs.io/projects/flytekit/en/master/>`_, `FlyteKit Java <https://github.com/spotify/flytekit-java>`_, `Flytectl <https://github.com/flyteorg/flytectl>`_ and the `FlyteAdmin Service <https://pkg.go.dev/github.com/lyft/flyteadmin>`_.

.. toctree::
   :maxdepth: 1
   :caption: Flytekit Learn by Example

   Flyte Documentation <https://flyte.readthedocs.io/en/latest/>
   Flytekit Learn by Example <http://flytecookbook.readthedocs.io/>

.. toctree::
   :maxdepth: 2
   :caption: FlyteIDL API reference & Development

   gen/pb-protodoc/flyteidl/index
   developing
   

Indices and tables
==================

* :ref:`genindex`
* :ref:`modindex`
* :ref:`search`
