.. _api_file_flyteidl/admin/workflow.proto:

workflow.proto
=============================

.. _api_msg_flyteidl.admin.WorkflowCreateRequest:

flyteidl.admin.WorkflowCreateRequest
------------------------------------

`[flyteidl.admin.WorkflowCreateRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/workflow.proto#L12>`_

Represents a request structure to create a revision of a workflow.

.. code-block:: json

  {
    "id": "{...}",
    "spec": "{...}"
  }

.. _api_field_flyteidl.admin.WorkflowCreateRequest.id:

id
  (:ref:`flyteidl.core.Identifier <api_msg_flyteidl.core.Identifier>`) id represents the unique identifier of the workflow.
  
  
.. _api_field_flyteidl.admin.WorkflowCreateRequest.spec:

spec
  (:ref:`flyteidl.admin.WorkflowSpec <api_msg_flyteidl.admin.WorkflowSpec>`) Represents the specification for workflow.
  
  


.. _api_msg_flyteidl.admin.WorkflowCreateResponse:

flyteidl.admin.WorkflowCreateResponse
-------------------------------------

`[flyteidl.admin.WorkflowCreateResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/workflow.proto#L20>`_


.. code-block:: json

  {}




.. _api_msg_flyteidl.admin.Workflow:

flyteidl.admin.Workflow
-----------------------

`[flyteidl.admin.Workflow proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/workflow.proto#L27>`_

Represents the workflow structure stored in the Admin
A workflow is created by ordering tasks and associating outputs to inputs
in order to produce a directed-acyclic execution graph.

.. code-block:: json

  {
    "id": "{...}",
    "closure": "{...}"
  }

.. _api_field_flyteidl.admin.Workflow.id:

id
  (:ref:`flyteidl.core.Identifier <api_msg_flyteidl.core.Identifier>`) id represents the unique identifier of the workflow.
  
  
.. _api_field_flyteidl.admin.Workflow.closure:

closure
  (:ref:`flyteidl.admin.WorkflowClosure <api_msg_flyteidl.admin.WorkflowClosure>`) closure encapsulates all the fields that maps to a compiled version of the workflow.
  
  


.. _api_msg_flyteidl.admin.WorkflowList:

flyteidl.admin.WorkflowList
---------------------------

`[flyteidl.admin.WorkflowList proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/workflow.proto#L36>`_

Represents a list of workflows returned from the admin.

.. code-block:: json

  {
    "workflows": [],
    "token": "..."
  }

.. _api_field_flyteidl.admin.WorkflowList.workflows:

workflows
  (:ref:`flyteidl.admin.Workflow <api_msg_flyteidl.admin.Workflow>`) A list of workflows returned based on the request.
  
  
.. _api_field_flyteidl.admin.WorkflowList.token:

token
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) In the case of multiple pages of results, the server-provided token can be used to fetch the next page
  in a query. If there are no more results, this value will be empty.
  
  


.. _api_msg_flyteidl.admin.WorkflowSpec:

flyteidl.admin.WorkflowSpec
---------------------------

`[flyteidl.admin.WorkflowSpec proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/workflow.proto#L46>`_

Represents a structure that encapsulates the specification of the workflow.

.. code-block:: json

  {
    "template": "{...}",
    "sub_workflows": []
  }

.. _api_field_flyteidl.admin.WorkflowSpec.template:

template
  (:ref:`flyteidl.core.WorkflowTemplate <api_msg_flyteidl.core.WorkflowTemplate>`) Template of the task that encapsulates all the metadata of the workflow.
  
  
.. _api_field_flyteidl.admin.WorkflowSpec.sub_workflows:

sub_workflows
  (:ref:`flyteidl.core.WorkflowTemplate <api_msg_flyteidl.core.WorkflowTemplate>`) Workflows that are embedded into other workflows need to be passed alongside the parent workflow to the
  propeller compiler (since the compiler doesn't have any knowledge of other workflows - ie, it doesn't reach out
  to Admin to see other registered workflows).  In fact, subworkflows do not even need to be registered.
  
  


.. _api_msg_flyteidl.admin.WorkflowClosure:

flyteidl.admin.WorkflowClosure
------------------------------

`[flyteidl.admin.WorkflowClosure proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/workflow.proto#L57>`_

A container holding the compiled workflow produced from the WorkflowSpec and additional metadata.

.. code-block:: json

  {
    "compiled_workflow": "{...}",
    "created_at": "{...}"
  }

.. _api_field_flyteidl.admin.WorkflowClosure.compiled_workflow:

compiled_workflow
  (:ref:`flyteidl.core.CompiledWorkflowClosure <api_msg_flyteidl.core.CompiledWorkflowClosure>`) Represents the compiled representation of the workflow from the specification provided.
  
  
.. _api_field_flyteidl.admin.WorkflowClosure.created_at:

created_at
  (:ref:`google.protobuf.Timestamp <api_msg_google.protobuf.Timestamp>`) Time at which the workflow was created.
  
  

