.. _api_file_flyteidl/plugins/sidecar.proto:

sidecar.proto
==============================

.. _api_msg_flyteidl.plugins.SidecarJob:

flyteidl.plugins.SidecarJob
---------------------------

`[flyteidl.plugins.SidecarJob proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/sidecar.proto#L11>`_

A sidecar job brings up the desired pod_spec.
The plugin executor is responsible for keeping the pod alive until the primary container terminates
or the task itself times out.

.. code-block:: json

  {
    "pod_spec": "{...}",
    "primary_container_name": "..."
  }

.. _api_field_flyteidl.plugins.SidecarJob.pod_spec:

pod_spec
  (:ref:`k8s.io.api.core.v1.PodSpec <api_msg_k8s.io.api.core.v1.PodSpec>`) 
  
.. _api_field_flyteidl.plugins.SidecarJob.primary_container_name:

primary_container_name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  

