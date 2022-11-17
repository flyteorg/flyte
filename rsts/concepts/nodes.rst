.. _divedeep-nodes:

Nodes
=====

.. tags:: Basic, Glossary

A node represents a unit of execution or work within a workflow. Ordinarily, a node encapsulates an instance of 
a :ref:`task <divedeep-tasks>`, but it can also contain an entire subworkflow or trigger an external workflow. 
Nodes can have inputs and outputs, which are used to coordinate task inputs and outputs. 
Moreover, node outputs can be used as inputs to other nodes within a workflow.

Tasks are always encapsulated within a node. Like tasks, nodes can come in a variety of flavors determined by their *target*.
These targets include :ref:`task nodes <divedeep-task-nodes>`, :ref:`workflow nodes <divedeep-workflow-nodes>`, and :ref:`branch nodes <divedeep-branch-nodes>`.

.. _divedeep-task-nodes:

Task Nodes
----------

Tasks referenced in a workflow are always enclosed in nodes. This extends to all task types. 
For example, an array task will be enclosed by a single node.

.. _divedeep-workflow-nodes:

Workflow Nodes
--------------
A node can contain an entire sub-workflow. Since workflow executions always require a launch plan, workflow nodes have a reference to a launch plan to trigger their enclosed workflows.

.. _divedeep-branch-nodes:

Branch Nodes
------------
Branch nodes alter the flow of the workflow graph. Conditions at runtime are evaluated to determine the control flow.
