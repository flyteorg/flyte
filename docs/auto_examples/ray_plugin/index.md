(kube-ray-op)=

# KubeRay

```{tags} Integration, DistributedComputing, KubernetesOperator, Advanced

```

```{image} https://img.shields.io/badge/Blog-Ray-blue?style=for-the-badge
:target: https://blog.flyte.org/ray-and-flyte
:alt: Ray Blog Post
```

[KubeRay](https://github.com/ray-project/kuberay) is an open-source toolkit designed for running Ray applications on Kubernetes. It offers various tools to enhance the execution and management of Ray on Kubernetes.

- Ray Operator
- Backend services for creating and deleting cluster resources
- Kubectl plugin/CLI for interacting with CRD objects
- Native integration with Clusters for Job and Serving functionalities

## Installation

To install the Ray plugin, run the following command:

```bash
pip install flytekitplugins-ray
```

To enable the plugin in the backend, refer to the instructions provided in the {std:ref}`flyte:deployment-plugin-setup-k8s` guide.

Here are two quick examples that give you an overview of the integration.
If you're interested in a more detailed exploration, please refer to the example page for further details.

### Submitting a Ray job to an existing cluster

```{eval-rst}
.. testcode:: ray-quickstart-1
    import ray
    from flytekit import task
    from flytekitplugins.ray import RayJobConfig

    @ray.remote
    def f(x):
        return x * x

    @task(task_config=RayJobConfig(
        address=<RAY_CLUSTER_ADDRESS>
        runtime_env={"pip": ["numpy", "pandas"]})
    )
    def ray_task() -> typing.List[int]:
        futures = [f.remote(i) for i in range(5)]
        return ray.get(futures)

```

### Creating a Ray cluster managed by Flyte and running a Ray job on it

```{eval-rst}
.. testcode:: ray-quickstart-2
    import ray
    from flytekit import task
    from flytekitplugins.ray import RayJobConfig, WorkerNodeConfig, HeadNodeConfig

    @task(task_config=RayJobConfig(worker_node_config=[WorkerNodeConfig(group_name="test-group", replicas=10)])
    def ray_task() -> typing.List[int]:
        futures = [f.remote(i) for i in range(5)]
        return ray.get(futures)
```

```{auto-examples-toc}
ray_example
```
