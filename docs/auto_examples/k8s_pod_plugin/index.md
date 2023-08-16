# Kubernetes Pods

```{eval-rst}
.. tags:: Integration, Kubernetes, Advanced
```

Flyte tasks, represented by the {py:func}`@task <flytekit:flytekit.task>` decorator, are essentially single functions that run in one container.
However, there may be situations where you need to run a job with more than one container or require additional capabilities, such as:

- Running a hyper-parameter optimizer that stores state in a Redis database
- Simulating a service locally
- Running a sidecar container for logging and monitoring purposes
- Running a pod with additional capabilities, such as mounting volumes

To support these use cases, Flyte provides a Pod configuration that allows you to customize the pod specification used to run the task.
This simplifies the process of implementing the Kubernetes pod abstraction for running multiple containers.

:::{note}
A Kubernetes pod will not exit if it contains any sidecar containers (containers that do not exit automatically).
You do not need to write any additional code to handle this, as Flyte automatically manages pod tasks.
:::

## Installation

To use the Flytekit pod plugin, run the following command:

```{eval-rst}
.. prompt:: bash

    pip install flytekitplugins-pod
```

```{auto-examples-toc}
pod
```
