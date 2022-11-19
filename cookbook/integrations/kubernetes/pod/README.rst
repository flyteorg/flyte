Kubernetes Pods
===============

.. tags:: Integration, Kubernetes, Advanced

Flyte tasks (Python functions decorated with :py:func:`@task <flytekit:flytekit.task>`) are essentially single functions loaded in one container.
But often, there is a need to run a job with more than one container, in cases such as:

- A special hyper-parameter optimizer with state stored in a Redis DB
- Simulating a service locally
- Running a sidecar to fetch or orchestrate data to handle additional logging and monitoring
- Running a pod with its complete set of possibilities: mounting volumes, etc.

Flyte provides a simplified interface to implement Kubernetes pod abstraction to execute multiple containers.
This is done using ``Pod`` task_config, which enables customization of the pod spec used to run the task.

.. note::

    Typically, a Kubernetes pod will not exit if the pod contains any sidecars (containers that do not exit automatically).
    No additional code is required to handle this as Flyte automatically takes care of pod tasks.

Installation
------------

To use the Flytekit pod plugin, run the following command:

.. prompt:: bash

   pip install flytekitplugins-pod
