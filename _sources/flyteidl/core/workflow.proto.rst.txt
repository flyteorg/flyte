.. _api_file_flyteidl/core/workflow.proto:

workflow.proto
============================

.. _api_msg_flyteidl.core.IfBlock:

flyteidl.core.IfBlock
---------------------

`[flyteidl.core.IfBlock proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/workflow.proto#L15>`_

Defines a condition and the execution unit that should be executed if the condition is satisfied.

.. code-block:: json

  {
    "condition": "{...}",
    "then_node": "{...}"
  }

.. _api_field_flyteidl.core.IfBlock.condition:

condition
  (:ref:`flyteidl.core.BooleanExpression <api_msg_flyteidl.core.BooleanExpression>`) 
  
.. _api_field_flyteidl.core.IfBlock.then_node:

then_node
  (:ref:`flyteidl.core.Node <api_msg_flyteidl.core.Node>`) 
  


.. _api_msg_flyteidl.core.IfElseBlock:

flyteidl.core.IfElseBlock
-------------------------

`[flyteidl.core.IfElseBlock proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/workflow.proto#L22>`_

Defines a series of if/else blocks. The first branch whose condition evaluates to true is the one to execute.
If no conditions were satisfied, the else_node or the error will execute.

.. code-block:: json

  {
    "case": "{...}",
    "other": [],
    "else_node": "{...}",
    "error": "{...}"
  }

.. _api_field_flyteidl.core.IfElseBlock.case:

case
  (:ref:`flyteidl.core.IfBlock <api_msg_flyteidl.core.IfBlock>`) required. First condition to evaluate.
  
  
.. _api_field_flyteidl.core.IfElseBlock.other:

other
  (:ref:`flyteidl.core.IfBlock <api_msg_flyteidl.core.IfBlock>`) optional. Additional branches to evaluate.
  
  
.. _api_field_flyteidl.core.IfElseBlock.else_node:

else_node
  (:ref:`flyteidl.core.Node <api_msg_flyteidl.core.Node>`) The node to execute in case none of the branches were taken.
  
  required.
  
  
  Only one of :ref:`else_node <api_field_flyteidl.core.IfElseBlock.else_node>`, :ref:`error <api_field_flyteidl.core.IfElseBlock.error>` may be set.
  
.. _api_field_flyteidl.core.IfElseBlock.error:

error
  (:ref:`flyteidl.core.Error <api_msg_flyteidl.core.Error>`) An error to throw in case none of the branches were taken.
  
  required.
  
  
  Only one of :ref:`else_node <api_field_flyteidl.core.IfElseBlock.else_node>`, :ref:`error <api_field_flyteidl.core.IfElseBlock.error>` may be set.
  


.. _api_msg_flyteidl.core.BranchNode:

flyteidl.core.BranchNode
------------------------

`[flyteidl.core.BranchNode proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/workflow.proto#L41>`_

BranchNode is a special node that alter the flow of the workflow graph. It allows the control flow to branch at
runtime based on a series of conditions that get evaluated on various parameters (e.g. inputs, primtives).

.. code-block:: json

  {
    "if_else": "{...}"
  }

.. _api_field_flyteidl.core.BranchNode.if_else:

if_else
  (:ref:`flyteidl.core.IfElseBlock <api_msg_flyteidl.core.IfElseBlock>`) required
  
  


.. _api_msg_flyteidl.core.TaskNode:

flyteidl.core.TaskNode
----------------------

`[flyteidl.core.TaskNode proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/workflow.proto#L47>`_

Refers to the task that the Node is to execute.

.. code-block:: json

  {
    "reference_id": "{...}"
  }

.. _api_field_flyteidl.core.TaskNode.reference_id:

reference_id
  (:ref:`flyteidl.core.Identifier <api_msg_flyteidl.core.Identifier>`) A globally unique identifier for the task.
  
  
  


.. _api_msg_flyteidl.core.WorkflowNode:

flyteidl.core.WorkflowNode
--------------------------

`[flyteidl.core.WorkflowNode proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/workflow.proto#L55>`_

Refers to a the workflow the node is to execute.

.. code-block:: json

  {
    "launchplan_ref": "{...}",
    "sub_workflow_ref": "{...}"
  }

.. _api_field_flyteidl.core.WorkflowNode.launchplan_ref:

launchplan_ref
  (:ref:`flyteidl.core.Identifier <api_msg_flyteidl.core.Identifier>`) A globally unique identifier for the launch plan.
  
  
  
  Only one of :ref:`launchplan_ref <api_field_flyteidl.core.WorkflowNode.launchplan_ref>`, :ref:`sub_workflow_ref <api_field_flyteidl.core.WorkflowNode.sub_workflow_ref>` may be set.
  
.. _api_field_flyteidl.core.WorkflowNode.sub_workflow_ref:

sub_workflow_ref
  (:ref:`flyteidl.core.Identifier <api_msg_flyteidl.core.Identifier>`) Reference to a subworkflow, that should be defined with the compiler context
  
  
  
  Only one of :ref:`launchplan_ref <api_field_flyteidl.core.WorkflowNode.launchplan_ref>`, :ref:`sub_workflow_ref <api_field_flyteidl.core.WorkflowNode.sub_workflow_ref>` may be set.
  


.. _api_msg_flyteidl.core.NodeMetadata:

flyteidl.core.NodeMetadata
--------------------------

`[flyteidl.core.NodeMetadata proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/workflow.proto#L66>`_

Defines extra information about the Node.

.. code-block:: json

  {
    "name": "...",
    "timeout": "{...}",
    "retries": "{...}",
    "interruptible": "..."
  }

.. _api_field_flyteidl.core.NodeMetadata.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) A friendly name for the Node
  
  
.. _api_field_flyteidl.core.NodeMetadata.timeout:

timeout
  (:ref:`google.protobuf.Duration <api_msg_google.protobuf.Duration>`) The overall timeout of a task.
  
  
.. _api_field_flyteidl.core.NodeMetadata.retries:

retries
  (:ref:`flyteidl.core.RetryStrategy <api_msg_flyteidl.core.RetryStrategy>`) Number of retries per task.
  
  
.. _api_field_flyteidl.core.NodeMetadata.interruptible:

interruptible
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  Identify whether node is interruptible
  
  


.. _api_msg_flyteidl.core.Alias:

flyteidl.core.Alias
-------------------

`[flyteidl.core.Alias proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/workflow.proto#L84>`_

Links a variable to an alias.

.. code-block:: json

  {
    "var": "...",
    "alias": "..."
  }

.. _api_field_flyteidl.core.Alias.var:

var
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Must match one of the output variable names on a node.
  
  
.. _api_field_flyteidl.core.Alias.alias:

alias
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) A workflow-level unique alias that downstream nodes can refer to in their input.
  
  


.. _api_msg_flyteidl.core.Node:

flyteidl.core.Node
------------------

`[flyteidl.core.Node proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/workflow.proto#L94>`_

A Workflow graph Node. One unit of execution in the graph. Each node can be linked to a Task, a Workflow or a branch
node.

.. code-block:: json

  {
    "id": "...",
    "metadata": "{...}",
    "inputs": [],
    "upstream_node_ids": [],
    "output_aliases": [],
    "task_node": "{...}",
    "workflow_node": "{...}",
    "branch_node": "{...}"
  }

.. _api_field_flyteidl.core.Node.id:

id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) A workflow-level unique identifier that identifies this node in the workflow. "inputs" and "outputs" are reserved
  node ids that cannot be used by other nodes.
  
  
.. _api_field_flyteidl.core.Node.metadata:

metadata
  (:ref:`flyteidl.core.NodeMetadata <api_msg_flyteidl.core.NodeMetadata>`) Extra metadata about the node.
  
  
.. _api_field_flyteidl.core.Node.inputs:

inputs
  (:ref:`flyteidl.core.Binding <api_msg_flyteidl.core.Binding>`) Specifies how to bind the underlying interface's inputs. All required inputs specified in the underlying interface
  must be fullfilled.
  
  
.. _api_field_flyteidl.core.Node.upstream_node_ids:

upstream_node_ids
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) optional Specifies execution depdendency for this node ensuring it will only get scheduled to run after all its
  upstream nodes have completed. This node will have an implicit depdendency on any node that appears in inputs
  field.
  
  
.. _api_field_flyteidl.core.Node.output_aliases:

output_aliases
  (:ref:`flyteidl.core.Alias <api_msg_flyteidl.core.Alias>`) optional. A node can define aliases for a subset of its outputs. This is particularly useful if different nodes
  need to conform to the same interface (e.g. all branches in a branch node). Downstream nodes must refer to this
  nodes outputs using the alias if one's specified.
  
  
.. _api_field_flyteidl.core.Node.task_node:

task_node
  (:ref:`flyteidl.core.TaskNode <api_msg_flyteidl.core.TaskNode>`) Information about the Task to execute in this node.
  
  Information about the target to execute in this node.
  
  
  Only one of :ref:`task_node <api_field_flyteidl.core.Node.task_node>`, :ref:`workflow_node <api_field_flyteidl.core.Node.workflow_node>`, :ref:`branch_node <api_field_flyteidl.core.Node.branch_node>` may be set.
  
.. _api_field_flyteidl.core.Node.workflow_node:

workflow_node
  (:ref:`flyteidl.core.WorkflowNode <api_msg_flyteidl.core.WorkflowNode>`) Information about the Workflow to execute in this mode.
  
  Information about the target to execute in this node.
  
  
  Only one of :ref:`task_node <api_field_flyteidl.core.Node.task_node>`, :ref:`workflow_node <api_field_flyteidl.core.Node.workflow_node>`, :ref:`branch_node <api_field_flyteidl.core.Node.branch_node>` may be set.
  
.. _api_field_flyteidl.core.Node.branch_node:

branch_node
  (:ref:`flyteidl.core.BranchNode <api_msg_flyteidl.core.BranchNode>`) Information about the branch node to evaluate in this node.
  
  Information about the target to execute in this node.
  
  
  Only one of :ref:`task_node <api_field_flyteidl.core.Node.task_node>`, :ref:`workflow_node <api_field_flyteidl.core.Node.workflow_node>`, :ref:`branch_node <api_field_flyteidl.core.Node.branch_node>` may be set.
  


.. _api_msg_flyteidl.core.WorkflowMetadata:

flyteidl.core.WorkflowMetadata
------------------------------

`[flyteidl.core.WorkflowMetadata proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/workflow.proto#L131>`_

Metadata for the entire workflow.
To be used in the future.

.. code-block:: json

  {
    "queuing_budget": "{...}"
  }

.. _api_field_flyteidl.core.WorkflowMetadata.queuing_budget:

queuing_budget
  (:ref:`google.protobuf.Duration <api_msg_google.protobuf.Duration>`) Total wait time a workflow can be delayed by queueing.
  
  


.. _api_msg_flyteidl.core.WorkflowMetadataDefaults:

flyteidl.core.WorkflowMetadataDefaults
--------------------------------------

`[flyteidl.core.WorkflowMetadataDefaults proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/workflow.proto#L137>`_

Default Workflow Metadata for the entire workflow.

.. code-block:: json

  {
    "interruptible": "..."
  }

.. _api_field_flyteidl.core.WorkflowMetadataDefaults.interruptible:

interruptible
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Identify whether workflow is interruptible.
  The value set at the workflow level will be the defualt value used for nodes
  unless explicitly set at the node level.
  
  


.. _api_msg_flyteidl.core.WorkflowTemplate:

flyteidl.core.WorkflowTemplate
------------------------------

`[flyteidl.core.WorkflowTemplate proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/workflow.proto#L146>`_

Flyte Workflow Structure that encapsulates task, branch and subworkflow nodes to form a statically analyzable,
directed acyclic graph.

.. code-block:: json

  {
    "id": "{...}",
    "metadata": "{...}",
    "interface": "{...}",
    "nodes": [],
    "outputs": [],
    "failure_node": "{...}",
    "metadata_defaults": "{...}"
  }

.. _api_field_flyteidl.core.WorkflowTemplate.id:

id
  (:ref:`flyteidl.core.Identifier <api_msg_flyteidl.core.Identifier>`) A globally unique identifier for the workflow.
  
  
.. _api_field_flyteidl.core.WorkflowTemplate.metadata:

metadata
  (:ref:`flyteidl.core.WorkflowMetadata <api_msg_flyteidl.core.WorkflowMetadata>`) Extra metadata about the workflow.
  
  
.. _api_field_flyteidl.core.WorkflowTemplate.interface:

interface
  (:ref:`flyteidl.core.TypedInterface <api_msg_flyteidl.core.TypedInterface>`) Defines a strongly typed interface for the Workflow. This can include some optional parameters.
  
  
.. _api_field_flyteidl.core.WorkflowTemplate.nodes:

nodes
  (:ref:`flyteidl.core.Node <api_msg_flyteidl.core.Node>`) A list of nodes. In addition, "globals" is a special reserved node id that can be used to consume workflow inputs.
  
  
.. _api_field_flyteidl.core.WorkflowTemplate.outputs:

outputs
  (:ref:`flyteidl.core.Binding <api_msg_flyteidl.core.Binding>`) A list of output bindings that specify how to construct workflow outputs. Bindings can pull node outputs or
  specify literals. All workflow outputs specified in the interface field must be bound in order for the workflow
  to be validated. A workflow has an implicit dependency on all of its nodes to execute successfully in order to
  bind final outputs.
  Most of these outputs will be Binding's with a BindingData of type OutputReference.  That is, your workflow can
  just have an output of some constant (`Output(5)`), but usually, the workflow will be pulling
  outputs from the output of a task.
  
  
.. _api_field_flyteidl.core.WorkflowTemplate.failure_node:

failure_node
  (:ref:`flyteidl.core.Node <api_msg_flyteidl.core.Node>`) optional A catch-all node. This node is executed whenever the execution engine determines the workflow has failed.
  The interface of this node must match the Workflow interface with an additional input named "error" of type
  pb.lyft.flyte.core.Error.
  
  
.. _api_field_flyteidl.core.WorkflowTemplate.metadata_defaults:

metadata_defaults
  (:ref:`flyteidl.core.WorkflowMetadataDefaults <api_msg_flyteidl.core.WorkflowMetadataDefaults>`) workflow defaults
  
  

