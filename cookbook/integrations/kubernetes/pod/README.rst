
Kubernetes Pods
================

By default Flyte tasks - ``@task`` are essentially single functions that are loaded in one container. But, often times there is a need to run a job, with more than one containers. For example,

  - a special Hyper-parameter optimizer with state stored in a redis db
  - simulating a service locally
  - run a sidecar to fetch or orchestrate data, handle additional logging or monitoring
  - run a pod with it's complete set of possibilities- mounting volumes etc

In this case, a regular task is not enough. For such a case, users can simply use a kubernetes pod abstraction provided by Flyte to execute multiple containers.
This can be done by using ``Pod`` task_config can be leveraged to fully customize the pod spec used to run your task.

.. note::

    Flyte will automatically take care of pod tasks. Typically a K8s pod will not exit, if the pod contains any sidecars (containers that do not exit automatically). No additional code is required.

Installation
------------

To use the flytekit pod plugin simply run the following:

.. prompt:: bash

   pip install flytekitplugins-pod