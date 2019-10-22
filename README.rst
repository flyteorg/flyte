Flyteadmin
=============

Flyteadmin is the control plane for Flyte responsible for managing entities (task, workflows, launch plans) and
administering workflow executions. Flyteadmin implements the
`AdminService <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/service/admin.proto>`_ which
defines a stateless REST/gRPC service for interacting with registered Flyte entities and executions.
Flyteadmin uses a relational style Metadata Store abstracted by `GORM <http://gorm.io/>`_ ORM library.

Before Check-In
~~~~~~~~~~~~~~~

Flyte Admin has a few useful make targets for linting and testing. Please use these before checking in to help suss out
minor bugs and linting errors.

  .. code-block:: console

    make goimports

  .. code-block:: console

    make test_unit

  .. code-block:: console

    make lint
