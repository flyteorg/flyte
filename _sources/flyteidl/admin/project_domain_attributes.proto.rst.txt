.. _api_file_flyteidl/admin/project_domain_attributes.proto:

project_domain_attributes.proto
==============================================

.. _api_msg_flyteidl.admin.ProjectDomainAttributes:

flyteidl.admin.ProjectDomainAttributes
--------------------------------------

`[flyteidl.admin.ProjectDomainAttributes proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project_domain_attributes.proto#L5>`_


.. code-block:: json

  {
    "project": "...",
    "domain": "...",
    "attributes": "{...}"
  }

.. _api_field_flyteidl.admin.ProjectDomainAttributes.project:

project
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Unique project id for which this set of attributes will be applied.
  
  
.. _api_field_flyteidl.admin.ProjectDomainAttributes.domain:

domain
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Unique domain id for which this set of attributes will be applied.
  
  
.. _api_field_flyteidl.admin.ProjectDomainAttributes.attributes:

attributes
  (map<`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_, `string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_>) Custom resource attributes which will be applied in cluster resource creation (e.g. quotas).
  Map keys are the *case-sensitive* names of variables in templatized resource files.
  Map values should be the custom values which get substituted during resource creation.
  
  


.. _api_msg_flyteidl.admin.ProjectDomainAttributesUpdateRequest:

flyteidl.admin.ProjectDomainAttributesUpdateRequest
---------------------------------------------------

`[flyteidl.admin.ProjectDomainAttributesUpdateRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project_domain_attributes.proto#L19>`_

Sets custom attributes for a project-domain combination.

.. code-block:: json

  {
    "attributes": "{...}"
  }

.. _api_field_flyteidl.admin.ProjectDomainAttributesUpdateRequest.attributes:

attributes
  (:ref:`flyteidl.admin.ProjectDomainAttributes <api_msg_flyteidl.admin.ProjectDomainAttributes>`) 
  


.. _api_msg_flyteidl.admin.ProjectDomainAttributesUpdateResponse:

flyteidl.admin.ProjectDomainAttributesUpdateResponse
----------------------------------------------------

`[flyteidl.admin.ProjectDomainAttributesUpdateResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project_domain_attributes.proto#L24>`_

Purposefully empty, may be populated in the future.

.. code-block:: json

  {}



