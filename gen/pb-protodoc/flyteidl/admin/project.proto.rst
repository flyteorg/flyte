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

`[flyteidl.admin.Project proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project.proto#L19>`_

Top-level namespace used to classify different entities like workflows and executions.

.. code-block:: json

  {
    "id": "...",
    "name": "...",
    "domains": [],
    "description": "...",
    "labels": "{...}",
    "state": "..."
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
  
  
.. _api_field_flyteidl.admin.Project.state:

state
  (:ref:`flyteidl.admin.Project.ProjectState <api_enum_flyteidl.admin.Project.ProjectState>`) 
  

.. _api_enum_flyteidl.admin.Project.ProjectState:

Enum flyteidl.admin.Project.ProjectState
----------------------------------------

`[flyteidl.admin.Project.ProjectState proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project.proto#L21>`_

The state of the project is used to control its visibility in the UI and validity.

.. _api_enum_value_flyteidl.admin.Project.ProjectState.ACTIVE:

ACTIVE
  *(DEFAULT)* ⁣By default, all projects are considered active.
  
  
.. _api_enum_value_flyteidl.admin.Project.ProjectState.ARCHIVED:

ARCHIVED
  ⁣Archived projects are no longer visible in the UI and no longer valid.
  
  
.. _api_enum_value_flyteidl.admin.Project.ProjectState.SYSTEM_GENERATED:

SYSTEM_GENERATED
  ⁣System generated projects that aren't explicitly created or managed by a user.
  
  

.. _api_msg_flyteidl.admin.Projects:

flyteidl.admin.Projects
-----------------------

`[flyteidl.admin.Projects proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project.proto#L48>`_


.. code-block:: json

  {
    "projects": [],
    "token": "..."
  }

.. _api_field_flyteidl.admin.Projects.projects:

projects
  (:ref:`flyteidl.admin.Project <api_msg_flyteidl.admin.Project>`) 
  
.. _api_field_flyteidl.admin.Projects.token:

token
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) In the case of multiple pages of results, the server-provided token can be used to fetch the next page
  in a query. If there are no more results, this value will be empty.
  
  


.. _api_msg_flyteidl.admin.ProjectListRequest:

flyteidl.admin.ProjectListRequest
---------------------------------

`[flyteidl.admin.ProjectListRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project.proto#L56>`_


.. code-block:: json

  {
    "limit": "...",
    "token": "...",
    "filters": "...",
    "sort_by": "{...}"
  }

.. _api_field_flyteidl.admin.ProjectListRequest.limit:

limit
  (`uint32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Indicates the number of projects to be returned.
  
  
.. _api_field_flyteidl.admin.ProjectListRequest.token:

token
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) In the case of multiple pages of results, this server-provided token can be used to fetch the next page
  in a query.
  +optional
  
  
.. _api_field_flyteidl.admin.ProjectListRequest.filters:

filters
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Indicates a list of filters passed as string.
  More info on constructing filters : <Link>
  +optional
  
  
.. _api_field_flyteidl.admin.ProjectListRequest.sort_by:

sort_by
  (:ref:`flyteidl.admin.Sort <api_msg_flyteidl.admin.Sort>`) Sort ordering.
  +optional
  
  


.. _api_msg_flyteidl.admin.ProjectRegisterRequest:

flyteidl.admin.ProjectRegisterRequest
-------------------------------------

`[flyteidl.admin.ProjectRegisterRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project.proto#L73>`_


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

`[flyteidl.admin.ProjectRegisterResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project.proto#L77>`_


.. code-block:: json

  {}




.. _api_msg_flyteidl.admin.ProjectUpdateResponse:

flyteidl.admin.ProjectUpdateResponse
------------------------------------

`[flyteidl.admin.ProjectUpdateResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/project.proto#L80>`_


.. code-block:: json

  {}



