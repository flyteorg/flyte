.. _divedeep-workflows:

Workflows
=========

A workflow is a directed acyclic graph (DAG) of units of work encapsulated by nodes. Specific instantiations of a workflow (commonly with bound input arguments)
are referred to as workflow executions, or just executions. In other words, a workflow is a template for ordered task execution.

Flyte workflows are defined in protobuf and the Flyte SDK facilitates writing workflows. Users can define workflows as a collection of nodes.
Nodes within a workflow can produce outputs that subsequent nodes depend on as inputs. These dependencies dictate the workflow structure.
Workflows written using the SDK do not need to explicitly define nodes to enclose execution units (tasks, sub-workflows, launch plans), these will be injected by the SDK and captured at registration time.

Structure
---------

Workflows can accept inputs and/or produce outputs and re-use task definitions across :ref:`projects <concepts-projects>` and :ref:`domains <concepts-domains>`.
Workflow structure is flexible - nodes can be executed in parallel, the same task definition can be re-used within the same workflow. A single workflow
can contain any combination of task types. A workflow can contain a single functional node, it can contain many nodes in all sorts of arrangements and even launch other workflows. 

At execution time, node executions will be triggered as soon as their inputs are available. Workflow nodes are naturally run in parallel when possible.

Flyte-specific structure
^^^^^^^^^^^^^^^^^^^^^^^^
During registration, the Flyte platform will validate workflow structure and save the workflow. The registration process also updates the workflow graph.
A compiled workflow will always have a start and end node injected into the workflow graph. In addition a failure handler will catch and process execution failures.

Versioning
----------
Like :ref:`tasks <concepts-tasks>`, workflows are versioned. Registered workflows are also immutable meaning that an instance of a workflow defined
by project, domain, name and version cannot be updated. The tasks referenced in a workflow version are themselves immutable and are tied to specific tasks' versions.

Executions
----------

A workflow can only be executed through a :ref:`launch plan <concepts-launchplans>`.
A workflow can be launched many times with a variety of launch plans and inputs. Workflows that produce inputs and outputs can take advantage of :ref:`task caching <features-task_cache>` to cache intermediate inputs and outputs and speed-up subsequent executions.

.. _concepts-nodes:

Nodes
=====

A node represents a unit of execution or work within a workflow. Ordinarily a node will encapsulate an instance of a :ref:`task <concepts-tasks>` but a node can also contain an entire subworkflow or trigger a child workflow. Nodes can have inputs and outputs, which are used to coordinate task inputs and outputs.  Node outputs can be used as inputs to other nodes within a workflow.

Tasks are always encapsulated within a node, however, like tasks, nodes can come in a variety of flavors determined by their *target*.
These targets include :ref:`task nodes <concepts-task-nodes>`, :ref:`workflow nodes <concepts-workflow-nodes>`, and :ref:`branch nodes <concepts-branch-nodes>`.

.. _concepts-task-nodes:

Task Nodes
----------

Tasks referenced in a workflow are always enclosed in nodes. This extends to all task types, for example an array task will be enclosed by a single node. For more information about different task types, refer to :ref:`task types <tasktypes>`.

.. _concepts-workflow-nodes:

Workflow Nodes
--------------
A node can contain an entire sub-workflow. Because workflow executions always require a launch plan, workflow nodes have a reference to a launch plan used
to trigger their enclosed workflows.

.. _concepts-branch-nodes:

Branch Nodes
------------
Branch nodes alter the flow of the workflow graph. Conditions at runtime are evaluated to determine control flow.
