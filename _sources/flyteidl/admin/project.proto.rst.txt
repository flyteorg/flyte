.. _api_file_flyteidl/admin/project.proto:

project.proto
============================

.. _api_msg_flyteidl.admin.Domain:

flyteidl.admin.Domain
---------------------

`[flyteidl.admin.Domain proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project.proto#L10>`_

Namespace within a project commonly used to differentiate between different service instances.
e.g. "production", "development", etc.

.. code-block:: json

  {
    "id": "...",
    "name": "..."
  }

.. _api_field_flyteidl.admin.Domain.id:

id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.admin.Domain.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Display name.
  
  


.. _api_msg_flyteidl.admin.Project:

flyteidl.admin.Project
----------------------

`[flyteidl.admin.Project proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project.proto#L18>`_

Top-level namespace used to classify different entities like workflows and executions.

.. code-block:: json

  {
    "id": "...",
    "name": "...",
    "domains": [],
    "description": "...",
    "labels": "{...}"
  }

.. _api_field_flyteidl.admin.Project.id:

id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.admin.Project.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Display name.
  
  
.. _api_field_flyteidl.admin.Project.domains:

domains
  (:ref:`flyteidl.admin.Domain <api_msg_flyteidl.admin.Domain>`) 
  
.. _api_field_flyteidl.admin.Project.description:

description
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.admin.Project.labels:

labels
  (:ref:`flyteidl.admin.Labels <api_msg_flyteidl.admin.Labels>`) Leverage Labels from flyteidel.admin.common.proto to
  tag projects with ownership information.
  
  


.. _api_msg_flyteidl.admin.Projects:

flyteidl.admin.Projects
-----------------------

`[flyteidl.admin.Projects proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project.proto#L33>`_


.. code-block:: json

  {
    "projects": []
  }

.. _api_field_flyteidl.admin.Projects.projects:

projects
  (:ref:`flyteidl.admin.Project <api_msg_flyteidl.admin.Project>`) 
  


.. _api_msg_flyteidl.admin.ProjectListRequest:

flyteidl.admin.ProjectListRequest
---------------------------------

`[flyteidl.admin.ProjectListRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project.proto#L37>`_


.. code-block:: json

  {}




.. _api_msg_flyteidl.admin.ProjectRegisterRequest:

flyteidl.admin.ProjectRegisterRequest
-------------------------------------

`[flyteidl.admin.ProjectRegisterRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project.proto#L40>`_


.. code-block:: json

  {
    "project": "{...}"
  }

.. _api_field_flyteidl.admin.ProjectRegisterRequest.project:

project
  (:ref:`flyteidl.admin.Project <api_msg_flyteidl.admin.Project>`) 
  


.. _api_msg_flyteidl.admin.ProjectRegisterResponse:

flyteidl.admin.ProjectRegisterResponse
--------------------------------------

`[flyteidl.admin.ProjectRegisterResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project.proto#L44>`_


.. code-block:: json

  {}




.. _api_msg_flyteidl.admin.ProjectUpdateResponse:

flyteidl.admin.ProjectUpdateResponse
------------------------------------

`[flyteidl.admin.ProjectUpdateResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project.proto#L47>`_


.. code-block:: json

  {}



