KubeRay
========

`KubeRay <https://github.com/ray-project/kuberay>`__ is an open source toolkit to run Ray applications on Kubernetes. It provides tools to improve running and managing Ray on Kubernetes.

- Ray Operator
- Backend services to create/delete cluster resources
- Kubectl plugin/CLI to operate CRD objects
- Native Job and Serving integration with Clusters

Installation
------------

To install the Ray plugin, run the following command:

.. code-block:: bash

    pip install flytekitplugins-ray

To enable the plugin in the backend, follow instructions outlined in the :std:ref:`flyte:deployment-plugin-setup-k8s` guide.

Submit a Ray Job to Existing Cluster
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
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


Create a Ray Cluster Managed by Flyte and Run a Ray Job on This Cluster
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
.. testcode:: ray-quickstart-2
    import ray
    from flytekit import task
    from flytekitplugins.ray import RayJobConfig, WorkerNodeConfig, HeadNodeConfig

    @task(task_config=RayJobConfig(worker_node_config=[WorkerNodeConfig(group_name="test-group", replicas=10)])
    def ray_task() -> typing.List[int]:
        futures = [f.remote(i) for i in range(5)]
        return ray.get(futures)

.. panels::
    :header: text-center
    :column: col-lg-12 p-2

    .. link-button:: https://blog.flyte.org/ray-and-flyte
       :type: url
       :text: Blog Post
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    An article detailing Ray and Flyte integration.

.. toctree::
    :maxdepth: -1
    :caption: Contents
    :hidden:

    Blog Post <https://blog.flyte.org/ray-and-flyte>
