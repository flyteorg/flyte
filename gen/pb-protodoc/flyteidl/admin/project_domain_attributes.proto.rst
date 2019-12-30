.. _api_file_flyteidl/admin/project_domain_attributes.proto:

project_domain_attributes.proto
==============================================

.. _api_msg_flyteidl.admin.ProjectDomainAttributes:

flyteidl.admin.ProjectDomainAttributes
--------------------------------------

`[flyteidl.admin.ProjectDomainAttributes proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project_domain_attributes.proto#L7>`_


.. code-block:: json

  {
    "project": "...",
    "domain": "...",
    "matching_attributes": "{...}"
  }

.. _api_field_flyteidl.admin.ProjectDomainAttributes.project:

project
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Unique project id for which this set of attributes will be applied.
  
  
.. _api_field_flyteidl.admin.ProjectDomainAttributes.domain:

domain
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Unique domain id for which this set of attributes will be applied.
  
  
.. _api_field_flyteidl.admin.ProjectDomainAttributes.matching_attributes:

matching_attributes
  (:ref:`flyteidl.admin.MatchingAttributes <api_msg_flyteidl.admin.MatchingAttributes>`) 
  


.. _api_msg_flyteidl.admin.ProjectDomainAttributesUpdateRequest:

flyteidl.admin.ProjectDomainAttributesUpdateRequest
---------------------------------------------------

`[flyteidl.admin.ProjectDomainAttributesUpdateRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project_domain_attributes.proto#L18>`_

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

`[flyteidl.admin.ProjectDomainAttributesUpdateResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project_domain_attributes.proto#L23>`_

Purposefully empty, may be populated in the future.

.. code-block:: json

  {}



