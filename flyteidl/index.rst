.. flyteidl documentation master file, created by

Flyteidl: Flyte's Core Language Specification
==============================================

``Flyteidl`` contains the core language specification of Flyte, using `Google's Protocol Buffers <https://developers.google.com/protocol-buffers>`_. 
The Specification contains:

#. The core specification for Flyte workflows, tasks, and the type system
#. The specification for FlyteAdmin's `gRPC <https://grpc.io/>`_ and ``REST`` endpoints
#. Some of the core plugin APIs like - Spark, Sagemaker, etc.

This specification is used to generate client stubs for `Flytekit <https://flyte.readthedocs.io/projects/flytekit>`_, `Flytekit Java <https://github.com/spotify/flytekit-java>`_, `Flytectl <https://github.com/flyteorg/flytectl>`_ and the `FlyteAdmin Service <https://pkg.go.dev/github.com/lyft/flyteadmin>`_.


.. toctree::
   :maxdepth: 1
   :hidden:

   |plane| Getting Started <https://docs.flyte.org/en/latest/getting_started.html>
   |book-reader| User Guide <https://docs.flyte.org/projects/cookbook/en/latest/user_guide.html>
   |chalkboard| Tutorials <https://docs.flyte.org/projects/cookbook/en/latest/tutorials.html>
   |project-diagram| Concepts <https://docs.flyte.org/en/latest/concepts/basics.html>
   |rocket| Deployment <https://docs.flyte.org/en/latest/deployment/index.html>
   |book| API Reference <https://docs.flyte.org/en/latest/reference/index.html>
   |hands-helping| Community <https://docs.flyte.org/en/latest/community/index.html>

.. NOTE: the caption text is important for the sphinx theme to correctly render the nav header
.. https://github.com/flyteorg/furo
.. toctree::
   :maxdepth: -1
   :caption: FlyteIDL
   :hidden:

   Flyteidl <self>
   protos/index
   Contributing Guide <README>
