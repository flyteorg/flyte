---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.14.7
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

# Ray Tasks

The Ray task offers the capability to execute a Ray job either on a pre-existing Ray cluster or by creating a new Ray cluster using the Ray operator.

Let's begin with an example to demonstrate the usage of the Ray task.

:::{Warning}
**Known Bugs and Compatibility Issue with Kuberay Operator Versions**
Please note that there have been reports of various bugs and compatibility issues with recent versions of the Kuberay operator.
- Kuberay Operator Version 0.4: In this version, the reconcile logic of the operator was changed. As a result, when running a Ray job, the status of the job will always remain "pending." It is important to note that this change was not thoroughly tested by the Kuberay community before its release.
- Kuberay Operator Version 0.5: In this version, an init container is injected into the worker nodes. However, the operator failed to set the required resource limits (CPU, memory) for this init container. Consequently, running Ray jobs becomes problematic in Flyte because Kubernetes imposes resource quotas in every project-domain namespace, and these quotas necessitate that every container sets the CPU and memory limits.

**Given these issues, it is crucial to be aware that the current Ray extension is compatible only with version 0.3.0 of the Kuberay operator.** It is recommended to use this specific version to ensure the proper functioning of the Ray extension.
:::

+++ {"lines_to_next_cell": 0}

First, import the required libraries.

```{code-cell}
import typing

import ray
from flytekit import Resources, task, workflow
from flytekitplugins.ray import HeadNodeConfig, RayJobConfig, WorkerNodeConfig
```

+++ {"lines_to_next_cell": 0}

Create a Ray example where a [remote function](https://docs.ray.io/en/latest/ray-core/tasks.html#tasks) is defined, which will be executed asynchronously in the Ray cluster.

```{code-cell}
@ray.remote
def f(x):
    return x * x
```

+++ {"lines_to_next_cell": 0}

Prepare a `HeadNodeConfig` and `WorkerNodeConfig` for the Ray job, which will be utilized by the Ray operator to launch a Ray cluster prior to executing the task.
* `ray_start_params`: [RayStartParams](https://docs.ray.io/en/latest/ray-core/package-ref.html#ray-start) are the parameters used in the start command, such as the address and object-store-memory.
* `replicas`: Specifies the desired number of replicas for the worker group. The default value is 1.
* `group_name`: A Ray cluster can consist of multiple worker groups, distinguished by their respective names.
* `runtime_env`: A [runtime environment](https://docs.ray.io/en/latest/ray-core/handling-dependencies.html#runtime-environments) defines the dependencies necessary for your Ray application to run. These dependencies are dynamically installed on the cluster at runtime.

```{code-cell}
ray_config = RayJobConfig(
    head_node_config=HeadNodeConfig(ray_start_params={"log-color": "True"}),
    worker_node_config=[WorkerNodeConfig(group_name="ray-group", replicas=2)],
    runtime_env={"pip": ["numpy", "pandas"]},  # or runtime_env="./requirements.txt"
)
```

+++ {"lines_to_next_cell": 0}

To execute the Ray tasks, utilize [Ray job submission](https://docs.ray.io/en/latest/cluster/job-submission.html#job-submission-architecture) architecture.
The `ray_task` function will be invoked on the Ray head node, while `f.remote(i)` will be asynchronously executed on individual Ray workers.

:::{note}
`Resources` configuration is employed to specify the resource requirements for the worker nodes.
:::

```{code-cell}
@task(task_config=ray_config, limits=Resources(mem="2000Mi", cpu="1"))
def ray_task(n: int) -> typing.List[int]:
    futures = [f.remote(i) for i in range(n)]
    return ray.get(futures)
```

+++ {"lines_to_next_cell": 0}

Finally establish a workflow to call the `ray_task`.

```{code-cell}
@workflow
def ray_workflow(n: int) -> typing.List[int]:
    return ray_task(n=n)
```

+++ {"lines_to_next_cell": 0}

Execute the code locally, wherein Flyte generates a self-contained Ray cluster locally.

```{code-cell}
if __name__ == "__main__":
    print(ray_workflow(n=10))
```

## Troubleshooting

If you encounter the situation where the head and worker pods are not being created, ensure that the ``ray[default]`` package is installed since it supports the cluster and dashboard launcher.

Another potential error you may come across is related to ingress errors in the kuberay-operator logs. If you observe an error message similar to the following:

````
ERROR controllers.RayCluster Ingress create error!
{
    "Ingress.Error": "Internal error occurred: failed calling webhook "validate.nginx.ingress.kubernetes.io": failed to call webhook: Post "<https://nginx-ingress-ingress-nginx-controller-admission.default.svc:443/networking/v1/ingresses?timeout=10s>": no endpoints available for service "nginx-ingress-ingress-nginx-controller-admission"",
    "error": "Internal error occurred: failed calling webhook "validate.nginx.ingress.kubernetes.io": failed to call webhook: Post "<https://nginx-ingress-ingress-nginx-controller-admission.default.svc:443/networking/v1/ingresses?timeout=10s>": no endpoints available for service "nginx-ingress-ingress-nginx-controller-admission""
}
````

You need to ensure that the ingress controller is properly installed. Refer to the documentation on how to install the ingress controller: [Installation Guide](https://docs.flyte.org/en/latest/deployment/gcp/manual.html#ingress).
