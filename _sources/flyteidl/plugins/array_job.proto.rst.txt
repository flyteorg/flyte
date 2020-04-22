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
    "min_successes": "..."
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
  +required.
  
  
.. _api_field_flyteidl.plugins.ArrayJob.min_successes:

min_successes
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) An absolute number of the minimum number of successful completions of subtasks. As soon as this criteria is met,
  the array job will be marked as successful and outputs will be computed. This has to be a non-negative number if
  assigned. Default value is size.
  
  

