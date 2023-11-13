.. _divedeep-state-machine:

################################################
Understanding the State Transition in a Workflow
################################################

.. tags:: Basic, Design

High Level Overview of How a Workflow Progresses to Success
===========================================================

.. mermaid::

   flowchart TD
     id1(( ))
     id1 --> Ready
     Ready --> Running
     subgraph Running
     id2(( ))
     id2 --> NodeQueued
     NodeQueued --> NodeRunning
     subgraph NodeRunning
     id3(( ))
     id3 --> TaskQueued
     TaskQueued --> TaskRunning
     TaskRunning --> TaskSuccess
     end
     TaskSuccess --> NodeSuccess
     end
     NodeSuccess --> Success


This state diagram illustrates a high-level, simplistic view of the state transitions that a workflow with a single task and node would go through as the user observes success.

The following sections explain the various observable (and some hidden) states for workflow, node, and task state transitions.

Workflow States
===============

.. mermaid::

   flowchart TD
     Queued -->|On system errors more than threshold| Aborted
     Queued --> Ready
     Ready--> |Write inputs to workflow| Running
     Running--> |On system error| Running
     Running--> |On all Nodes Success| Succeeding
     Succeeding--> |On successful event send to Admin| Succeeded
     Succeeding--> |On system error| Succeeding
     Ready--> |On precondition failure| Failing
     Running--> |On any Node Failure| Failing
     Ready--> |On user initiated abort| Aborting
     Running--> |On user initiated abort| Aborting
     Succeeding--> |On user initiated abort| Aborting
     Failing--> |If Failure node exists| HandleFailureNode
     Failing--> |On user initiated abort| Aborting
     HandleFailureNode--> |On completing failure node| Failed
     HandleFailureNode--> |On user initiated abort| Aborting
     Failing--> |On successful send of Failure node| Failed
     Aborting--> |On successful event send to Admin| Aborted

A workflow always starts in the ``Ready`` state and ends either in ``Failed``, ``Succeeded``, or ``Aborted`` state.
Any system error within a state causes a retry on that state. These retries are capped by :ref:`system retries <system-retry>` which eventually lead to an ``Aborted`` state if the failure persists.

Every transition between states is recorded in FlyteAdmin using :std:ref:`workflowexecutionevent <ref_flyteidl.event.WorkflowExecutionEvent>`.

The phases in the above state diagram are captured in the admin database as specified here :std:ref:`workflowexecution.phase <ref_flyteidl.core.WorkflowExecution.Phase>` and are sent as a part of the Execution event.

The state machine specification for the illustration can be found `here <https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoic3RhdGVEaWFncmFtLXYyXG4gICAgWypdIC0tPiBBYm9ydGVkIDogT24gc3lzdGVtIGVycm9ycyBtb3JlIHRoYW4gdGhyZXNob2xkXG4gICAgWypdIC0tPiBSZWFkeVxuICAgIFJlYWR5IC0tPiBSdW5uaW5nIDogV3JpdGUgaW5wdXRzIHRvIHdvcmtmbG93XG4gICAgUnVubmluZyAtLT4gUnVubmluZyA6IE9uIHN5c3RlbSBlcnJvclxuICAgIFJ1bm5pbmcgLS0-IFN1Y2NlZWRpbmcgOiBPbiBhbGwgTm9kZXMgU3VjY2Vzc1xuICAgIFN1Y2NlZWRpbmcgLS0-IFN1Y2NlZWRlZCA6IE9uIHN1Y2Nlc3NmdWwgZXZlbnQgc2VuZCB0byBBZG1pblxuICAgIFN1Y2NlZWRpbmcgLS0-IFN1Y2NlZWRpbmcgOiBPbiBzeXN0ZW0gZXJyb3JcbiAgICBSZWFkeSAtLT4gRmFpbGluZyA6IE9uIHByZWNvbmRpdGlvbiBmYWlsdXJlXG4gICAgUnVubmluZyAtLT4gRmFpbGluZyA6IE9uIGFueSBOb2RlIEZhaWx1cmVcbiAgICBSZWFkeSAtLT4gQWJvcnRlZCA6IE9uIHVzZXIgaW5pdGlhdGVkIGFib3J0XG4gICAgUnVubmluZyAtLT4gQWJvcnRlZCA6IE9uIHVzZXIgaW5pdGlhdGVkIGFib3J0XG4gICAgU3VjY2VlZGluZyAtLT4gQWJvcnRlZCA6IE9uIHVzZXIgaW5pdGlhdGVkIGFib3J0XG5cbiAgICBGYWlsaW5nIC0tPiBIYW5kbGVGYWlsdXJlTm9kZSA6IElmIEZhaWx1cmUgbm9kZSBleGlzdHNcbiAgICBGYWlsaW5nIC0tPiBBYm9ydGVkIDogT24gdXNlciBpbml0aWF0ZWQgYWJvcnRcbiAgICBIYW5kbGVGYWlsdXJlTm9kZSAtLT4gRmFpbGVkIDogT24gY29tcGxldGluZyBmYWlsdXJlIG5vZGVcbiAgICBIYW5kbGVGYWlsdXJlTm9kZSAtLT4gQWJvcnRlZCA6IE9uIHVzZXIgaW5pdGlhdGVkIGFib3J0XG4gICAgRmFpbGluZyAtLT4gRmFpbGVkIDogT24gc3VjY2Vzc2Z1bCBzZW5kIG9mIEZhaWx1cmUgbm9kZVxuICAgICIsIm1lcm1haWQiOnt9LCJ1cGRhdGVFZGl0b3IiOmZhbHNlfQ>`__.


Node States
===========

.. mermaid::

   flowchart TD
     id1(( ))
     id1-->NotYetStarted
     id1-->|Will stop the node execution |Aborted
     NotYetStarted-->|If all upstream nodes are ready, i.e, inputs are ready | Queued
     NotYetStarted--> |If the branch was not taken |Skipped
     Queued-->|Start task execution- attempt 0 | Running
     Running-->|If task timeout has elapsed and retry_attempts >= max_retries|TimingOut
     Running-->|Internal state|Succeeding
     Running-->|For dynamic nodes generating workflows| DynamicRunning
     DynamicRunning-->TimingOut
     DynamicRunning-->RetryableFailure
     TimingOut-->|If total node timeout has elapsed|TimedOut
     DynamicRunning-->Succeeding
     Succeeding-->|User observes the task as succeeded| Succeeded
     Running-->|on retryable failure| RetryableFailure
     RetryableFailure-->|if retry_attempts < max_retries|Running
     RetryableFailure-->|retry_attempts >= max_retries|Failing
     Failing-->Failed
     Succeeded-->id2(( ))
     Failed-->id2(( ))


This state diagram illustrates the node transition through various states. This is the core finite state machine for a node.
From the user's perspective, a workflow simply consists of a sequence of tasks. But to Flyte, a workflow internally creates a meta entity known as **node**.

Once a Workflow enters the ``Running`` state, it triggers the phantom ``start node`` of the workflow. The ``start node`` is considered to be the entry node of any workflow. 
The ``start node`` begins by executing all its child-nodes using a modified Depth First Search algorithm recursively.

Nodes can be of different types as listed below, but all the nodes traverse through the same transitions:

#. Start Node - Only exists during the execution and is not modeled in the core spec.
#. :std:ref:`Task Node <ref_flyteidl.core.TaskNode>`
#. :std:ref:`Branch Node <ref_flyteidl.core.BranchNode>`
#. :std:ref:`Workflow Node <ref_flyteidl.core.WorkflowNode>`
#. Dynamic Node - Just a task node that does not return output but constitutes a dynamic workflow. 
   When the task runs, it remains in the ``RUNNING`` state. Once the task completes and Flyte starts executing the dynamic workflow, 
   the overarching node that contains both the original task and the dynamic workflow enters `DYNAMIC_RUNNING` state.
#. End Node - Only exists during the execution and is not modeled in the core spec

Every transition between states is recorded in FlyteAdmin using :std:ref:`nodeexecutionevent <ref_flyteidl.event.NodeExecutionEvent>`.

Every ``NodeExecutionEvent`` can have any :std:ref:`nodeexecution.phase <ref_flyteidl.core.NodeExecution.Phase>`.

.. note:: TODO: Add explanation for each phase.

The state machine specification for the illustration can be found `here <https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoic3RhdGVEaWFncmFtLXYyXG4gICAgWypdIC0tPiBOb3RZZXRTdGFydGVkXG4gICAgWypdIC0tPiBBYm9ydGVkIDogV2lsbCBzdG9wIHRoZSBub2RlIGV4ZWN1dGlvblxuICAgIE5vdFlldFN0YXJ0ZWQgLS0-IFF1ZXVlZCA6IElmIGFsbCB1cHN0cmVhbSBub2RlcyBhcmUgcmVhZHkgaS5lLCBpbnB1dHMgYXJlIHJlYWR5XG4gICAgTm90WWV0U3RhcnRlZCAtLT4gU2tpcHBlZCA6IElmIHRoZSBicmFuY2ggd2FzIG5vdCB0YWtlblxuICAgIFF1ZXVlZCAtLT4gUnVubmluZyA6IFN0YXJ0IHRhc2sgZXhlY3V0aW9uIC0gYXR0ZW1wdCAwXG4gICAgUnVubmluZyAtLT4gVGltaW5nT3V0IDogSWYgdGFzayB0aW1lb3V0IGhhcyBlbGFwc2VkIGFuZCByZXRyeV9hdHRlbXB0cyA-PSBtYXhfcmV0cmllc1xuICAgIFRpbWluZ091dCAtLT4gVGltZWRPdXQgOiBJdCB0b3RhbCBub2RlIHRpbWVvdXQgaGFzIGVsYXBzZWRcbiAgICBSdW5uaW5nIC0tPiBSZXRyeWFibGVGYWlsdXJlIDogb24gcmV0cnlhYmxlIGZhaWx1cmVcbiAgICBSdW5uaW5nIC0tPiBEeW5hbWljUnVubmluZyA6IEZvciBkeW5hbWljIG5vZGVzIGdlbmVyYXRpbmcgd29ya2Zsb3dzXG4gICAgUmV0cnlhYmxlRmFpbHVyZSAtLT4gUnVubmluZyA6IGlmIHJldHJ5X2F0dGVtcHRzIDwgbWF4X3JldHJpZXNcbiAgICBSZXRyeWFibGVGYWlsdXJlIC0tPiBGYWlsaW5nIDogcmV0cnlfYXR0ZW1wdHMgPj0gbWF4X3JldHJpZXNcbiAgICBGYWlsaW5nIC0tPiBGYWlsZWRcbiAgICBSdW5uaW5nIC0tPiBTdWNjZWVkaW5nIDogSW50ZXJuYWwgc3RhdGVcbiAgICBEeW5hbWljUnVubmluZyAtLT4gU3VjY2VlZGluZ1xuICAgIER5bmFtaWNSdW5uaW5nIC0tPiBSZXRyeWFibGVGYWlsdXJlXG4gICAgRHluYW1pY1J1bm5pbmcgLS0-IFRpbWluZ091dFxuICAgIFN1Y2NlZWRpbmcgLS0-IFN1Y2NlZWRlZCA6IFVzZXIgb2JzZXJ2ZXMgdGhlIHRhc2sgYXMgc3VjY2VlZGVkXG4gICAgU3VjY2VlZGVkIC0tPiBbKl1cbiAgICBGYWlsZWQgLS0-IFsqXVxuIiwibWVybWFpZCI6e30sInVwZGF0ZUVkaXRvciI6ZmFsc2V9>`__.

Task States
===========

.. mermaid::

   flowchart TD
     id1(( ))
     id1-->|Aborted by NodeHandler- timeouts, external abort, etc,.| NotReady
     id1-->Aborted
     NotReady-->|Optional-Blocked on resource quota or resource pool | WaitingForResources
     WaitingForResources--> |Optional- Has been submitted, but hasn't started |Queued
     Queued-->|Optional- Prestart initialization | Initializing
     Initializing-->|Actual execution of user code has started|Running
     Running-->|Successful execution|Success
     Running-->|Failed with a retryable error|RetryableFailure
     Running-->|Unrecoverable failure, will stop all execution|PermanentFailure
     Success-->id2(( ))
     RetryableFailure-->id2(( ))
     PermanentFailure-->id2(( ))


The state diagram above illustrates the various states through which a task transitions. This is the core finite state machine for a task.

Every transition between states is recorded in FlyteAdmin using :std:ref:`taskexecutionevent <ref_flyteidl.event.TaskExecutionEvent>`.

Every ``TaskExecutionEvent`` can have any :std:ref:`taskexecution.phase <ref_flyteidl.core.TaskExecution.Phase>`.

.. note:: TODO: Add explanation for each phase.

The state machine specification for the illustration can be found `here <https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoic3RhdGVEaWFncmFtLXYyXG4gICAgWypdIC0tPiBOb3RSZWFkeVxuICAgIFsqXSAtLT4gQWJvcnRlZCA6IEFib3J0ZWQgYnkgTm9kZUhhbmRsZXIgLSB0aW1lb3V0cywgZXh0cmVuYWwgYWJvcnQsIGV0Y1xuICAgIE5vdFJlYWR5IC0tPiBXYWl0aW5nRm9yUmVzb3VyY2VzIDogQmxvY2tlZCBvbiByZXNvdXJjZSBxdW90YSBvciByZXNvdXJjZSBwb29sIChvcHRpb25hbClcbiAgICBXYWl0aW5nRm9yUmVzb3VyY2VzIC0tPiBRdWV1ZWQgOiBIYXMgYmVlbiBzdWJtaXR0ZWQsIGJ1dCBoYXMgbm90IHN0YXJ0ZWQgKG9wdGlvbmFsKVxuICAgIFF1ZXVlZCAtLT4gSW5pdGlhbGl6aW5nIDogUHJlc3RhcnQgaW5pdGlhbGl6YXRpb24gKG9wdGlvbmFsKVxuICAgIEluaXRpYWxpemluZyAtLT4gUnVubmluZyA6IEFjdHVhbCBleGVjdXRpb24gb2YgdXNlciBjb2RlIGhhcyBzdGFydGVkXG4gICAgUnVubmluZyAtLT4gU3VjY2VzcyA6IFN1Y2Nlc3NmdWwgZXhlY3V0aW9uXG4gICAgUnVubmluZyAtLT4gUmV0cnlhYmxlRmFpbHVyZSA6IEZhaWxlZCB3aXRoIGEgcmV0cnlhYmxlIGVycm9yXG4gICAgUnVubmluZyAtLT4gUGVybWFuZW50RmFpbHVyZSA6IFVucmVjb3ZlcmFibGUgZmFpbHVyZSwgd2lsbCBzdG9wIGFsbCBleGVjdXRpb25cbiAgICBTdWNjZXNzIC0tPiBbKl1cbiAgICBSZXRyeWFibGVGYWlsdXJlIC0tPiBbKl1cbiAgICBQZXJtYW5lbnRGYWlsdXJlIC0tPiBbKl1cbiIsIm1lcm1haWQiOnt9LCJ1cGRhdGVFZGl0b3IiOmZhbHNlfQ>`__.
