.. _components-admin:

##########
FlyteAdmin
##########

Admin Structure
===============

FlyteAdmin serves the main Flyte API. It processes all client requests to the system. Clients include the FlyteConsole, which calls FlyteAdmin to list workflows, get execution details, etc., and FlyteKit, which calls FlyteAdmin to registering workflows, launch workflows etc.

Below, we'll dive into each component defined in admin in more detail.


RPC
---

FlyteAdmin uses the `grpc-gateway <https://github.com/grpc-ecosystem/grpc-gateway>`__ library to serve
incoming gRPC and HTTP requests with identical handlers. For a more detailed overview of the API,
including request and response entities, see the admin
service `definition <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/service/admin.proto>`__. The
RPC handlers are a thin shim that enforce some request structure validation and call out to appropriate :ref:`manager <components-admin-manager>`.
methods to process requests.

.. _components-admin-manager:

Managers
--------

The Admin API is broken up into entities:

- Executions
- Launch plans
- Node Executions
- Projects (and their respective domains)
- Task Executions
- Tasks
- Workflows

Each API entity has an entity manager in FlyteAdmin reposible for implementing business logic for the entity.
Entity managers handle full validation for create, update and get requests, and
data persistence in the backing store (see the :ref:`repository <components-admin-repository>` section).


Additional Components
+++++++++++++++++++++

The managers utilize additional components to process requests. These additional components include:

- **:ref:`workflow engine <components-admin-workflowengine>`**: compiles workflows and launches workflow executions from launch plans.
- **:ref:`data <components-admin-data>` (remote cloud storage)**: offloads data blobs to the configured cloud provider.
- **:ref:`runtime <components-admin-config>`**: loads values from a config file to assign task resources, initialization values, execution queues and more.
- **:ref:`async processes <components-admin-async>`**: provides functions for scheduling and executing workflows as well as enqueuing and triggering notifications

.. _components-admin-repository:

Repository
----------
Serialized entities (tasks, workflows, launch plans) and executions (workflow-, node- and task-) are stored as protos defined
`here <https://github.com/lyft/flyteidl/tree/master/protos/flyteidl/admin>`__.
We use the excellent `gorm <http://doc.gorm.io/>`__ library to interface with our database, which currently supports a postgres
implementation.  The actual code for issuing queries with gorm can be found in the
`gormimpl <https://github.com/lyft/flyteadmin/blob/master/pkg/repositories/gormimpl>`__ directory.

Models
++++++
Database models are defined in the `models <https://github.com/lyft/flyteadmin/blob/master/pkg/repositories/models>`__ directory and correspond 1:1 with database tables [0]_.

The full set of database tables includes:

- executions
- execution_events
- launch_plans
- node_executions
- node_execution_events
- tasks
- task_executions
- workflows

These database models inherit primary keys and indexes as defined in the corresponding `models <https://github.com/lyft/flyteadmin/blob/master/pkg/repositories/models>`__ file.

The repositories code also includes `transformers <https://github.com/lyft/flyteadmin/blob/master/pkg/repositories/transformers>`__.
These convert entities from the database format to a response format for the external API.
If you change either of these structures, you will find you must change the corresponding transformers.


.. _components-admin-async:

Component Details
=================

This section dives into detail for each top-level directories defined in ``pkg/``.

Asynchronous Components
-----------------------

Notifications and schedules are handled by async routines that are reponsible for enqueing and subsequently processing dequeued messages.

Flyteadmin uses the `gizmo toolkit <https://github.com/nytimes/gizmo>`__ to abstract queueing implementation. Gizmo's
`pubsub <https://github.com/nytimes/gizmo#pubsub>`__ library offers implementations for Amazon SNS/SQS, Google's Pubsub, Kafka topics and publishing over HTTP.

For the sandbox development, no-op implementations of the notifications and schedule handlers are used to remove external cloud dependencies.


Common
------

As the name implies, ``common`` houses shared components used across different flyteadmin components in a single, top-level directory to avoid cyclic dependencies. These components include execution naming and phase utils, query filter definitions, query sorting definitions, and named constants.

.. _components-admin-data:

Data
----

Data interfaces are primarily handled by the `storage <https://github.com/lyft/flytestdlib>`__ library implemented in flytestdlib. However, neither this nor the underlying `stow <https://github.com/graymeta/stow>`__ library expose `HEAD <https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/HEAD>`__ support so the data package in admin exists as the layer responsible for additional, remote data operations.

Errors
------

The errors directory contains centrally defined errors that are designed for compatibility with gRPC statuses.

.. _components-admin-config:

Runtime
-------
Values specific to the flyteadmin application as well as task and workflow registration and execution are configured in the `runtime <https://github.com/lyft/flyteadmin/tree/master/pkg/runtime>`__ directory. These interfaces expose values configured in the ``flyteadmin`` top-level key in the application config.

.. _components-admin-workflowengine:

Workflowengine
--------------

This directory contains interfaces to build and execute workflows leveraging flytepropeller compiler and client components.

.. [0] Unfortunately, given unique naming constraints, some models are redefined in `migration_models <https://github.com/lyft/flyteadmin/blob/master/pkg/repositories/config/migration_models.go>`__ to guarantee unique index values.
