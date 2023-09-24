.. _divedeep-workflows:

Workflows
=========

.. tags:: Basic, Glossary

A workflow is a directed acyclic graph (DAG) of units of work encapsulated by :ref:`nodes <divedeep-nodes>`.
Specific instantiations of a workflow (commonly bound with input arguments) are referred to as **workflow executions**,
or just executions. In other words, a workflow is a template for an ordered task execution.

Flyte workflows are defined in ``protobuf`` and the flytekit SDK facilitates writing workflows. Users can define workflows as a collection of nodes.
Nodes within a workflow can produce outputs that subsequent nodes could consume as inputs. These dependencies dictate the structure of the workflow.

Workflows written using the SDK don't need to explicitly define nodes to enclose execution units (tasks, sub-workflows, launch plans);
they will be injected by the SDK and captured at registration time.

Structure
---------

Workflows accept inputs and produce outputs and reuse task definitions across :ref:`projects <divedeep-projects>` and :ref:`domains <divedeep-domains>`. Every workflow has a default :ref:`launchplan <divedeep-launchplans>` with the same name as that of the workflow.

Workflow structure is flexible because:

- Nodes can be executed in parallel.
- The same task definition can be re-used within a different workflow.
- A single workflow can contain any combination of task types.
- A workflow can contain a single functional node.
- A workflow can contain multiple nodes in all sorts of arrangements.
- A workflow can launch other workflows.

At execution time, node executions are triggered as soon as their inputs are available.

**Workflow nodes naturally run in parallel when possible**.
For example, when a workflow has five independent nodes, i.e., when these five nodes don't consume outputs produced by other nodes,
Flyte runs these nodes in parallel in accordance with the data and resource constraints.

Flyte-Specific Structure
^^^^^^^^^^^^^^^^^^^^^^^^

During :ref:`registration <divedeep-registration>`, Flyte validates the workflow structure and saves the workflow.
The registration process updates the workflow graph.
A compiled workflow will always have a start and end node injected into the workflow graph.
In addition, a failure handler will catch and process execution failures.

Versioning
----------

Like :ref:`tasks <divedeep-tasks>`, workflows are versioned too. Registered workflows are immutable, i.e., an instance of a
workflow defined by a specific {Project, Domain, Name, Version} combination can't be updated.
Tasks referenced in a workflow version are immutable and are tied to specific tasks' versions.