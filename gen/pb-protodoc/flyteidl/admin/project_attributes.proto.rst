.. _api_file_flyteidl/admin/project_attributes.proto:

project_attributes.proto
=======================================

.. _api_msg_flyteidl.admin.ProjectAttributes:

flyteidl.admin.ProjectAttributes
--------------------------------

`[flyteidl.admin.ProjectAttributes proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project_attributes.proto#L8>`_


.. code-block:: json

  {
    "project": "...",
    "matching_attributes": "{...}"
  }

.. _api_field_flyteidl.admin.ProjectAttributes.project:

project
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Unique project id for which this set of attributes will be applied.
  
  
.. _api_field_flyteidl.admin.ProjectAttributes.matching_attributes:

matching_attributes
  (:ref:`flyteidl.admin.MatchingAttributes <api_msg_flyteidl.admin.MatchingAttributes>`) 
  


.. _api_msg_flyteidl.admin.ProjectAttributesUpdateRequest:

flyteidl.admin.ProjectAttributesUpdateRequest
---------------------------------------------

`[flyteidl.admin.ProjectAttributesUpdateRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project_attributes.proto#L16>`_

Sets custom attributes for a project combination.

.. code-block:: json

  {
    "attributes": "{...}"
  }

.. _api_field_flyteidl.admin.ProjectAttributesUpdateRequest.attributes:

attributes
  (:ref:`flyteidl.admin.ProjectAttributes <api_msg_flyteidl.admin.ProjectAttributes>`) 
  


.. _api_msg_flyteidl.admin.ProjectAttributesUpdateResponse:

flyteidl.admin.ProjectAttributesUpdateResponse
----------------------------------------------

`[flyteidl.admin.ProjectAttributesUpdateResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project_attributes.proto#L21>`_

Purposefully empty, may be populated in the future.

.. code-block:: json

  {}



