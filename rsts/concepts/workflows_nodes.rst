.. _divedeep-workflows:

Workflows
=========

A workflow is a directed acyclic graph (DAG) of units of work encapsulated by :ref:`nodes <divedeep-nodes>`.
Specific instantiations of a workflow (commonly with bound input arguments) are referred to as **workflow executions**,
or just executions. In other words, a workflow is a template for an ordered task execution.

Flyte workflows are defined in protobuf and the Flytekit SDK facilitates writing workflows. Users can define workflows as a collection of nodes.
Nodes within a workflow can produce outputs that subsequent nodes consume as inputs. These dependencies dictate the workflow structure.
Workflows written using the SDK do not need to explicitly define nodes to enclose execution units (tasks, sub-workflows, launch plans);
these will be injected by the SDK and captured at registration time.

Structure
---------

Workflows can accept inputs and produce outputs and re-use task definitions across :ref:`projects <divedeep-projects>` and :ref:`domains <divedeep-domains>`.
Workflow structure is flexible because:

- Nodes can be executed in parallel
- The same task definition can be re-used within a different workflow
- A single workflow can contain any combination of task types
- A workflow can contain a single functional node
- A workflow can contain many nodes in all sorts of arrangements and it can even launch other workflows

At execution time, node executions will be triggered as soon as their inputs are available.

**Workflow nodes are naturally run in parallel when possible**.
For example, when a workflow has five independent nodes, i.e., when the five nodes do not consume outputs produced by other nodes,
then Flyte runs the nodes in parallel as per the data and resource constraints.

Flyte-specific Structure
^^^^^^^^^^^^^^^^^^^^^^^^

During registration, Flyte will validate the workflow structure and save the workflow.
The registration process also updates the workflow graph.
A compiled workflow will always have a start and end node injected into the workflow graph.
In addition, a failure handler will catch and process execution failures.

Versioning
----------

Like :ref:`tasks <divedeep-tasks>`, workflows are versioned. Registered workflows are also immutable, meaning that an instance of a
workflow defined by project, domain, name and version cannot be updated.
Tasks referenced in a workflow version are themselves immutable and are tied to specific tasks' versions.

Executions
----------

A workflow can only be executed through a :ref:`launch plan <divedeep-launchplans>`.
A workflow can be launched many times with a variety of launch plans and inputs. Workflows that consume inputs and produce
outputs can take advantage of :std:ref:`User Guide <cookbook:sphx_glr_auto_core_flyte_basics_task_cache.py>` to cache
intermediate inputs and outputs and speed-up the subsequent executions.

.. _divedeep-nodes:

Nodes
=====

A node represents a unit of execution or work within a workflow. Ordinarily, a node will encapsulate an instance of
a :ref:`task <divedeep-tasks>`, but it can also contain an entire subworkflow or trigger a child workflow.
Nodes can have inputs and outputs, which would be used to coordinate the task's inputs and outputs.
Moreover, node outputs can be used as inputs to other nodes within a workflow.

Tasks are always encapsulated within a node, however, like tasks, nodes can come in a variety of flavors determined by their *target*.
These targets include :ref:`task nodes <divedeep-task-nodes>`, :ref:`workflow nodes <divedeep-workflow-nodes>`, and :ref:`branch nodes <divedeep-branch-nodes>`.

.. _divedeep-task-nodes:

Task Nodes
----------

Tasks referenced in a workflow are always enclosed in nodes. This extends to all task types.
For example, an array task will be enclosed by a single node.

.. _divedeep-workflow-nodes:

Workflow Nodes
--------------

A node can contain an entire sub-workflow. Since workflow executions always require launch plans,
workflow nodes have reference to launch plans used to trigger their enclosed workflows.

.. _divedeep-branch-nodes:

Branch Nodes
------------

Branch nodes alter the flow of the workflow graph. Conditions at runtime are evaluated to determine the control flow.
