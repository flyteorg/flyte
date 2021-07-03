.. _divedeep-admin:

###########
Flyte Admin
###########

Admin Structure
===============

FlyteAdmin serves the main Flyte API. It processes all client requests to the system. Clients include the FlyteConsole, which calls FlyteAdmin to list workflows, get execution details, etc., and FlyteKit, which calls FlyteAdmin to registering workflows, launch workflows etc.

Below, we'll dive into each component defined in admin in more detail.


RPC
---

FlyteAdmin uses the `grpc-gateway <https://github.com/grpc-ecosystem/grpc-gateway>`__ library to serve
incoming gRPC and HTTP requests with identical handlers. For a more detailed overview of the API,
including request and response entities, see the admin
service :std:ref:`definition <protos/docs/service/service:flyteidl/service/admin.proto>`. The
RPC handlers are a thin shim that enforce some request structure validation and call out to appropriate :ref:`manager <divedeep-admin-manager>`.
methods to process requests.

A detailed explanation of the service can be found in the :ref:`admin service <divedeep-admin-service>` page.

.. _divedeep-admin-manager:

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
data persistence in the backing store (see the :ref:`divedeep-admin-repository` section).


Additional Components
+++++++++++++++++++++

The managers utilize additional components to process requests. These additional components include:

- :ref:`workflow engine <divedeep-admin-workflowengine>`: compiles workflows and launches workflow executions from launch plans.
- :ref:`data <divedeep-admin-data>` (remote cloud storage): offloads data blobs to the configured cloud provider.
- :ref:`runtime <divedeep-admin-config>`: loads values from a config file to assign task resources, initialization values, execution queues and more.
- :ref:`async processes <divedeep-admin-async>`: provides functions for scheduling and executing workflows as well as enqueuing and triggering notifications

.. _divedeep-admin-repository:

Repository
----------
Serialized entities (tasks, workflows, launch plans) and executions (workflow-, node- and task-) are stored as protos defined
`here <https://github.com/flyteorg/flyteidl/tree/master/protos/flyteidl/admin>`__.
We use the excellent `gorm <https://gorm.io/docs/index.html>`__ library to interface with our database, which currently supports a postgres
implementation.  The actual code for issuing queries with gorm can be found in the
`gormimpl <https://github.com/flyteorg/flyteadmin/blob/master/pkg/repositories/gormimpl>`__ directory.

Models
++++++
Database models are defined in the `models <https://github.com/flyteorg/flyteadmin/blob/master/pkg/repositories/models>`__ directory and correspond 1:1 with database tables [0]_.

The full set of database tables includes:

- executions
- execution_events
- launch_plans
- node_executions
- node_execution_events
- tasks
- task_executions
- workflows

These database models inherit primary keys and indexes as defined in the corresponding `models <https://github.com/flyteorg/flyteadmin/blob/master/pkg/repositories/models>`__ file.

The repositories code also includes `transformers <https://github.com/flyteorg/flyteadmin/blob/master/pkg/repositories/transformers>`__.
These convert entities from the database format to a response format for the external API.
If you change either of these structures, you will find you must change the corresponding transformers.


.. _divedeep-admin-async:

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

.. _divedeep-admin-data:

Data
----

Data interfaces are primarily handled by the `storage <https://github.com/flyteorg/flytestdlib>`__ library implemented in flytestdlib. However, neither this nor the underlying `stow <https://github.com/graymeta/stow>`__ library expose `HEAD <https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/HEAD>`__ support so the data package in admin exists as the layer responsible for additional, remote data operations.

Errors
------

The errors directory contains centrally defined errors that are designed for compatibility with gRPC statuses.

.. _divedeep-admin-config:

Runtime
-------
Values specific to the flyteadmin application as well as task and workflow registration and execution are configured in the `runtime <https://github.com/flyteorg/flyteadmin/tree/master/pkg/runtime>`__ directory. These interfaces expose values configured in the ``flyteadmin`` top-level key in the application config.

.. _divedeep-admin-workflowengine:

Workflowengine
--------------

This directory contains interfaces to build and execute workflows leveraging flytepropeller compiler and client components.

.. [0] Unfortunately, given unique naming constraints, some models are redefined in `migration_models <https://github.com/flyteorg/flyteadmin/blob/master/pkg/repositories/config/migration_models.go>`__ to guarantee unique index values.

.. _divedeep-admin-service:


FlyteAdmin Service Background
=============================

Entities
---------

The  :std:ref:`admin service definition <protos/docs/service/service:flyteidl/service/admin.proto>` defines REST operations for the entities
flyteadmin administers.

As a refresher, the primary :ref:`entities <divedeep>` across Flyte map similarly to FlyteAdmin entities.

Static entities
+++++++++++++++

These include:

- Workflows
- Tasks
- Launch Plans

Permitted operations:

- Create
- Get
- List

The above are designated by an :std:ref:`identifier <protos/docs/core/core:identifier>`
which consists of a project, domain, name and version specification. These entities are for the most part immutable. To update one of these specific entities, the updated
version must be reregistered with a unique and new version identifier attribute.

One caveat is that launch plan state can toggle between :std:ref:`ACTIVE or INACTIVE <protos/docs/admin/admin:launchplan>`.
At most one launch plan version across a shared project, domain and name specification can be active at a time. The state affects scheduled launch plans only.
An inactive launch plan can still be used to launch individual executions. However, only an active launch plan runs on a schedule (if it has a schedule defined).


Static entities metadata (Named Entities)
+++++++++++++++++++++++++++++++++++++++++
A :std:ref:`named entity <protos/docs/admin/admin:namedentity>` includes metadata for one of the above entities
(workflow, task or launch plan) across versions. A named entity includes a resource type (workflow, task or launch plan) and an
:std:ref:`id <protos/docs/admin/admin:namedentityidentifier>` which is composed of project, domain and name.
A named entity also includes metadata, which are mutable attributes about the referenced entity.

This metadata includes:

- Description: a human readable description for the Named Entity collection
- State (workflows only): this determines whether the workflow is shown on the overview list of workflows scoped by project and domain

Permitted operations:

- Create
- Update
- Get
- List


Execution entities
++++++++++++++++++

These include:

- (Workflow) executions
- Node executions
- Task executions

Permitted operations:

- Create
- Get
- List

After an execution begins, flyte propeller monitors the execution and sends events which admin uses to update the above executions. 

These :std:ref:`events <protos/docs/event/event:flyteidl/event/event.proto>` include

- WorkflowExecutionEvent
- NodeExecutionEvent
- TaskExecutionEvent

and include information about respective phase transitions, phase transition time and optional output data if the event concerns a terminal phase change.

These events are the **only** way to update an execution. No raw Update endpoint exists.

To track the lifecycle of an execution admin stores attributes such as duration, timestamp at which an execution transitioned to running, and end time.

For debug purposes admin also stores Workflow and Node execution events in its database, but does not currently expose them through an API. Because array tasks can yield very many executions,
admin does **not** store TaskExecutionEvents.


Platform entities
+++++++++++++++++
Projects: like named entities, project have mutable metadata such as human-readable names and descriptions, in addition to their unique string ids.

Permitted project operations:

- Register
- List

Matchable resources
+++++++++++++++++++

A thorough background on :std:ref:`matchable resources <protos/docs/admin/admin:matchableresource>` explains
their purpose and application logic. As a summary, these are used to override system level defaults for kubernetes cluster
resource management, default execution values, and more all across different levels of specificity.

These entities consist of:

- ProjectDomainAttributes
- WorkflowAttributes

Where ProjectDomainAttributes configure customizable overrides at the project and domain level and WorkflowAttributes configure customizable overrides at the project, domain and workflow level.

Permitted attribute operations:

- Update (implicitly creates if there is no existing override)
- Get
- Delete

Using the Admin Service
-----------------------

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
     - States are defined in :std:ref:`launchplanstate <protos/docs/admin/admin:launchplanstate>`.
- Named Entity Metadata

  - state (you must use the integer enum e.g. 1)	
     - States are defined in :std:ref:`namedentitystate <protos/docs/admin/admin:namedentitystate>`.
- Executions (Workflow executions)	

  - project	
  - domain	
  - name	
  - workflow.{any workflow field above} (for example: workflow.domain)	
  - launch_plan.{any launch plan field above} (for example: launch_plan.name)	
  - phase (you must use the upper-cased string name e.g. RUNNING)	
     - Phases are defined in :std:ref:`workflowexecution.phase <protos/docs/core/core:workflowexecution.phase>`.
  - execution_created_at	
  - execution_updated_at	
  - duration (in seconds)	
  - mode (you must use the integer enum e.g. 1)	
     - Modes are defined in :std:ref:`executionmode <protos/docs/admin/admin:executionmetadata.executionmode>`.
  - user (authenticated user or role from flytekit config)

- Node Executions	

  - node_id	
  - execution.{any execution field above} (for example: execution.domain)	
  - phase (you must use the upper-cased string name e.g. QUEUED)	
     - Phases are defined in :std:ref:`nodeexecution.phase <protos/docs/core/core:nodeexecution.phase>`.
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
     - Phases are defined in :std:ref:`taskexecution.phase <protos/docs/core/core:taskexecution.phase>`.
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
     - States are defined in :std:ref:`launchplanstate <protos/docs/admin/admin:launchplanstate>`.
- ListWorkflowIds	

  - project	
  - domain	
- ListExecutions	

  - project	
  - domain	
  - name	
  - phase (you must use the upper-cased string name e.g. RUNNING)	
     - Phases are defined in :std:ref:`workflowexecution.phase <protos/docs/core/core:workflowexecution.phase>`.
  - execution_created_at	
  - execution_updated_at	
  - duration (in seconds)	
  - mode (you must use the integer enum e.g. 1)	
     - Modes are defined :std:ref:`execution.proto <protos/docs/admin/admin:executionmetadata.executionmode>`.
- ListNodeExecutions	

  - node_id	
  - retry_attempt	
  - phase (you must use the upper-cased string name e.g. QUEUED)	
     - Phases are defined in :std:ref:`nodeexecution.phase <protos/docs/core/core:nodeexecution.phase>`.
  - started_at	
  - node_execution_created_at	
  - node_execution_updated_at	
  - duration (in seconds)	
- ListTaskExecutions	

  - retry_attempt	
  - phase (you must use the upper-cased string name e.g. SUCCEEDED)	
     - Phases are defined in :std:ref:`taskexecution.phase <protos/docs/core/core:taskexecution.phase>`.
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

