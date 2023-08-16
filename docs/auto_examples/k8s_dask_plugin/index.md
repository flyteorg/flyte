(plugins-dask-k8s)=

# Kubernetes Dask Jobs

```{eval-rst}
.. tags:: Dask, Integration, DistributedComputing, Data, Advanced
```

Flyte can execute dask jobs natively on a Kubernetes Cluster, which manages a virtual `dask` cluster's lifecycle. To
do so, it leverages the open-sourced [Dask Kubernetes Operator](https://kubernetes.dask.org/en/latest/operator.html)
and can be enabled without signing up for any service. This is like running an ephemeral `dask` cluster, which gets
created for the specific Flyte task and gets torn down after completion.

In Flyte/K8s, the cost is amortized because pods are faster to create than a machine, but the penalty of downloading
Docker images may affect the performance. Also, remember that starting a pod is not as fast as running a process.

Flytekit makes it possible to write `dask` code natively as a task and the `dask` cluster will be automatically
configured using the decorated `Dask()` config. The examples in this section provide a hands-on tutorial for writing
`dask` Flyte tasks.

The plugin has been tested against the `2022.12.0` version of the `dask-kubernetes-operator`.

## Why use K8s dask?

Managing Python dependencies is hard. Flyte makes it easy to version  and manage dependencies using containers. The
K8s `dask` plugin brings all the benefits of containerization to `dask` without needing to manage special `dask`
clusters.

**Pros:**

1. Extremely easy to get started; get complete isolation between workloads
2. Every job runs in isolation and has its own virtual cluster - no more nightmarish dependency management!
3. Flyte manages everything for you!

**Cons:**

1. Short running, bursty jobs are not a great fit because of the container overhead
2. No interactive Dask capabilities are available with Flyte K8s dask, which is more suited for running adhoc and
   scheduled jobs.

## Step 1: Deploy the Dask Plugin in the Flyte Backend

Flyte dask uses the [Dask Kubernetes Operator](https://kubernetes.dask.org/en/latest/operator.html) and a custom
built [Flyte Dask Plugin](https://pkg.go.dev/github.com/flyteorg/flyteplugins@v1.0.28/go/tasks/plugins/k8s/dask).
This is a backend plugin which has to be enabled in your deployment; you can follow the steps mentioned in the
{ref}`flyte:deployment-plugin-setup-k8s` section.

## Step 2: Environment setup

1. Install `flytekitplugins-dask` using `pip` in your environment.

   ```bash
   pip install flytekitplugins-dask
   ```

2. Ensure you have enough resources on your K8s cluster. Based on the resources required for your `dask` job (across job runner, scheduler and workers), you may have to tweak resource quotas for the namespace.

## Implementation details

### Local execution

When running the `dask` task locally, it will use a local [distributed Client](https://distributed.dask.org/en/stable/client.html). In case you would like to connect the to a remote cluster for
when developing locally, you can set the `DASK_SCHEDULER_ADDRESS` environment variable to the URL of the remote
scheduler and the `Client()` will use the cluster automatically.

### Resource specification

It is advised to set `limits` as this will set the `--nthreads` and `--memory-limit` arguments for the workers
as recommended by `dask` [best practices](https://kubernetes.dask.org/en/latest/kubecluster.html?highlight=--nthreads#best-practices).
When specifying resources, the following precedence is followed for all components of the `dask` job (job-runner pod,
scheduler pod and worker pods):

1. If no resources are specified, the [platform resources](https://github.com/flyteorg/flyte/blob/1e3d515550cb338c2edb3919d79c6fa1f0da5a19/charts/flyte-core/values.yaml#L520-L531) are used

2. When `task` resources are used, those will be applied to all components of the `dask` job

   > ```python
   > from flytekit import Resources, task
   > from flytekitplugins.dask import Dask
   >
   > @task(
   >   task_config=Dask(),
   >   limits=Resources(cpu="1", mem="10Gi")  # Will be applied to all components
   > )
   > def my_dask_task():
   >    ...
   > ```

3. When resources are specified for the single components, they take the highest precedence

   > ```python
   > from flytekit import Resources, task
   > from flytekitplugins.dask import Dask, Scheduler, WorkerGroup
   >
   > @task(
   >   task_config=Dask(
   >       scheduler=Scheduler(
   >           limits=Resources(cpu="1", mem="2Gi"),  # Will be applied to the job pod
   >       ),
   >       workers=WorkerGroup(
   >           limits=Resources(cpu="4", mem="10Gi"), # Will be applied to the scheduler and worker pods
   >       ),
   >   ),
   > )
   > def my_dask_task():
   >    ...
   > ```

### Images

By default, all components of the deployed `dask` job (job runner pod, scheduler pod and worker pods) will all use the
the image that was used while registering (this image should have `dask[distributed]` installed in its Python
environment). This helps keeping the Python environments of all cluster components in sync.
However, there is the possibility to specify different images for the components. This allows for use cases such as using
different images between tasks of the same workflow. While it is possible to use different images for the different
components of the `dask` job, it is not advised, as this can quickly lead to Python environments getting our of sync.

> ```python
> from flytekit import Resources, task
> from flytekitplugins.dask import Dask, Scheduler, WorkerGroup
>
> @task(
>   task_config=Dask(
>       scheduler=Scheduler(
>           image="my_image:0.1.0",  # Will be used by the job pod
>       ),
>       workers=WorkerGroup(
>           image="my_image:0.1.0", # Will be used by the scheduler and worker pods
>       ),
>   ),
> )
> def my_dask_task():
>    ...
> ```

### Environment Variables

Environment variables set in the `@task` decorator will be passed on to all `dask` job components (job runner pod,
scheduler pod and worker pods)

> ```python
> from flytekit import Resources, task
> from flytekitplugins.dask import Dask
>
> @task(
>   task_config=Dask(),
>   env={"FOO": "BAR"}  # Will be applied to all components
> )
> def my_dask_task():
>    ...
> ```

### Labels and Annotations

Labels and annotations set in a `LaunchPlan` will be passed on to all `dask` job components (job runner pod,
scheduler pod and worker pods)

> ```python
> from flytekit import Resources, task, workflow, Labels, Annotations
> from flytekitplugins.dask import Dask
>
> @task(task_config=Dask())
> def my_dask_task():
>    ...
>
> @workflow
> def my_dask_workflow():
>    my_dask_task()
>
> # Labels and annotations will be passed on to all dask cluster components
> my_launch_plan = my_dask_workflow.create_launch_plan(
>   labels=Labels({"myexecutionlabel": "bar", ...}),
>   annotations=Annotations({"region": "SEA", ...}),
> )
> ```

### Interruptible Tasks

The `dask` backend plugin supports running on interruptible nodes. When `interruptible==True`, the plugin will add
the configured tolerations and node selectors to all worker pods. Please note that the job runner as well as the
scheduler will not be run on interruptible nodes.

> ```python
> from flytekit import Resources, task, workflow, Labels, Annotations
> from flytekitplugins.dask import Dask
>
> @task(
>   task_config=Dask(),
>   interruptible=True,
> )
> def my_dask_task():
>    ...
> ```

## Examples

```{auto-examples-toc}
dask_example
```
