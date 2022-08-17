"""
Ray Tasks
----------
Ray task allows you to run a Ray job on an existing Ray cluster or create a Ray cluster by using the Ray operator.


Let's get started with an example!
"""

# %%
# First, we load the libraries.
import typing

import ray
from flytekit import Resources, task, workflow
from flytekitplugins.ray import HeadNodeConfig, RayJobConfig, WorkerNodeConfig


# %%
# Ray Task
# =========
#
# We define a ray_example `remote function <https://docs.ray.io/en/latest/ray-core/tasks.html#tasks>`__ that will be executed asynchronously in the Ray cluster.
@ray.remote
def f(x):
    return x * x


# %%
# Defining a Ray Config
# ====================
#
# We create a HeadNodeConfig and WorkerNodeConfig for the Ray job, and these config will be used by Ray operator to launch a Ray cluster before running the task.
#
# * ``ray_start_params``: `RayStartParams <https://docs.ray.io/en/latest/ray-core/package-ref.html#ray-start>`__ are the params of the start command: address, object-store-memory
# * ``replicas``: Desired replicas of the worker group. Defaults to 1.
# * ``group_name``: RayCluster can have multiple worker groups, and it distinguishes them by name
# * ``runtime_env``: A `runtime environment <https://docs.ray.io/en/latest/ray-core/handling-dependencies.html#runtime-environments>`__ describes the dependencies your Ray application needs to run, and it's installed dynamically on the cluster at runtime.
#
ray_config = RayJobConfig(
    head_node_config=HeadNodeConfig(ray_start_params={"log-color": "True"}),
    worker_node_config=[WorkerNodeConfig(group_name="ray-group", replicas=5)],
    runtime_env={"pip": ["numpy", "pandas"]},  # or runtime_env="./requirements.txt"
)


# %%
# Defining a Ray Task
# ===================
# We use `Ray job submission <https://docs.ray.io/en/latest/cluster/job-submission.html#job-submission-architecture>`__ to run our ray_example tasks.
# ray_task will be called in the Ray head node, and f.remote(i) will be executed asynchronously on separate Ray workers
#
# .. note::
#    The Resources here is used to define the resource of worker nodes
@task(task_config=ray_config, limits=Resources(mem="2000Mi", cpu="2"))
def ray_task(n: int) -> typing.List[int]:
    futures = [f.remote(i) for i in range(n)]
    return ray.get(futures)


# %%
# Workflow
# ========
#
# Finally we define a workflow to call the ``ray_workflow`` task.
@workflow
def ray_workflow(n: int) -> typing.List[int]:
    return ray_task(n=n)


# %%
# We can run the code locally wherein Flyte creates a standalone Ray cluster locally.
if __name__ == "__main__":
    print(ray_workflow(n=10))
