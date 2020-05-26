.. _api_file_flyteidl/plugins/pytorch.proto:

pytorch.proto
==============================

.. _api_msg_flyteidl.plugins.DistributedPyTorchTrainingTask:

flyteidl.plugins.DistributedPyTorchTrainingTask
-----------------------------------------------

`[flyteidl.plugins.DistributedPyTorchTrainingTask proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/pytorch.proto#L7>`_

Custom proto for plugin that enables distributed training using https://github.com/kubeflow/pytorch-operator

.. code-block:: json

  {
    "workers": "..."
  }

.. _api_field_flyteidl.plugins.DistributedPyTorchTrainingTask.workers:

workers
  (`int32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) number of worker replicas spawned in the cluster for this job
  
  

