.. _api_file_flyteidl/admin/workflow_attributes.proto:

workflow_attributes.proto
========================================

.. _api_msg_flyteidl.admin.WorkflowAttributes:

flyteidl.admin.WorkflowAttributes
---------------------------------

`[flyteidl.admin.WorkflowAttributes proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/workflow_attributes.proto#L7>`_


.. code-block:: json

  {
    "project": "...",
    "domain": "...",
    "workflow": "...",
    "matching_attributes": "{...}"
  }

.. _api_field_flyteidl.admin.WorkflowAttributes.project:

project
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Unique project id for which this set of attributes will be applied.
  
  
.. _api_field_flyteidl.admin.WorkflowAttributes.domain:

domain
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Unique domain id for which this set of attributes will be applied.
  
  
.. _api_field_flyteidl.admin.WorkflowAttributes.workflow:

workflow
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Workflow name for which this set of attributes will be applied.
  
  
.. _api_field_flyteidl.admin.WorkflowAttributes.matching_attributes:

matching_attributes
  (:ref:`flyteidl.admin.MatchingAttributes <api_msg_flyteidl.admin.MatchingAttributes>`) 
  


.. _api_msg_flyteidl.admin.WorkflowAttributesUpdateRequest:

flyteidl.admin.WorkflowAttributesUpdateRequest
----------------------------------------------

`[flyteidl.admin.WorkflowAttributesUpdateRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/workflow_attributes.proto#L21>`_

Sets custom attributes for a project, domain and workflow combination.

.. code-block:: json

  {
    "attributes": "{...}"
  }

.. _api_field_flyteidl.admin.WorkflowAttributesUpdateRequest.attributes:

attributes
  (:ref:`flyteidl.admin.WorkflowAttributes <api_msg_flyteidl.admin.WorkflowAttributes>`) 
  


.. _api_msg_flyteidl.admin.WorkflowAttributesUpdateResponse:

flyteidl.admin.WorkflowAttributesUpdateResponse
-----------------------------------------------

`[flyteidl.admin.WorkflowAttributesUpdateResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/workflow_attributes.proto#L26>`_

Purposefully empty, may be populated in the future.

.. code-block:: json

  {}



