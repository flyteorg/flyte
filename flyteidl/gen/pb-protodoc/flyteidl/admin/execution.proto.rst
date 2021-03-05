.. _api_file_flyteidl/admin/execution.proto:

execution.proto
==============================

.. _api_msg_flyteidl.admin.ExecutionCreateRequest:

flyteidl.admin.ExecutionCreateRequest
-------------------------------------

`[flyteidl.admin.ExecutionCreateRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/execution.proto#L14>`_

Request to launch an execution with the given project, domain and optionally name.

.. code-block:: json

  {
    "project": "...",
    "domain": "...",
    "name": "...",
    "spec": "{...}",
    "inputs": "{...}"
  }

.. _api_field_flyteidl.admin.ExecutionCreateRequest.project:

project
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Name of the project the execution belongs to.
  
  
.. _api_field_flyteidl.admin.ExecutionCreateRequest.domain:

domain
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Name of the domain the execution belongs to.
  A domain can be considered as a subset within a specific project.
  
  
.. _api_field_flyteidl.admin.ExecutionCreateRequest.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) User provided value for the resource.
  If none is provided the system will generate a unique string.
  +optional
  
  
.. _api_field_flyteidl.admin.ExecutionCreateRequest.spec:

spec
  (:ref:`flyteidl.admin.ExecutionSpec <api_msg_flyteidl.admin.ExecutionSpec>`) Additional fields necessary to launch the execution.
  
  
.. _api_field_flyteidl.admin.ExecutionCreateRequest.inputs:

inputs
  (:ref:`flyteidl.core.LiteralMap <api_msg_flyteidl.core.LiteralMap>`) The inputs required to start the execution. All required inputs must be
  included in this map. If not required and not provided, defaults apply.
  
  


.. _api_msg_flyteidl.admin.ExecutionRelaunchRequest:

flyteidl.admin.ExecutionRelaunchRequest
---------------------------------------

`[flyteidl.admin.ExecutionRelaunchRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/execution.proto#L36>`_

Request to relaunch the referenced execution.

.. code-block:: json

  {
    "id": "{...}",
    "name": "..."
  }

.. _api_field_flyteidl.admin.ExecutionRelaunchRequest.id:

id
  (:ref:`flyteidl.core.WorkflowExecutionIdentifier <api_msg_flyteidl.core.WorkflowExecutionIdentifier>`) Identifier of the workflow execution to relaunch.
  
  
.. _api_field_flyteidl.admin.ExecutionRelaunchRequest.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) User provided value for the relaunched execution.
  If none is provided the system will generate a unique string.
  +optional
  
  


.. _api_msg_flyteidl.admin.ExecutionCreateResponse:

flyteidl.admin.ExecutionCreateResponse
--------------------------------------

`[flyteidl.admin.ExecutionCreateResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/execution.proto#L48>`_

The unique identifier for a successfully created execution.
If the name was *not* specified in the create request, this identifier will include a generated name.

.. code-block:: json

  {
    "id": "{...}"
  }

.. _api_field_flyteidl.admin.ExecutionCreateResponse.id:

id
  (:ref:`flyteidl.core.WorkflowExecutionIdentifier <api_msg_flyteidl.core.WorkflowExecutionIdentifier>`) 
  


.. _api_msg_flyteidl.admin.WorkflowExecutionGetRequest:

flyteidl.admin.WorkflowExecutionGetRequest
------------------------------------------

`[flyteidl.admin.WorkflowExecutionGetRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/execution.proto#L53>`_

A message used to fetch a single workflow execution entity.

.. code-block:: json

  {
    "id": "{...}"
  }

.. _api_field_flyteidl.admin.WorkflowExecutionGetRequest.id:

id
  (:ref:`flyteidl.core.WorkflowExecutionIdentifier <api_msg_flyteidl.core.WorkflowExecutionIdentifier>`) Uniquely identifies an individual workflow execution.
  
  


.. _api_msg_flyteidl.admin.Execution:

flyteidl.admin.Execution
------------------------

`[flyteidl.admin.Execution proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/execution.proto#L61>`_

A workflow execution represents an instantiated workflow, including all inputs and additional
metadata as well as computed results included state, outputs, and duration-based attributes.
Used as a response object used in Get and List execution requests.

.. code-block:: json

  {
    "id": "{...}",
    "spec": "{...}",
    "closure": "{...}"
  }

.. _api_field_flyteidl.admin.Execution.id:

id
  (:ref:`flyteidl.core.WorkflowExecutionIdentifier <api_msg_flyteidl.core.WorkflowExecutionIdentifier>`) Unique identifier of the workflow execution.
  
  
.. _api_field_flyteidl.admin.Execution.spec:

spec
  (:ref:`flyteidl.admin.ExecutionSpec <api_msg_flyteidl.admin.ExecutionSpec>`) User-provided configuration and inputs for launching the execution.
  
  
.. _api_field_flyteidl.admin.Execution.closure:

closure
  (:ref:`flyteidl.admin.ExecutionClosure <api_msg_flyteidl.admin.ExecutionClosure>`) Execution results.
  
  


.. _api_msg_flyteidl.admin.ExecutionList:

flyteidl.admin.ExecutionList
----------------------------

`[flyteidl.admin.ExecutionList proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/execution.proto#L73>`_

Used as a response for request to list executions.

.. code-block:: json

  {
    "executions": [],
    "token": "..."
  }

.. _api_field_flyteidl.admin.ExecutionList.executions:

executions
  (:ref:`flyteidl.admin.Execution <api_msg_flyteidl.admin.Execution>`) 
  
.. _api_field_flyteidl.admin.ExecutionList.token:

token
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) In the case of multiple pages of results, the server-provided token can be used to fetch the next page
  in a query. If there are no more results, this value will be empty.
  
  


.. _api_msg_flyteidl.admin.LiteralMapBlob:

flyteidl.admin.LiteralMapBlob
-----------------------------

`[flyteidl.admin.LiteralMapBlob proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/execution.proto#L82>`_

Input/output data can represented by actual values or a link to where values are stored

.. code-block:: json

  {
    "values": "{...}",
    "uri": "..."
  }

.. _api_field_flyteidl.admin.LiteralMapBlob.values:

values
  (:ref:`flyteidl.core.LiteralMap <api_msg_flyteidl.core.LiteralMap>`) Data in LiteralMap format
  
  
  
  Only one of :ref:`values <api_field_flyteidl.admin.LiteralMapBlob.values>`, :ref:`uri <api_field_flyteidl.admin.LiteralMapBlob.uri>` may be set.
  
.. _api_field_flyteidl.admin.LiteralMapBlob.uri:

uri
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) In the event that the map is too large, we return a uri to the data
  
  
  
  Only one of :ref:`values <api_field_flyteidl.admin.LiteralMapBlob.values>`, :ref:`uri <api_field_flyteidl.admin.LiteralMapBlob.uri>` may be set.
  


.. _api_msg_flyteidl.admin.AbortMetadata:

flyteidl.admin.AbortMetadata
----------------------------

`[flyteidl.admin.AbortMetadata proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/execution.proto#L92>`_


.. code-block:: json

  {
    "cause": "...",
    "principal": "..."
  }

.. _api_field_flyteidl.admin.AbortMetadata.cause:

cause
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) In the case of a user-specified abort, this will pass along the user-supplied cause.
  
  
.. _api_field_flyteidl.admin.AbortMetadata.principal:

principal
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Identifies the entity (if any) responsible for terminating the execution
  
  


.. _api_msg_flyteidl.admin.ExecutionClosure:

flyteidl.admin.ExecutionClosure
-------------------------------

`[flyteidl.admin.ExecutionClosure proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/execution.proto#L101>`_

Encapsulates the results of the Execution

.. code-block:: json

  {
    "outputs": "{...}",
    "error": "{...}",
    "abort_cause": "...",
    "abort_metadata": "{...}",
    "computed_inputs": "{...}",
    "phase": "...",
    "started_at": "{...}",
    "duration": "{...}",
    "created_at": "{...}",
    "updated_at": "{...}",
    "notifications": [],
    "workflow_id": "{...}"
  }

.. _api_field_flyteidl.admin.ExecutionClosure.outputs:

outputs
  (:ref:`flyteidl.admin.LiteralMapBlob <api_msg_flyteidl.admin.LiteralMapBlob>`) A map of outputs in the case of a successful execution.
  
  A result produced by a terminated execution.
  A pending (non-terminal) execution will not have any output result.
  
  
  Only one of :ref:`outputs <api_field_flyteidl.admin.ExecutionClosure.outputs>`, :ref:`error <api_field_flyteidl.admin.ExecutionClosure.error>`, :ref:`abort_cause <api_field_flyteidl.admin.ExecutionClosure.abort_cause>`, :ref:`abort_metadata <api_field_flyteidl.admin.ExecutionClosure.abort_metadata>` may be set.
  
.. _api_field_flyteidl.admin.ExecutionClosure.error:

error
  (:ref:`flyteidl.core.ExecutionError <api_msg_flyteidl.core.ExecutionError>`) Error information in the case of a failed execution.
  
  A result produced by a terminated execution.
  A pending (non-terminal) execution will not have any output result.
  
  
  Only one of :ref:`outputs <api_field_flyteidl.admin.ExecutionClosure.outputs>`, :ref:`error <api_field_flyteidl.admin.ExecutionClosure.error>`, :ref:`abort_cause <api_field_flyteidl.admin.ExecutionClosure.abort_cause>`, :ref:`abort_metadata <api_field_flyteidl.admin.ExecutionClosure.abort_metadata>` may be set.
  
.. _api_field_flyteidl.admin.ExecutionClosure.abort_cause:

abort_cause
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) In the case of a user-specified abort, this will pass along the user-supplied cause.
  
  A result produced by a terminated execution.
  A pending (non-terminal) execution will not have any output result.
  
  
  Only one of :ref:`outputs <api_field_flyteidl.admin.ExecutionClosure.outputs>`, :ref:`error <api_field_flyteidl.admin.ExecutionClosure.error>`, :ref:`abort_cause <api_field_flyteidl.admin.ExecutionClosure.abort_cause>`, :ref:`abort_metadata <api_field_flyteidl.admin.ExecutionClosure.abort_metadata>` may be set.
  
.. _api_field_flyteidl.admin.ExecutionClosure.abort_metadata:

abort_metadata
  (:ref:`flyteidl.admin.AbortMetadata <api_msg_flyteidl.admin.AbortMetadata>`) In the case of a user-specified abort, this will pass along the user and their supplied cause.
  
  A result produced by a terminated execution.
  A pending (non-terminal) execution will not have any output result.
  
  
  Only one of :ref:`outputs <api_field_flyteidl.admin.ExecutionClosure.outputs>`, :ref:`error <api_field_flyteidl.admin.ExecutionClosure.error>`, :ref:`abort_cause <api_field_flyteidl.admin.ExecutionClosure.abort_cause>`, :ref:`abort_metadata <api_field_flyteidl.admin.ExecutionClosure.abort_metadata>` may be set.
  
.. _api_field_flyteidl.admin.ExecutionClosure.computed_inputs:

computed_inputs
  (:ref:`flyteidl.core.LiteralMap <api_msg_flyteidl.core.LiteralMap>`) Inputs computed and passed for execution.
  computed_inputs depends on inputs in ExecutionSpec, fixed and default inputs in launch plan
  
  
.. _api_field_flyteidl.admin.ExecutionClosure.phase:

phase
  (:ref:`flyteidl.core.WorkflowExecution.Phase <api_enum_flyteidl.core.WorkflowExecution.Phase>`) Most recent recorded phase for the execution.
  
  
.. _api_field_flyteidl.admin.ExecutionClosure.started_at:

started_at
  (:ref:`google.protobuf.Timestamp <api_msg_google.protobuf.Timestamp>`) Reported ime at which the execution began running.
  
  
.. _api_field_flyteidl.admin.ExecutionClosure.duration:

duration
  (:ref:`google.protobuf.Duration <api_msg_google.protobuf.Duration>`) The amount of time the execution spent running.
  
  
.. _api_field_flyteidl.admin.ExecutionClosure.created_at:

created_at
  (:ref:`google.protobuf.Timestamp <api_msg_google.protobuf.Timestamp>`) Reported time at which the execution was created.
  
  
.. _api_field_flyteidl.admin.ExecutionClosure.updated_at:

updated_at
  (:ref:`google.protobuf.Timestamp <api_msg_google.protobuf.Timestamp>`) Reported time at which the execution was last updated.
  
  
.. _api_field_flyteidl.admin.ExecutionClosure.notifications:

notifications
  (:ref:`flyteidl.admin.Notification <api_msg_flyteidl.admin.Notification>`) The notification settings to use after merging the CreateExecutionRequest and the launch plan
  notification settings. An execution launched with notifications will always prefer that definition
  to notifications defined statically in a launch plan.
  
  
.. _api_field_flyteidl.admin.ExecutionClosure.workflow_id:

workflow_id
  (:ref:`flyteidl.core.Identifier <api_msg_flyteidl.core.Identifier>`) Identifies the workflow definition for this execution.
  
  


.. _api_msg_flyteidl.admin.SystemMetadata:

flyteidl.admin.SystemMetadata
-----------------------------

`[flyteidl.admin.SystemMetadata proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/execution.proto#L147>`_

Represents system rather than user-facing metadata about an execution.

.. code-block:: json

  {
    "execution_cluster": "..."
  }

.. _api_field_flyteidl.admin.SystemMetadata.execution_cluster:

execution_cluster
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Which execution cluster this execution ran on.
  
  


.. _api_msg_flyteidl.admin.ExecutionMetadata:

flyteidl.admin.ExecutionMetadata
--------------------------------

`[flyteidl.admin.ExecutionMetadata proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/execution.proto#L155>`_

Represents attributes about an execution which are not required to launch the execution but are useful to record.
These attributes are assigned at launch time and do not change.

.. code-block:: json

  {
    "mode": "...",
    "principal": "...",
    "nesting": "...",
    "scheduled_at": "{...}",
    "parent_node_execution": "{...}",
    "reference_execution": "{...}",
    "system_metadata": "{...}"
  }

.. _api_field_flyteidl.admin.ExecutionMetadata.mode:

mode
  (:ref:`flyteidl.admin.ExecutionMetadata.ExecutionMode <api_enum_flyteidl.admin.ExecutionMetadata.ExecutionMode>`) 
  
.. _api_field_flyteidl.admin.ExecutionMetadata.principal:

principal
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Identifier of the entity that triggered this execution.
  For systems using back-end authentication any value set here will be discarded in favor of the
  authenticated user context.
  
  
.. _api_field_flyteidl.admin.ExecutionMetadata.nesting:

nesting
  (`uint32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Indicates the "nestedness" of this execution.
  If a user launches a workflow execution, the default nesting is 0.
  If this execution further launches a workflow (child workflow), the nesting level is incremented by 0 => 1
  Generally, if workflow at nesting level k launches a workflow then the child workflow will have
  nesting = k + 1.
  
  
.. _api_field_flyteidl.admin.ExecutionMetadata.scheduled_at:

scheduled_at
  (:ref:`google.protobuf.Timestamp <api_msg_google.protobuf.Timestamp>`) For scheduled executions, the requested time for execution for this specific schedule invocation.
  
  
.. _api_field_flyteidl.admin.ExecutionMetadata.parent_node_execution:

parent_node_execution
  (:ref:`flyteidl.core.NodeExecutionIdentifier <api_msg_flyteidl.core.NodeExecutionIdentifier>`) Which subworkflow node launched this execution
  
  
.. _api_field_flyteidl.admin.ExecutionMetadata.reference_execution:

reference_execution
  (:ref:`flyteidl.core.WorkflowExecutionIdentifier <api_msg_flyteidl.core.WorkflowExecutionIdentifier>`) Optional, a reference workflow execution related to this execution.
  In the case of a relaunch, this references the original workflow execution.
  
  
.. _api_field_flyteidl.admin.ExecutionMetadata.system_metadata:

system_metadata
  (:ref:`flyteidl.admin.SystemMetadata <api_msg_flyteidl.admin.SystemMetadata>`) Optional, platform-specific metadata about the execution.
  In this the future this may be gated behind an ACL or some sort of authorization.
  
  

.. _api_enum_flyteidl.admin.ExecutionMetadata.ExecutionMode:

Enum flyteidl.admin.ExecutionMetadata.ExecutionMode
---------------------------------------------------

`[flyteidl.admin.ExecutionMetadata.ExecutionMode proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/execution.proto#L157>`_

The method by which this execution was launched.

.. _api_enum_value_flyteidl.admin.ExecutionMetadata.ExecutionMode.MANUAL:

MANUAL
  *(DEFAULT)* ⁣The default execution mode, MANUAL implies that an execution was launched by an individual.
  
  
.. _api_enum_value_flyteidl.admin.ExecutionMetadata.ExecutionMode.SCHEDULED:

SCHEDULED
  ⁣A schedule triggered this execution launch.
  
  
.. _api_enum_value_flyteidl.admin.ExecutionMetadata.ExecutionMode.SYSTEM:

SYSTEM
  ⁣A system process was responsible for launching this execution rather an individual.
  
  
.. _api_enum_value_flyteidl.admin.ExecutionMetadata.ExecutionMode.RELAUNCH:

RELAUNCH
  ⁣This execution was launched with identical inputs as a previous execution.
  
  
.. _api_enum_value_flyteidl.admin.ExecutionMetadata.ExecutionMode.CHILD_WORKFLOW:

CHILD_WORKFLOW
  ⁣This execution was triggered by another execution.
  
  

.. _api_msg_flyteidl.admin.NotificationList:

flyteidl.admin.NotificationList
-------------------------------

`[flyteidl.admin.NotificationList proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/execution.proto#L202>`_


.. code-block:: json

  {
    "notifications": []
  }

.. _api_field_flyteidl.admin.NotificationList.notifications:

notifications
  (:ref:`flyteidl.admin.Notification <api_msg_flyteidl.admin.Notification>`) 
  


.. _api_msg_flyteidl.admin.ExecutionSpec:

flyteidl.admin.ExecutionSpec
----------------------------

`[flyteidl.admin.ExecutionSpec proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/execution.proto#L208>`_

An ExecutionSpec encompasses all data used to launch this execution. The Spec does not change over the lifetime
of an execution as it progresses across phase changes..

.. code-block:: json

  {
    "launch_plan": "{...}",
    "inputs": "{...}",
    "metadata": "{...}",
    "notifications": "{...}",
    "disable_all": "...",
    "labels": "{...}",
    "annotations": "{...}",
    "auth_role": "{...}",
    "security_context": "{...}",
    "quality_of_service": "{...}"
  }

.. _api_field_flyteidl.admin.ExecutionSpec.launch_plan:

launch_plan
  (:ref:`flyteidl.core.Identifier <api_msg_flyteidl.core.Identifier>`) Launch plan to be executed
  
  
.. _api_field_flyteidl.admin.ExecutionSpec.inputs:

inputs
  (:ref:`flyteidl.core.LiteralMap <api_msg_flyteidl.core.LiteralMap>`) Input values to be passed for the execution
  
  
.. _api_field_flyteidl.admin.ExecutionSpec.metadata:

metadata
  (:ref:`flyteidl.admin.ExecutionMetadata <api_msg_flyteidl.admin.ExecutionMetadata>`) Metadata for the execution
  
  
.. _api_field_flyteidl.admin.ExecutionSpec.notifications:

notifications
  (:ref:`flyteidl.admin.NotificationList <api_msg_flyteidl.admin.NotificationList>`) List of notifications based on Execution status transitions
  When this list is not empty it is used rather than any notifications defined in the referenced launch plan.
  When this list is empty, the notifications defined for the launch plan will be applied.
  
  
  
  Only one of :ref:`notifications <api_field_flyteidl.admin.ExecutionSpec.notifications>`, :ref:`disable_all <api_field_flyteidl.admin.ExecutionSpec.disable_all>` may be set.
  
.. _api_field_flyteidl.admin.ExecutionSpec.disable_all:

disable_all
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) This should be set to true if all notifications are intended to be disabled for this execution.
  
  
  
  Only one of :ref:`notifications <api_field_flyteidl.admin.ExecutionSpec.notifications>`, :ref:`disable_all <api_field_flyteidl.admin.ExecutionSpec.disable_all>` may be set.
  
.. _api_field_flyteidl.admin.ExecutionSpec.labels:

labels
  (:ref:`flyteidl.admin.Labels <api_msg_flyteidl.admin.Labels>`) Labels to apply to the execution resource.
  
  
.. _api_field_flyteidl.admin.ExecutionSpec.annotations:

annotations
  (:ref:`flyteidl.admin.Annotations <api_msg_flyteidl.admin.Annotations>`) Annotations to apply to the execution resource.
  
  
.. _api_field_flyteidl.admin.ExecutionSpec.auth_role:

auth_role
  (:ref:`flyteidl.admin.AuthRole <api_msg_flyteidl.admin.AuthRole>`) Optional: auth override to apply this execution.
  
  
.. _api_field_flyteidl.admin.ExecutionSpec.security_context:

security_context
  (:ref:`flyteidl.core.SecurityContext <api_msg_flyteidl.core.SecurityContext>`) Optional: security context override to apply this execution.
  
  
.. _api_field_flyteidl.admin.ExecutionSpec.quality_of_service:

quality_of_service
  (:ref:`flyteidl.core.QualityOfService <api_msg_flyteidl.core.QualityOfService>`) Indicates the runtime priority of the execution.
  
  


.. _api_msg_flyteidl.admin.ExecutionTerminateRequest:

flyteidl.admin.ExecutionTerminateRequest
----------------------------------------

`[flyteidl.admin.ExecutionTerminateRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/execution.proto#L252>`_

Request to terminate an in-progress execution.  This action is irreversible.
If an execution is already terminated, this request will simply be a no-op.
This request will fail if it references a non-existent execution.
If the request succeeds the phase "ABORTED" will be recorded for the termination
with the optional cause added to the output_result.

.. code-block:: json

  {
    "id": "{...}",
    "cause": "..."
  }

.. _api_field_flyteidl.admin.ExecutionTerminateRequest.id:

id
  (:ref:`flyteidl.core.WorkflowExecutionIdentifier <api_msg_flyteidl.core.WorkflowExecutionIdentifier>`) Uniquely identifies the individual workflow execution to be terminated.
  
  
.. _api_field_flyteidl.admin.ExecutionTerminateRequest.cause:

cause
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Optional reason for aborting.
  
  


.. _api_msg_flyteidl.admin.ExecutionTerminateResponse:

flyteidl.admin.ExecutionTerminateResponse
-----------------------------------------

`[flyteidl.admin.ExecutionTerminateResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/execution.proto#L260>`_


.. code-block:: json

  {}




.. _api_msg_flyteidl.admin.WorkflowExecutionGetDataRequest:

flyteidl.admin.WorkflowExecutionGetDataRequest
----------------------------------------------

`[flyteidl.admin.WorkflowExecutionGetDataRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/execution.proto#L265>`_

Request structure to fetch inputs and output urls for an execution.

.. code-block:: json

  {
    "id": "{...}"
  }

.. _api_field_flyteidl.admin.WorkflowExecutionGetDataRequest.id:

id
  (:ref:`flyteidl.core.WorkflowExecutionIdentifier <api_msg_flyteidl.core.WorkflowExecutionIdentifier>`) The identifier of the execution for which to fetch inputs and outputs.
  
  


.. _api_msg_flyteidl.admin.WorkflowExecutionGetDataResponse:

flyteidl.admin.WorkflowExecutionGetDataResponse
-----------------------------------------------

`[flyteidl.admin.WorkflowExecutionGetDataResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/execution.proto#L271>`_

Response structure for WorkflowExecutionGetDataRequest which contains inputs and outputs for an execution.

.. code-block:: json

  {
    "outputs": "{...}",
    "inputs": "{...}",
    "full_inputs": "{...}",
    "full_outputs": "{...}"
  }

.. _api_field_flyteidl.admin.WorkflowExecutionGetDataResponse.outputs:

outputs
  (:ref:`flyteidl.admin.UrlBlob <api_msg_flyteidl.admin.UrlBlob>`) Signed url to fetch a core.LiteralMap of execution outputs.
  
  
.. _api_field_flyteidl.admin.WorkflowExecutionGetDataResponse.inputs:

inputs
  (:ref:`flyteidl.admin.UrlBlob <api_msg_flyteidl.admin.UrlBlob>`) Signed url to fetch a core.LiteralMap of execution inputs.
  
  
.. _api_field_flyteidl.admin.WorkflowExecutionGetDataResponse.full_inputs:

full_inputs
  (:ref:`flyteidl.core.LiteralMap <api_msg_flyteidl.core.LiteralMap>`) Optional, full_inputs will only be populated if they are under a configured size threshold.
  
  
.. _api_field_flyteidl.admin.WorkflowExecutionGetDataResponse.full_outputs:

full_outputs
  (:ref:`flyteidl.core.LiteralMap <api_msg_flyteidl.core.LiteralMap>`) Optional, full_outputs will only be populated if they are under a configured size threshold.
  
  

