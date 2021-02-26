.. _api_file_flyteidl/plugins/array_job.proto:

array_job.proto
================================

.. _api_msg_flyteidl.plugins.ArrayJob:

flyteidl.plugins.ArrayJob
-------------------------

`[flyteidl.plugins.ArrayJob proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/array_job.proto#L8>`_

Describes a job that can process independent pieces of data concurrently. Multiple copies of the runnable component
will be executed concurrently.

.. code-block:: json

  {
    "parallelism": "...",
    "size": "...",
    "min_successes": "...",
    "min_success_ratio": "..."
  }

.. _api_field_flyteidl.plugins.ArrayJob.parallelism:

parallelism
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Defines the minimum number of instances to bring up concurrently at any given point. Note that this is an
  optimistic restriction and that, due to network partitioning or other failures, the actual number of currently
  running instances might be more. This has to be a positive number if assigned. Default value is size.
  
  
.. _api_field_flyteidl.plugins.ArrayJob.size:

size
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Defines the number of instances to launch at most. This number should match the size of the input if the job
  requires processing of all input data. This has to be a positive number.
  In the case this is not defined, the back-end will determine the size at run-time by reading the inputs.
  
  
.. _api_field_flyteidl.plugins.ArrayJob.min_successes:

min_successes
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) An absolute number of the minimum number of successful completions of subtasks. As soon as this criteria is met,
  the array job will be marked as successful and outputs will be computed. This has to be a non-negative number if
  assigned. Default value is size (if specified).
  
  
  
  Only one of :ref:`min_successes <api_field_flyteidl.plugins.ArrayJob.min_successes>`, :ref:`min_success_ratio <api_field_flyteidl.plugins.ArrayJob.min_success_ratio>` may be set.
  
.. _api_field_flyteidl.plugins.ArrayJob.min_success_ratio:

min_success_ratio
  (`float <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) If the array job size is not known beforehand, the min_success_ratio can instead be used to determine when an array
  job can be marked successful.
  
  
  
  Only one of :ref:`min_successes <api_field_flyteidl.plugins.ArrayJob.min_successes>`, :ref:`min_success_ratio <api_field_flyteidl.plugins.ArrayJob.min_success_ratio>` may be set.
  

