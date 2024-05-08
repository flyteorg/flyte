.. _deployment-configuration-performance:

######################################################
Optimizing Performance
######################################################

.. tags:: Infrastructure, Kubernetes, Advanced

.. tip:: Before getting started, it is always important to measure the performance. Flyte project publishes Grafana dashboard templates as described in - :ref:`deployment-configuration-monitoring`.

Introduction
============

There are some base design attributes and assumptions that FlytePropeller applies:

a. Every workflow execution is independent and can be performed by a completeley distinct process.
b. When a workflow definition is compiled, the resulting DAG structure is traversed by the controller and the goal is to gracefully transition each task to Success.
c. Node executions are performed by various FlytePlugins; a diverse collection of operations spanning Kubernetes and other remote services. FlytePropeller is only responsible for effectively monitoring and managing these executions.

In the following sections you will learn how Flyte takes care of the correct and reliable execution of workflows through multiple stages, and what strategies you can apply to help the system efficiently handle increasing load.

Summarized steps of a workflow execution
========================================

Let's revisit the lifecycle of a workflow execution. 
The following diagram aims to summarize the process described in the `FlytePropeller Architecture <https://docs.flyte.org/en/latest/concepts/component_architecture/flytepropeller_architecture.html>`__ and :ref:`divedeep-execution-timeline` sections, focusing on the main steps. 

.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/configuration/perf_optimization/propeller-perf-lifecycle-01.png

The ``Worker`` is the independent, lightweight and idempotent process that interacts with all the components in the Propeller controller to drive executions. 
It's implemented as a ``goroutine``, and illustrated here as a hard-working gopher which:

1. Pulls from the ``WorkQueue`` and loads what it needs to do the job: the workflow specification (desired state) and the previously recorded execution status.
2. Observes the actual state by querying the Kubernetes API (or the Informer cache).
3. Calculates the difference between desired and observed state, and triggers an effect to reconcile both states (eg. Launch/kill a Pod, handle failures, schedule a node execution, etc), interacting with the Propeller executors to process Inputs, Outputs and Offloaded data as indicated in the workflow spec.
4. Keeps a local copy of the execution status, besides what the K8s API stores in ``etcd``.
5. Reports status to the control plane and, hence, to the user.

While there are multiple metrics that could indicate a slow down in execution performance, ``round_latency`` -or the time it takes FlytePropeller to perform a single iteration of workflow evaluation- is typically the "golden signal". 
Optimizing ``round_latency`` is one of the main goals of the recommendations provided in the following sections.

Performance tuning at each stage
--------------------------------

1. Workers, the WorkQueue and the evaluation loop 
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. list-table:: Important Properties
   :widths: 25 50 25 50 25
   :header-rows: 1

   * - Property
     - Description
     - Relevant metric
     - Impact on performance
     - Configuration parameter
   * - ``workers``
     - Number of processes that can work concurrently. Also implies number of workflows that can be executed in parallel. Since FlytePropeller uses ``goroutines``, it can accommodate significantly more processes than the number of physical cores.
     - ``flyte:propeller:all:free_workers_count``
     -  A low number may result in higher overall latency for each workflow evaluation round. Larger the number, implies more workflows can be evaluated in parallel. But it should depend on number of CPU cores assigned to FlytePropeller and evaluated against the cost of context switching. A number around 500 - 800 with 4-8 CPU cores usually works fine.
     - ``plugins.workqueue.config.workers`` Default value: ``10``. 
   * - Workqueue depth
     - Current number of workflow IDs in the queue awaiting processing
     - ``sum(rate(flyte:propeller:all:main_depth[5m]))``
     - A growing trend indicates the processing queue depth is long and is taking longer to drain, delaying start time for executions.
     - ``plugins.workqueue.config.maxItems``. Default value: ``10000``

2. Query observed state
^^^^^^^^^^^^^^^^^^^^^^^

The Kube client config controls the request throughput from FlytePropeller to the Kube API server. These requests may include creating/monitoring Pods or creating/updating FlyteWorkflow CRDs to track workflow execution. 
The `default configuration provided by K8s <https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client/config#GetConfigWithContext>`__ results in very conservative rate-limiting. FlytePropeller provides a default configuration that may offer better performance. 
However, if your workload involves larger scales (e.g., >5k fanout dynamic or map tasks, >8k concurrent workflows, etc.,) the kube-client rate limiting config may still contribute to a noticeable drop in performance. 
Increasing the ``qps`` and ``burst`` values may help alleviate back pressure and improve FlytePropeller performance. The default kube-client config applied to Propeller is as follows:

.. code-block:: yaml

    propeller:
      kube-client-config:
        qps: 100 # Refers to max rate of requests (queries per second) to kube-apiserver
        burst: 25 # refers to max burst rate. 
        timeout: 30s # Refers to timeout when talking with the kube-apiserver

.. note::

   In the previous configuration, the kube-apiserver will accept ``100`` queries before blocking any query. Every second, ``25`` more queries will be accepted. A query blocked for ``30s`` will timeout.

It is worth noting that the Kube API server tends to throttle requests transparently. This means that even increasing the allowed frequency of API requests (e.g., increasing FlytePropeller workers or relaxing Kube client config rate-limiting), there may be steep performance decreases for no apparent reason. 
While it's possible to easily monitor Kube API saturation using system-level metrics like CPU, memory and network usage; it's recommended to look at kube-apiserver-specific metrics like ``workqueue_depth`` which can assist in identifying whether throttling is to blame. Unfortunately, there is no one-size-fits-all solution here, and customizing these parameters for your workload will require trial and error.
`Learn more about Kubernetes metrics <https://kubernetes.io/docs/reference/instrumentation/metrics/>`__

3. Evaluate the DAG and reconcile state as needed
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. list-table:: Important Properties
   :widths: 25 50 50 25
   :header-rows: 1

   * - Property
     - Description
     - Impact on performance
     - Configuration parameter
   * - ``workflow-reeval-duration``
     - Interval at which the system re-evaluates the state of a workflow when no external events have triggered a state change. This periodic re-evaluation helps in progressing workflows that may be waiting on conditions or timeouts to be met.
     - A shorter duration means workflows are checked more frequently, which can lead to quicker progression through workflow steps but at the cost of increased load on the system. Conversely, a longer duration reduces system load but may delay the progression of workflows.
     - ``propeller.workflow-reeval-duration``. Default value: ``10s``.
   * - ``downstream-eval-duration`` 
     - Interval at which the system checks for updates on the execution status of downstream tasks within a workflow. This setting is crucial for workflows where tasks are interdependent, as it determines how quickly Flyte reacts to changes or completions of tasks that other tasks depend on.
     - A shorter interval makes Flyte check more frequently for task updates, which can lead to quicker workflow progression if tasks complete faster than anticipated, at the cost of higher system load and reduced throughput.  Conversely, a higher value reduces the frequency of checks, which can decrease system load but may delay the progression of workflows, as the system reacts slower to task completions.
     - ``propeller.downstream-eval-duration``. Default value: ``5s``.
   * - ``max-streak-length``
     -  Maximum number of consecutive evaluation rounds that one propeller worker can use for one workflow. 
     -  A large ``max-streak-length`` value can lead to faster completion times for workflows that benefit from continuous processing, especially cached or computationally intensive workflows; at the cost of overall lower throughput and higher latency as workers would be spending most of their time on a few workflows. If set to `1`, the worker adds the workflowID back to the WorkQueue immediately after a single evaluation loop is completed and waits for another worker to pick it up before processing again, effectively prioritizing "hot workflows".
     -  ``propeller.max-streak-length``. Default value: ``8`` . 
   * - ``max-size_mbs``
     - Max size of the write-through in-memory cache that FlytePropeller can use to store Inputs/Outputs metadata for faster read operations. 
     - A too-small cache might lead to frequent cache misses, reducing the effectiveness of the cache and increasing latency. Conversely, a too-large cache might consume too much memory, potentially affecting the performance of other components. We recommend monitoring cache performance metrics such as `hit rates and miss rates <https://github.com/flyteorg/flyte/blob/8cc96177e7447d9630a1186215a8c8ad3d34d4a2/deployment/stats/prometheus/flytepropeller-dashboard.json#L1140>`__. These metrics can help determine if the cache size needs to be adjusted for optimal performance. 
     - ``storage.cache.max-size_mbs``. Default value: ``0`` (disabled).
   * - ``backoff.max-duration``
     - Maximum back-off interval in case of resource-quota errors.
     - A higher value will ensure retries do not happen too frequently, which could overwhelm resources or overload the Kubernetes API server at the cost of overall latency.
     - ``tasks.backoff.max-duration``. Default value: ``20s``.


4. Update execution status
^^^^^^^^^^^^^^^^^^^^^^^^^^

.. list-table:: Important Properties
   :widths: 25 50 50 25
   :header-rows: 1

   * - Property
     - Description
     - Impact on performance
     - Configuration parameter
   * - ``workflowStore Policy``
     - Specifies the strategy for workflow storage management.
     - This config uses a trick in etcD to minimize number of redundant loops in FlytePropeller, thus improving free slots.
     - ``propeller.workflowStore.policy``. Default value: ``ResourceVersionCache``.

**How ResourceVersionCache works?**

.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/configuration/perf_optimization/resourceversion-01.png 

Kubernetes stores the definition and state of all the resources under its management on ``etcd``, a fast, distributed and consistent key-value store.
Every resource has a ``resourceVersion`` field representing the version of that resource as stored in ``etcd``. 

Example:

.. code:block:: bash

   kubectl get datacatalog-589586b67f-l6v58 -n flyte -o yaml

Sample output (excerpt):

.. code:block:: yaml

    apiVersion: v1
    kind: Pod
    metadata:
      ...
      labels:
        app.kubernetes.io/instance: flyte-core
        app.kubernetes.io/managed-by: Helm
        app.kubernetes.io/name: datacatalog
        helm.sh/chart: flyte-core-v1.12.0
      name: datacatalog-589586b67f-l6v58
      namespace: flyte
    ...
      resourceVersion: "1055227"

Every time a resource (e.g. a Pod, a flyteworkflow CR, etc.) is modified, this counter is incremented.
As ``etcd`` is a distributed key-value store, it needs a way to manage writes from multiple clients (controllers in this case)
in a way that maintains consistency and performance.
That's why, in addition to using ``Revisions`` (implemented in Kubernetes as ``Resource Version``), ``etcd`` also prevents clients from writing if they're using
an outdated ``ResourceVersion``; something that could happen after a temporary client disconnection or whenever a status replication from the Kubernetes API to 
the Informer cache hasn't completed yet. Poorly handled by a controller, this could result into kube-server and FlytePropeller worker overload by repeatedly attempting to perform outdated (or "stale") writes.

FlytePropeller handles these situations by keeping a record of the last known ``ResourceVersion``. In the event that ``etcd`` denies a write operation due to an outdated version, FlytePropeller continues the Workflow
evaluation loop, waiting for the Informer cache to become consistent. This mechanism, enabled by default and known as ``ResourceVersionCache``, prevents from both overloading the K8s API and wasting ``workers`` resources on invalid operations.
It also mitigates the impact of cache propagation latency, which can be order of seconds.

If ``max-streak-length`` is enabled, instead of waiting for the Informer cache to become consistent during the evaluation loop, FlytePropeller runs multiple evaluation loops using its in-memory copy of the ``ResourceVersion`` and corresponding Resource state, as long 
as there are mutations in any of the resources associated with that particular workflow. When the ``max-streak-length`` limit is reached, the evaluation loop is done and, if further evaluation is required, the cycle will start again by trying to get the most recent ``Resource Version`` as stored in ``etcd``.

Other supported options for ``workflowStore.policy`` are described below:

- ``InMemory``: utilizes an in-memory store for workflows, primarily for testing purposes.
- ``PassThrough``: directly interacts with the underlying Kubernetes clientset or shared informer cache for workflow operations.
- ``TrackTerminated``: Specifically tracks terminated workflows.


5. Report status to the control plane
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. list-table:: Important Properties
   :widths: 25 50 25 50 25
   :header-rows: 1
* - Property
     - Description
     - Impact on performance
     - Configuration parameter
   * - ``admin-launcher.tps``, ``admin-launcher.cacheSize``, ``admin-launcher.workers``
     - propeller
     - This config is used to configure the max rate and launch-plans that FlytePropeller can launch against FlyteAdmin
     - It is essential to limit the number of writes from FlytePropeller to flyteadmin to prevent brown-outs or request throttling at the server. Also the cache reduces number of calls to the server.

   

   * - ``max-parallelism``
     - admin, per workflow, per execution
     - Refer to examples and documentation below
     - docs below



Concurrency vs parallelism
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
The worst case for FlytePropeller is workflows that have an extremely large fan-out. This is because FlytePropeller implements a greedy traversal algorithm, that tries to evaluate the entire unblocked nodes within a workflow in every round.
A solution for this is to limit the maximum number of nodes that can be evaluated. This can be done by setting max-parallelism for an execution.
This can done in multiple ways

#. Platform default: This allows to set platform-wide defaults for maximum concurrency within a Workflow execution. This can be overridden per Launch plan or per execution.
   The default `maxParallelism is configured to be 25 <https://github.com/flyteorg/flyteadmin/blob/master/pkg/runtime/application_config_provider.go#L40>`_.
   It can be overridden with this config block in flyteadmin

   .. code-block:: yaml

       flyteadmin:
          maxParallelism: 25

#. Default for a specific launch plan. For any launch plan, the maxParallelism value can be changed or altered. This can be done using :py:meth:`flytekit.LaunchPlan.get_or_create` or the :std:ref:`ref_flyteidl.admin.LaunchPlanCreateRequest`
   **Flytekit Example**

   .. code-block:: python

       LaunchPlan.get_or_create(
         name="my_cron_scheduled_lp",
         workflow=date_formatter_wf,
         max_parallelism=30,
       )

#. Specify for an execution. For any specific execution the max-parallelism can be overridden. This can be done using flytectl (and soon flyteconsole). Refer to :std:ref:`flyteCtl docs <flytectl:flytectl_create_execution>`




Scaling out FlyteAdmin
=======================
FlyteAdmin is a stateless service. Often time before needing to scale FlyteAdmin, you need to scale the backing database. Check out the FlyteAdmin Dashboard to see signs of latency degradation and increase the size of backing postgres instance.
FlyteAdmin is a stateless service and its replicas (in the kubernetes deployment) can be simply increased to allow higher throughput.

Scaling out Datacatalog
========================
Datacatalog is a stateless service. Often time before needing to scale Datacatalog, you need to scale the backing database. Check out the Datacatalog Dashboard to see signs of latency degradation and increase the size of backing postgres instance.
Datacatalog is a stateless service and its replicas (in the kubernetes deployment) can be simply increased to allow higher throughput.

Scaling out FlytePropeller
===========================

Manual scale-out
----------------
FlytePropeller can be run manually per namespace. This is not a recommended solution as it is harder to deploy, but if your organization can deploy and maintain multiple copies of FlytePropeller, you can use this.

Sharded scale-out
-------------------
FlytePropeller Manager is a new component introduced as part of `this RFC <https://github.com/flyteorg/flyte/blob/master/rfc/system/1483-flytepropeller-horizontal-scaling.md>`_ to facilitate horizontal scaling of FlytePropeller through sharding. Effectively, the Manager is responsible for maintaining liveness and proper configuration over a collection of FlytePropeller instances. This scheme uses k8s label selectors to deterministically assign FlyteWorkflow CRD responsibilities to FlytePropeller instances, effectively distributing processing load over the shards.

Deployment of FlytePropeller Manager requires k8s configuration updates including a modified FlytePropeller Deployment and a new PodTemplate defining managed FlytePropeller instances. The easiest way to apply these updates is by setting the "flytepropeller.manager" value to "true" in the `helm deployment <https://docs.flyte.org/en/latest/deployment/overview.html#usage-of-helm>`_ and setting the manager config at "configmap.core.manager".

Flyte provides a variety of Shard Strategies to configure how FlyteWorkflows are sharded among managed FlytePropeller instances. These include hash, which uses consistent hashing to load-balance evaluation over shards, and project / domain, which map the respective IDs to specific managed FlytePropeller instances. Below we include examples of helm configurations for each of the existing Shard Strategies.

The Hash Shard Strategy, denoted by "type: Hash" in the configuration below, uses consistent hashing to evenly distribute FlyteWorkflows over managed FlytePropeller instances. This configuration requires a "shard-count" variable which defines the number of managed FlytePropeller instances. You may change the shard count without impacting existing workflows. Note that changing the shard-count is a manual step, it is not auto-scaling.

.. code-block:: yaml

    configmap:
      core:
        # a configuration example using the "hash" shard type
        manager:
          # pod and scanning configuration redacted
          # ...
          shard:
            type: Hash     # use the "hash" shard strategy
            shard-count: 4 # the total number of shards
 
The Project and Domain Shard Strategies, denoted by "type: project" and "type: domain" respectively, use the FlyteWorkflow project and domain metadata to shard FlyteWorkflows. These Shard Strategies are configured using a "per-shard-mapping" option, which is a list of ID lists. Each element in the "per-shard-mapping" list defines a new shard and the ID list assigns responsibility for the specified IDs to that shard. A shard configured as a single wildcard ID (i.e. "*") is responsible for all IDs that are not covered by other shards. Only a single shard may be configured with a wildcard ID and on that shard their must be only one ID, namely the wildcard.

.. code-block:: yaml

    configmap:
      core:
        # a configuration example using the "project" shard type
        manager:
          # pod and scanning configuration redacted
          # ...
          shard:
            type: project       # use the "project" shard strategy
            per-shard-mapping:  # a list of per shard mappings - one shard is created for each element
              - ids:            # the list of ids to be managed by the first shard
                - flytesnacks
              - ids:            # the list of ids to be managed by the second shard
                - flyteexamples
                - flytelabs
              - ids:            # the list of ids to be managed by the third shard
                - "*"           # use the wildcard to manage all ids not managed by other shards
    
    configmap:
      core:
        # a configuration example using the "domain" shard type
        manager:
          # pod and scanning configuration redacted
          # ...
          shard:
            type: domain        # use the "domain" shard strategy
            per-shard-mapping:  # a list of per shard mappings - one shard is created for each element
              - ids:            # the list of ids to be managed by the first shard
                - production
              - ids:            # the list of ids to be managed by the second shard
                - "*"           # use the wildcard to manage all ids not managed by other shards
 
Multi-Cluster mode
===================
In our experience at Lyft, we saw that the Kubernetes cluster would have problems before FlytePropeller or FlyteAdmin would have impact. Thus Flyte supports adding multiple dataplane clusters by default. Each dataplane cluster, has one or more FlytePropellers running in them, and flyteadmin manages the routing and assigning of workloads to these clusters.


Improving etcd Performance
===========================

Offloading Static Workflow Information from CRD
-----------------------------------------------

Flyte uses a k8s CRD (Custom Resource Definition) to store and track workflow executions. This resource includes the workflow definition, for example tasks and subworkflows that are involved and the dependencies between nodes, but also includes the execution status of the workflow. The latter information (ie. runtime status) is dynamic, meaning changes during the workflow's execution as nodes transition phases and the workflow execution progresses. However, the former information (ie. workflow definition) remains static, meaning it will never change and is only consulted to retrieve node definitions and workflow dependencies.

CRDs are stored within etcd, a key-value datastore heavily used in kubernetes. Etcd requires a complete rewrite of the value data every time a single field changes. Consequently, the read / write performance of etcd, as with all key-value stores, is strongly correlated with the size of the data. In Flyte's case, to guarantee only-once execution of nodes we need to persist workflow state by updating the CRD at every node phase change. As the size of a workflow increases this means we are frequently rewriting a large CRD. In addition to poor read / write performance in etcd this update may be restricted by a hard limit on the overall CRD size.

To counter the challenges of large FlyteWorkflow CRDs Flyte includes a configuration option to offload the static portions of the CRD (ie. workflow / task / subworkflow definitions and node dependencies) to the blobstore. This functionality can be enabled by setting the ``useOffloadedWorkflowClosure`` option to ``true`` in the `FlyteAdmin configuration <https://docs.flyte.org/en/latest/deployment/cluster_config/flyteadmin_config.html#useoffloadedworkflowclosure-bool>`_. When set, the FlyteWorkflow CRD will populate a ``WorkflowClosureReference`` field on the CRD with the location of the static data and FlytePropeller will read this information (through a cache) during each workflow evaluation. One important note is that currently this requires FlyteAdmin and FlytePropeller to have access to the same blobstore since FlyteAdmin only specifies a blobstore location in the CRD.
