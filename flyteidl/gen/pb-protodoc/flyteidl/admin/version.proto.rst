.. _api_file_flyteidl/admin/version.proto:

version.proto
============================

.. _api_msg_flyteidl.admin.GetVersionResponse:

flyteidl.admin.GetVersionResponse
---------------------------------

`[flyteidl.admin.GetVersionResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/version.proto#L6>`_

Response for the GetVersion API

.. code-block:: json

  {
    "control_plane_version": "{...}"
  }

.. _api_field_flyteidl.admin.GetVersionResponse.control_plane_version:

control_plane_version
  (:ref:`flyteidl.admin.Version <api_msg_flyteidl.admin.Version>`) The control plane version information. FlyteAdmin and related components
  form the control plane of Flyte
  
  


.. _api_msg_flyteidl.admin.Version:

flyteidl.admin.Version
----------------------

`[flyteidl.admin.Version proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/version.proto#L13>`_

Provides Version information for a component

.. code-block:: json

  {
    "Build": "...",
    "Version": "...",
    "BuildTime": "..."
  }

.. _api_field_flyteidl.admin.Version.Build:

Build
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Specifies the GIT sha of the build
  
  
.. _api_field_flyteidl.admin.Version.Version:

Version
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Version for the build, should follow a semver
  
  
.. _api_field_flyteidl.admin.Version.BuildTime:

BuildTime
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Build timestamp
  
  


.. _api_msg_flyteidl.admin.GetVersionRequest:

flyteidl.admin.GetVersionRequest
--------------------------------

`[flyteidl.admin.GetVersionRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/version.proto#L25>`_

Empty request for GetVersion

.. code-block:: json

  {}



