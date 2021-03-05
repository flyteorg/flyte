.. _api_file_flyteidl/admin/launch_plan.proto:

launch_plan.proto
================================

.. _api_msg_flyteidl.admin.LaunchPlanCreateRequest:

flyteidl.admin.LaunchPlanCreateRequest
--------------------------------------

`[flyteidl.admin.LaunchPlanCreateRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/launch_plan.proto#L17>`_

Request to register a launch plan. A LaunchPlanSpec may include a complete or incomplete set of inputs required
to launch a workflow execution. By default all launch plans are registered in state INACTIVE. If you wish to
set the state to ACTIVE, you must submit a LaunchPlanUpdateRequest, after you have created a launch plan.

.. code-block:: json

  {
    "id": "{...}",
    "spec": "{...}"
  }

.. _api_field_flyteidl.admin.LaunchPlanCreateRequest.id:

id
  (:ref:`flyteidl.core.Identifier <api_msg_flyteidl.core.Identifier>`) Uniquely identifies a launch plan entity.
  
  
.. _api_field_flyteidl.admin.LaunchPlanCreateRequest.spec:

spec
  (:ref:`flyteidl.admin.LaunchPlanSpec <api_msg_flyteidl.admin.LaunchPlanSpec>`) User-provided launch plan details, including reference workflow, inputs and other metadata.
  
  


.. _api_msg_flyteidl.admin.LaunchPlanCreateResponse:

flyteidl.admin.LaunchPlanCreateResponse
---------------------------------------

`[flyteidl.admin.LaunchPlanCreateResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/launch_plan.proto#L25>`_


.. code-block:: json

  {}




.. _api_msg_flyteidl.admin.LaunchPlan:

flyteidl.admin.LaunchPlan
-------------------------

`[flyteidl.admin.LaunchPlan proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/launch_plan.proto#L43>`_

A LaunchPlan provides the capability to templatize workflow executions.
Launch plans simplify associating one or more schedules, inputs and notifications with your workflows.
Launch plans can be shared and used to trigger executions with predefined inputs even when a workflow
definition doesn't necessarily have a default value for said input.

.. code-block:: json

  {
    "id": "{...}",
    "spec": "{...}",
    "closure": "{...}"
  }

.. _api_field_flyteidl.admin.LaunchPlan.id:

id
  (:ref:`flyteidl.core.Identifier <api_msg_flyteidl.core.Identifier>`) 
  
.. _api_field_flyteidl.admin.LaunchPlan.spec:

spec
  (:ref:`flyteidl.admin.LaunchPlanSpec <api_msg_flyteidl.admin.LaunchPlanSpec>`) 
  
.. _api_field_flyteidl.admin.LaunchPlan.closure:

closure
  (:ref:`flyteidl.admin.LaunchPlanClosure <api_msg_flyteidl.admin.LaunchPlanClosure>`) 
  


.. _api_msg_flyteidl.admin.LaunchPlanList:

flyteidl.admin.LaunchPlanList
-----------------------------

`[flyteidl.admin.LaunchPlanList proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/launch_plan.proto#L51>`_

Response object for list launch plan requests.

.. code-block:: json

  {
    "launch_plans": [],
    "token": "..."
  }

.. _api_field_flyteidl.admin.LaunchPlanList.launch_plans:

launch_plans
  (:ref:`flyteidl.admin.LaunchPlan <api_msg_flyteidl.admin.LaunchPlan>`) 
  
.. _api_field_flyteidl.admin.LaunchPlanList.token:

token
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) In the case of multiple pages of results, the server-provided token can be used to fetch the next page
  in a query. If there are no more results, this value will be empty.
  
  


.. _api_msg_flyteidl.admin.Auth:

flyteidl.admin.Auth
-------------------

`[flyteidl.admin.Auth proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/launch_plan.proto#L61>`_

Defines permissions associated with executions created by this launch plan spec.
Deprecated.

.. code-block:: json

  {
    "assumable_iam_role": "...",
    "kubernetes_service_account": "..."
  }

.. _api_field_flyteidl.admin.Auth.assumable_iam_role:

assumable_iam_role
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
  
  Only one of :ref:`assumable_iam_role <api_field_flyteidl.admin.Auth.assumable_iam_role>`, :ref:`kubernetes_service_account <api_field_flyteidl.admin.Auth.kubernetes_service_account>` may be set.
  
.. _api_field_flyteidl.admin.Auth.kubernetes_service_account:

kubernetes_service_account
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
  
  Only one of :ref:`assumable_iam_role <api_field_flyteidl.admin.Auth.assumable_iam_role>`, :ref:`kubernetes_service_account <api_field_flyteidl.admin.Auth.kubernetes_service_account>` may be set.
  


.. _api_msg_flyteidl.admin.LaunchPlanSpec:

flyteidl.admin.LaunchPlanSpec
-----------------------------

`[flyteidl.admin.LaunchPlanSpec proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/launch_plan.proto#L69>`_

User-provided launch plan definition and configuration values.

.. code-block:: json

  {
    "workflow_id": "{...}",
    "entity_metadata": "{...}",
    "default_inputs": "{...}",
    "fixed_inputs": "{...}",
    "role": "...",
    "labels": "{...}",
    "annotations": "{...}",
    "auth": "{...}",
    "auth_role": "{...}",
    "security_context": "{...}",
    "quality_of_service": "{...}",
    "raw_output_data_config": "{...}"
  }

.. _api_field_flyteidl.admin.LaunchPlanSpec.workflow_id:

workflow_id
  (:ref:`flyteidl.core.Identifier <api_msg_flyteidl.core.Identifier>`) Reference to the Workflow template that the launch plan references
  
  
.. _api_field_flyteidl.admin.LaunchPlanSpec.entity_metadata:

entity_metadata
  (:ref:`flyteidl.admin.LaunchPlanMetadata <api_msg_flyteidl.admin.LaunchPlanMetadata>`) Metadata for the Launch Plan
  
  
.. _api_field_flyteidl.admin.LaunchPlanSpec.default_inputs:

default_inputs
  (:ref:`flyteidl.core.ParameterMap <api_msg_flyteidl.core.ParameterMap>`) Input values to be passed for the execution
  
  
.. _api_field_flyteidl.admin.LaunchPlanSpec.fixed_inputs:

fixed_inputs
  (:ref:`flyteidl.core.LiteralMap <api_msg_flyteidl.core.LiteralMap>`) Fixed, non-overridable inputs for the Launch Plan
  
  
.. _api_field_flyteidl.admin.LaunchPlanSpec.role:

role
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) String to indicate the role to use to execute the workflow underneath
  
  
.. _api_field_flyteidl.admin.LaunchPlanSpec.labels:

labels
  (:ref:`flyteidl.admin.Labels <api_msg_flyteidl.admin.Labels>`) Custom labels to be applied to the execution resource.
  
  
.. _api_field_flyteidl.admin.LaunchPlanSpec.annotations:

annotations
  (:ref:`flyteidl.admin.Annotations <api_msg_flyteidl.admin.Annotations>`) Custom annotations to be applied to the execution resource.
  
  
.. _api_field_flyteidl.admin.LaunchPlanSpec.auth:

auth
  (:ref:`flyteidl.admin.Auth <api_msg_flyteidl.admin.Auth>`) Indicates the permission associated with workflow executions triggered with this launch plan.
  
  
.. _api_field_flyteidl.admin.LaunchPlanSpec.auth_role:

auth_role
  (:ref:`flyteidl.admin.AuthRole <api_msg_flyteidl.admin.AuthRole>`) 
  
.. _api_field_flyteidl.admin.LaunchPlanSpec.security_context:

security_context
  (:ref:`flyteidl.core.SecurityContext <api_msg_flyteidl.core.SecurityContext>`) Indicates security context for permissions triggered with this launch plan
  
  
.. _api_field_flyteidl.admin.LaunchPlanSpec.quality_of_service:

quality_of_service
  (:ref:`flyteidl.core.QualityOfService <api_msg_flyteidl.core.QualityOfService>`) Indicates the runtime priority of the execution.
  
  
.. _api_field_flyteidl.admin.LaunchPlanSpec.raw_output_data_config:

raw_output_data_config
  (:ref:`flyteidl.admin.RawOutputDataConfig <api_msg_flyteidl.admin.RawOutputDataConfig>`) 
  


.. _api_msg_flyteidl.admin.LaunchPlanClosure:

flyteidl.admin.LaunchPlanClosure
--------------------------------

`[flyteidl.admin.LaunchPlanClosure proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/launch_plan.proto#L108>`_

Values computed by the flyte platform after launch plan registration.
These include expected_inputs required to be present in a CreateExecutionRequest
to launch the reference workflow as well timestamp values associated with the launch plan.

.. code-block:: json

  {
    "state": "...",
    "expected_inputs": "{...}",
    "expected_outputs": "{...}",
    "created_at": "{...}",
    "updated_at": "{...}"
  }

.. _api_field_flyteidl.admin.LaunchPlanClosure.state:

state
  (:ref:`flyteidl.admin.LaunchPlanState <api_enum_flyteidl.admin.LaunchPlanState>`) Indicate the Launch plan phase
  
  
.. _api_field_flyteidl.admin.LaunchPlanClosure.expected_inputs:

expected_inputs
  (:ref:`flyteidl.core.ParameterMap <api_msg_flyteidl.core.ParameterMap>`) Indicates the set of inputs to execute the Launch plan
  
  
.. _api_field_flyteidl.admin.LaunchPlanClosure.expected_outputs:

expected_outputs
  (:ref:`flyteidl.core.VariableMap <api_msg_flyteidl.core.VariableMap>`) Indicates the set of outputs from the Launch plan
  
  
.. _api_field_flyteidl.admin.LaunchPlanClosure.created_at:

created_at
  (:ref:`google.protobuf.Timestamp <api_msg_google.protobuf.Timestamp>`) Time at which the launch plan was created.
  
  
.. _api_field_flyteidl.admin.LaunchPlanClosure.updated_at:

updated_at
  (:ref:`google.protobuf.Timestamp <api_msg_google.protobuf.Timestamp>`) Time at which the launch plan was last updated.
  
  


.. _api_msg_flyteidl.admin.LaunchPlanMetadata:

flyteidl.admin.LaunchPlanMetadata
---------------------------------

`[flyteidl.admin.LaunchPlanMetadata proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/launch_plan.proto#L127>`_

Additional launch plan attributes included in the LaunchPlanSpec not strictly required to launch
the reference workflow.

.. code-block:: json

  {
    "schedule": "{...}",
    "notifications": []
  }

.. _api_field_flyteidl.admin.LaunchPlanMetadata.schedule:

schedule
  (:ref:`flyteidl.admin.Schedule <api_msg_flyteidl.admin.Schedule>`) Schedule to execute the Launch Plan
  
  
.. _api_field_flyteidl.admin.LaunchPlanMetadata.notifications:

notifications
  (:ref:`flyteidl.admin.Notification <api_msg_flyteidl.admin.Notification>`) List of notifications based on Execution status transitions
  
  


.. _api_msg_flyteidl.admin.LaunchPlanUpdateRequest:

flyteidl.admin.LaunchPlanUpdateRequest
--------------------------------------

`[flyteidl.admin.LaunchPlanUpdateRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/launch_plan.proto#L136>`_

Request to set the referenced launch plan state to the configured value.

.. code-block:: json

  {
    "id": "{...}",
    "state": "..."
  }

.. _api_field_flyteidl.admin.LaunchPlanUpdateRequest.id:

id
  (:ref:`flyteidl.core.Identifier <api_msg_flyteidl.core.Identifier>`) Identifier of launch plan for which to change state.
  
  
.. _api_field_flyteidl.admin.LaunchPlanUpdateRequest.state:

state
  (:ref:`flyteidl.admin.LaunchPlanState <api_enum_flyteidl.admin.LaunchPlanState>`) Desired state to apply to the launch plan.
  
  


.. _api_msg_flyteidl.admin.LaunchPlanUpdateResponse:

flyteidl.admin.LaunchPlanUpdateResponse
---------------------------------------

`[flyteidl.admin.LaunchPlanUpdateResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/launch_plan.proto#L145>`_

Purposefully empty, may be populated in the future.

.. code-block:: json

  {}




.. _api_msg_flyteidl.admin.ActiveLaunchPlanRequest:

flyteidl.admin.ActiveLaunchPlanRequest
--------------------------------------

`[flyteidl.admin.ActiveLaunchPlanRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/launch_plan.proto#L149>`_

Represents a request struct for finding an active launch plan for a given NamedEntityIdentifier

.. code-block:: json

  {
    "id": "{...}"
  }

.. _api_field_flyteidl.admin.ActiveLaunchPlanRequest.id:

id
  (:ref:`flyteidl.admin.NamedEntityIdentifier <api_msg_flyteidl.admin.NamedEntityIdentifier>`) 
  


.. _api_msg_flyteidl.admin.ActiveLaunchPlanListRequest:

flyteidl.admin.ActiveLaunchPlanListRequest
------------------------------------------

`[flyteidl.admin.ActiveLaunchPlanListRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/launch_plan.proto#L154>`_

Represents a request structure to list active launch plans within a project/domain.

.. code-block:: json

  {
    "project": "...",
    "domain": "...",
    "limit": "...",
    "token": "...",
    "sort_by": "{...}"
  }

.. _api_field_flyteidl.admin.ActiveLaunchPlanListRequest.project:

project
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Name of the project that contains the identifiers.
  
  
.. _api_field_flyteidl.admin.ActiveLaunchPlanListRequest.domain:

domain
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Name of the domain the identifiers belongs to within the project.
  
  
.. _api_field_flyteidl.admin.ActiveLaunchPlanListRequest.limit:

limit
  (`uint32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Indicates the number of resources to be returned.
  
  
.. _api_field_flyteidl.admin.ActiveLaunchPlanListRequest.token:

token
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) In the case of multiple pages of results, the server-provided token can be used to fetch the next page
  in a query.
  +optional
  
  
.. _api_field_flyteidl.admin.ActiveLaunchPlanListRequest.sort_by:

sort_by
  (:ref:`flyteidl.admin.Sort <api_msg_flyteidl.admin.Sort>`) Sort ordering.
  +optional
  
  

.. _api_enum_flyteidl.admin.LaunchPlanState:

Enum flyteidl.admin.LaunchPlanState
-----------------------------------

`[flyteidl.admin.LaunchPlanState proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/launch_plan.proto#L34>`_

By default any launch plan regardless of state can be used to launch a workflow execution.
However, at most one version of a launch plan
(e.g. a NamedEntityIdentifier set of shared project, domain and name values) can be
active at a time in regards to *schedules*. That is, at most one schedule in a NamedEntityIdentifier
group will be observed and trigger executions at a defined cadence.

.. _api_enum_value_flyteidl.admin.LaunchPlanState.INACTIVE:

INACTIVE
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.admin.LaunchPlanState.ACTIVE:

ACTIVE
  ⁣
  
