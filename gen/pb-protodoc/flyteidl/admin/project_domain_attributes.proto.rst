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




.. _api_msg_flyteidl.admin.ProjectDomainAttributesGetRequest:

flyteidl.admin.ProjectDomainAttributesGetRequest
------------------------------------------------

`[flyteidl.admin.ProjectDomainAttributesGetRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project_domain_attributes.proto#L26>`_


.. code-block:: json

  {
    "project": "...",
    "domain": "...",
    "resource_type": "..."
  }

.. _api_field_flyteidl.admin.ProjectDomainAttributesGetRequest.project:

project
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Unique project id which this set of attributes references.
  
  
.. _api_field_flyteidl.admin.ProjectDomainAttributesGetRequest.domain:

domain
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Unique domain id which this set of attributes references.
  
  
.. _api_field_flyteidl.admin.ProjectDomainAttributesGetRequest.resource_type:

resource_type
  (:ref:`flyteidl.admin.MatchableResource <api_enum_flyteidl.admin.MatchableResource>`) 
  


.. _api_msg_flyteidl.admin.ProjectDomainAttributesGetResponse:

flyteidl.admin.ProjectDomainAttributesGetResponse
-------------------------------------------------

`[flyteidl.admin.ProjectDomainAttributesGetResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project_domain_attributes.proto#L36>`_


.. code-block:: json

  {
    "attributes": "{...}"
  }

.. _api_field_flyteidl.admin.ProjectDomainAttributesGetResponse.attributes:

attributes
  (:ref:`flyteidl.admin.ProjectDomainAttributes <api_msg_flyteidl.admin.ProjectDomainAttributes>`) 
  


.. _api_msg_flyteidl.admin.ProjectDomainAttributesDeleteRequest:

flyteidl.admin.ProjectDomainAttributesDeleteRequest
---------------------------------------------------

`[flyteidl.admin.ProjectDomainAttributesDeleteRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project_domain_attributes.proto#L41>`_


.. code-block:: json

  {
    "project": "...",
    "domain": "...",
    "resource_type": "..."
  }

.. _api_field_flyteidl.admin.ProjectDomainAttributesDeleteRequest.project:

project
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Unique project id which this set of attributes references.
  
  
.. _api_field_flyteidl.admin.ProjectDomainAttributesDeleteRequest.domain:

domain
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Unique domain id which this set of attributes references.
  
  
.. _api_field_flyteidl.admin.ProjectDomainAttributesDeleteRequest.resource_type:

resource_type
  (:ref:`flyteidl.admin.MatchableResource <api_enum_flyteidl.admin.MatchableResource>`) 
  


.. _api_msg_flyteidl.admin.ProjectDomainAttributesDeleteResponse:

flyteidl.admin.ProjectDomainAttributesDeleteResponse
----------------------------------------------------

`[flyteidl.admin.ProjectDomainAttributesDeleteResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project_domain_attributes.proto#L52>`_

Purposefully empty, may be populated in the future.

.. code-block:: json

  {}



