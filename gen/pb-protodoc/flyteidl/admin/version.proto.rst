.. _api_file_flyteidl/admin/version.proto:

version.proto
============================

.. _api_msg_flyteidl.admin.Version:

flyteidl.admin.Version
----------------------

`[flyteidl.admin.Version proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/version.proto#L6>`_

Represents a request structure to get version of flyteadmin.

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

`[flyteidl.admin.GetVersionRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/version.proto#L18>`_

Represents a get version request.

.. code-block:: json

  {}



