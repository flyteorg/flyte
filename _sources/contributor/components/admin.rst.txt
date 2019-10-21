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


Admin Service
=============

Making Requests to FlyteAdmin
-----------------------------

Adding request filters	
++++++++++++++++++++++	

We use `gRPC Gateway <https://github.com/grpc-ecosystem/grpc-gateway>`_ to reverse proxy http requests into gRPC.	
While this allows for a single implementation for both HTTP and gRPC, an important limitation is that fields mapped to the path pattern cannot be	
repeated and must have a primitive (non-message) type. Unfortunately this means that repeated string filters cannot use a proper protobuf message. Instead use	
the internal syntax shown below::	

 func(field,value) or func(field, value)	

For example, multiple filters would be appended to an http request::	

 ?filters=ne(version, TheWorst)+eq(workflow.name, workflow)	

Timestamp fields use the RFC3339Nano spec (ex: "2006-01-02T15:04:05.999999999Z07:00")	

The fully supported set of filter functions are	

- contains	
- gt (greater than)	
- gte (greter than or equal to)	
- lt (less than)	
- lte (less than or equal to)	
- eq (equal)	
- ne (not equal)	
- value_in (for repeated sets of values)	

"value_in" is a special case where multiple values are passed to the filter expression. For example::	

 value_in(phase, 1;2;3)	

Filterable fields vary based on entity types:	

- Task	

  - project	
  - domain	
  - name	
  - version	
  - created_at	
- Workflow	

  - project	
  - domain	
  - name	
  - version	
  - created_at	
- Launch plans	

  - project	
  - domain	
  - name	
  - version	
  - created_at	
  - updated_at	
  - workflows.{any workflow field above} (for example: workflow.domain)	
  - state (you must use the integer enum e.g. 1)	
     - States are defined in `launch_plan.proto <https://github.com/lyft/flyteidl/blob/2c17791170ece1cced3e96daa08ea2692efe3d07/protos/flyteidl/admin/launch_plan.proto#L23>`_.	
- Executions (Workflow executions)	

  - project	
  - domain	
  - name	
  - workflow.{any workflow field above} (for example: workflow.domain)	
  - launch_plan.{any launch plan field above} (for example: launch_plan.name)	
  - phase (you must use the upper-cased string name e.g. RUNNING)	
     - Phases are defined in `execution.proto <https://github.com/lyft/flyteidl/blob/223537e15e05bc6925403a8c11c5a09d91008a80/protos/flyteidl/core/execution.proto#L11,L21>`_.	
  - execution_created_at	
  - execution_updated_at	
  - duration (in seconds)	
  - mode (you must use the integer enum e.g. 1)	
     - Modes are defined in `execution.proto <https://github.com/lyft/flyteidl/blob/182eeb9e1d0f2369e479a981f30cb51c2d7a0672/protos/flyteidl/admin/execution.proto#L96>`_.	

- Node Executions	

  - node_id	
  - execution.{any execution field above} (for example: execution.domain)	
  - phase (you must use the upper-cased string name e.g. QUEUED)	
     - Phases are defined in `execution.proto <https://github.com/lyft/flyteidl/blob/223537e15e05bc6925403a8c11c5a09d91008a80/protos/flyteidl/core/execution.proto#L26,L36>`_.	
  - started_at	
  - node_execution_created_at	
  - node_execution_updated_at	
  - duration (in seconds)	
- Task Executions	

  - retry_attempt	
  - task.{any task field above} (for example: task.version)	
  - execution.{any execution field above} (for example: execution.domain)	
  - node_execution.{any node execution field above} (for example: node_execution.phase)	
  - phase (you must use the upper-cased string name e.g. SUCCEEDED)	
     - Phases are defined in `execution.proto <https://github.com/lyft/flyteidl/blob/223537e15e05bc6925403a8c11c5a09d91008a80/protos/flyteidl/core/execution.proto#L42,L49>`_.	
  - started_at	
  - task_execution_created_at	
  - task_execution_updated_at	
  - duration (in seconds)	

Putting it all together	
-----------------------	

If you wanted to do query on specific executions that were launched with a specific launch plan for a workflow with specific attributes, you could do something like:	

::	

   gte(duration, 100)+value_in(phase,RUNNING;SUCCEEDED;FAILED)+eq(lauch_plan.project, foo)	
   +eq(launch_plan.domain, bar)+eq(launch_plan.name, baz)	
   +eq(launch_plan.version, 1234)	
   +lte(workflow.created_at,2018-11-29T17:34:05.000000000Z07:00)	
   	
   	

Adding sorting to requests	
++++++++++++++++++++++++++	

Only a subset of fields are supported for sorting list queries. The explicit list is below:	

- ListTasks	

  - project	
  - domain	
  - name	
  - version	
  - created_at	
- ListTaskIds	

  - project	
  - domain	
- ListWorkflows	

  - project	
  - domain	
  - name	
  - version	
  - created_at	
- ListWorkflowIds	

  - project	
  - domain	
- ListLaunchPlans	

  - project	
  - domain	
  - name	
  - version	
  - created_at	
  - updated_at	
  - state (you must use the integer enum e.g. 1)	
     - States are defined in `launch_plan.proto <https://github.com/lyft/flyteidl/blob/2c17791170ece1cced3e96daa08ea2692efe3d07/protos/flyteidl/admin/launch_plan.proto#L23>`_.	
- ListWorkflowIds	

  - project	
  - domain	
- ListExecutions	

  - project	
  - domain	
  - name	
  - phase (you must use the upper-cased string name e.g. RUNNING)	
     - Phases are defined in `execution.proto <https://github.com/lyft/flyteidl/blob/223537e15e05bc6925403a8c11c5a09d91008a80/protos/flyteidl/core/execution.proto#L11,L21>`_.	
  - execution_created_at	
  - execution_updated_at	
  - duration (in seconds)	
  - mode (you must use the integer enum e.g. 1)	
     - Modes are defined `execution.proto <https://github.com/lyft/flyteidl/blob/182eeb9e1d0f2369e479a981f30cb51c2d7a0672/protos/flyteidl/admin/execution.proto#L96>`_.	
- ListNodeExecutions	

  - node_id	
  - retry_attempt	
  - phase (you must use the upper-cased string name e.g. QUEUED)	
     - Phases are defined in `execution.proto <https://github.com/lyft/flyteidl/blob/223537e15e05bc6925403a8c11c5a09d91008a80/protos/flyteidl/core/execution.proto#L26,L36>`_.	
  - started_at	
  - node_execution_created_at	
  - node_execution_updated_at	
  - duration (in seconds)	
- ListTaskExecutions	

  - retry_attempt	
  - phase (you must use the upper-cased string name e.g. SUCCEEDED)	
     - Phases are defined in `execution.proto <https://github.com/lyft/flyteidl/blob/223537e15e05bc6925403a8c11c5a09d91008a80/protos/flyteidl/core/execution.proto#L42,L49>`_.	
  - started_at	
  - task_execution_created_at	
  - task_execution_updated_at	
  - duration (in seconds)	

Sorting syntax	
--------------	

Adding sorting to a request requires specifying the ``key``, e.g. the attribute you wish to sort on. Sorting can also optionally specify the direction (one of ``ASCENDING`` or ``DESCENDING``) where ``DESCENDING`` is the default.	

Example sorting http param:	

::	

   sort_by.key=created_at&sort_by.direction=DESCENDING	
   	
Alternatively, since descending is the default, the above could be rewritten as	

::	

   sort_by.key=created_at


.. [0] Unfortunately, given unique naming constraints, some models are redefined in `migration_models <https://github.com/lyft/flyteadmin/blob/master/pkg/repositories/config/migration_models.go>`__ to guarantee unique index values.
