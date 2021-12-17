.. _divedeep-workflows:

Workflows
=========

A workflow is a directed acyclic graph (DAG) of units of work encapsulated by :ref:`nodes <divedeep-nodes>`. 
Specific instantiations of a workflow (commonly with bound input arguments) are referred to as **workflow executions**, 
or just executions. In other words, a workflow is a template for an ordered task execution.

Flyte workflows are defined in protobuf and the Flytekit SDK facilitates writing workflows. Users can define workflows as a collection of nodes.
Nodes within a workflow can produce outputs that subsequent nodes depend on as inputs. These dependencies dictate the workflow structure.
Workflows written using the SDK do not need to explicitly define nodes to enclose execution units (tasks, sub-workflows, launch plans); 
these will be injected by the SDK and captured at registration time.

Structure
---------

Workflows can accept inputs and/or produce outputs and re-use task definitions across :ref:`projects <divedeep-projects>` and :ref:`domains <divedeep-domains>`.
Workflow structure is flexible because:

- Nodes can be executed in parallel
- The same task definition can be re-used within a different workflow
- A single workflow can contain any combination of task types 
- A workflow can contain a single functional node
- A workflow can contain many nodes in all sorts of arrangements and even launch other workflows

At execution time, node executions will be triggered as soon as their inputs are available. Workflow nodes are naturally run in parallel when possible.

Flyte-specific structure
^^^^^^^^^^^^^^^^^^^^^^^^
During registration, the Flyte platform will validate workflow structure and save the workflow. 
The registration process also updates the workflow graph.
A compiled workflow will always have a start and end node injected into the workflow graph. 
In addition, a failure handler will catch and process execution failures.

Versioning
----------

Like :ref:`tasks <divedeep-tasks>`, workflows are versioned. Registered workflows are also immutable, meaning that an instance of a 
workflow defined by project, domain, name and version cannot be updated. 
The tasks referenced in a workflow version are themselves immutable and are tied to specific tasks' versions.

