.. _deployment-configuration-performance:

######################################################
Optimizing Performance
######################################################

.. tags:: Infrastructure, Kubernetes, Advanced

.. tip:: Before getting started, it is always important to measure the performance. Consider using the Grafana dashboard templates as described in :ref:`deployment-configuration-monitoring`.

Introduction
============

There are some base design attributes and assumptions that FlytePropeller applies:

- Every workflow execution is independent and can be performed by a completeley distinct process.
- When a workflow definition is compiled, the resulting DAG structure is traversed by the controller and the goal is to gracefully transition each task to ``Success``.
- Task executions are performed by various FlytePlugins; which perform operations on Kubernetes and other remote services as declared in the workflow definition. FlytePropeller is only responsible for effectively monitoring and managing these executions.

In the following sections you will learn how Flyte ensures the correct and reliable execution of workflows through multiple stages, and what strategies you can apply to help the system efficiently handle increasing load.

Summarized steps of a workflow execution
========================================

Let's revisit the lifecycle of a workflow execution. 
The following diagram aims to summarize the process described in the `FlytePropeller Architecture <https://docs.flyte.org/en/latest/concepts/component_architecture/flytepropeller_architecture.html>`__ and :ref:`divedeep-execution-timeline` sections, focusing on the main steps. 

.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/configuration/perf_optimization/propeller-perf-lifecycle-01.png

The ``Worker`` is the independent, lightweight, and idempotent process that interacts with all the components in the Propeller controller to drive executions. 
It's implemented as a ``goroutine``, and illustrated here as a hard-working gopher which:

1. Pulls from the ``WorkQueue`` and loads what it needs to do the job: the workflow specification (desired state) and the previously recorded execution status.
2. Observes the actual state by querying the Kubernetes API (or the Informer cache).
3. Calculates the difference between desired and observed state, and triggers an effect to reconcile both states (eg. Launch/kill a Pod, handle failures, schedule a node execution, etc), interacting with the Propeller executors to process inputs, outputs and offloaded data as indicated in the workflow spec.
4. Keeps a local copy of the execution status, besides what the K8s API stores in ``etcd``.
5. Reports status to the control plane and, hence, to the user.

This process is known as the "evaluation loop".
While there are multiple metrics that could indicate a slow down in execution performance, ``round_latency`` -or the time it takes FlytePropeller to complete a single evaluation loop- is typically the "golden signal". 
Optimizing ``round_latency`` is one of the main goals of the recommendations provided in the following sections.

Performance tuning at each stage
--------------------------------

1. Workers, the WorkQueue, and the evaluation loop 
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

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
     -  A low number may result in higher overall latency for each workflow evaluation loop, while a higher number implies that more workflows can be evaluated in parallel, reducing latency. The number of workers depends on the number of CPU cores assigned to the FlytePropeller pod, and should be evaluated against the cost of context switching. A number around 500 - 800 workers with 4-8 CPU cores is usually adequate.
     - ``plugins.workqueue.config.workers`` Default value: ``10``. 
   * - Workqueue depth
     - Current number of workflow IDs in the queue awaiting processing
     - ``sum(rate(flyte:propeller:all:main_depth[5m]))``
     - A growing trend indicates the processing queue depth is long and is taking longer to drain, delaying start time for executions.
     - ``plugins.workqueue.config.maxItems``. Default value: ``10000``

2. Query observed state
^^^^^^^^^^^^^^^^^^^^^^^

The Kube client config controls the request throughput from FlytePropeller to the Kube API server. These requests may include creating/monitoring pods or creating/updating FlyteWorkflow CRDs to track workflow execution. 
The `default configuration provided by K8s <https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client/config#GetConfigWithContext>`__ results in very conservative rate-limiting. FlytePropeller provides a default configuration that may offer better performance. 
However, if your workload involves larger scales (e.g., >5k fanout dynamic or map tasks, >8k concurrent workflows, etc.,) the kube-client rate limiting config provided by FlytePropeller may still contribute to a noticeable drop in performance. 
Increasing the ``qps`` and ``burst`` values may help alleviate back pressure and improve FlytePropeller performance. The following is an example kube-client config applied to Propeller:

.. code-block:: yaml

    propeller:
      kube-client-config:
        qps: 100 # Refers to max rate of requests (queries per second) to kube-apiserver
        burst: 120 # refers to max burst rate. 
        timeout: 30s # Refers to timeout when talking with the kube-apiserver

.. note::

   In the previous example, the kube-apiserver will accept ``100`` queries per second, temporariliy admitting up to ``120`` before blocking any subsequent query. A query blocked for ``30s`` will timeout.

It is worth noting that the Kube API server tends to throttle requests transparently. This means that even after increasing the allowed frequency of API requests (e.g., increasing FlytePropeller workers or relaxing Kube client config rate-limiting), there may be steep performance decreases for no apparent reason. 
While it's possible to easily monitor Kube API saturation using system-level metrics like CPU, memory, and network usage, we recommend looking at kube-apiserver-specific metrics like ``workqueue_depth`` which can assist in identifying whether throttling is to blame. Unfortunately, there is no one-size-fits-all solution here, and customizing these parameters for your workload will require trial and error.
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
     -  A large value can lead to faster completion times for workflows that benefit from continuous processing, especially cached or computationally intensive workflows, but at the cost of lower throughput and higher latency as workers will spend most of their time on a few workflows. If set to ``1``, the worker adds the workflowID back to the WorkQueue immediately after a single evaluation loop is completed, and waits for another worker to pick it up before processing again, effectively prioritizing fast-changing or "hot" workflows.
     -  ``propeller.max-streak-length``. Default value: ``8`` . 
   * - ``max-size_mbs``
     - Max size of the write-through in-memory cache that FlytePropeller can use to store Inputs/Outputs metadata for faster read operations. 
     - A too-small cache might lead to frequent cache misses, reducing the effectiveness of the cache and increasing latency. Conversely, a too-large cache might consume too much memory, potentially affecting the performance of other components. We recommend monitoring cache performance metrics such as `hit rates and miss rates <https://github.com/flyteorg/flyte/blob/8cc96177e7447d9630a1186215a8c8ad3d34d4a2/deployment/stats/prometheus/flytepropeller-dashboard.json#L1140>`__. These metrics can help determine if the cache size needs to be adjusted for optimal performance. 
     - ``storage.cache.max-size_mbs``. Default value: ``0`` (disabled).
   * - ``backoff.max-duration``
     - Maximum back-off interval in case of resource-quota errors.
     - A higher value will ensure retries do not happen too frequently, which could overwhelm resources or overload the Kubernetes API server at the cost of overall latency.
     - ``tasks.backoff.max-duration``. Default value: ``20s``.


4. Record execution status
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
     - The default policy is designed to leverage ``etcd`` features to reduce latency. 
     - ``propeller.workflowStore.policy``. Default value: ``ResourceVersionCache``.

**How ``ResourceVersionCache`` works?**

.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/configuration/perf_optimization/resourceversion-01.png 

Kubernetes stores the definition and state of all the resources under its management on ``etcd``: a fast, distributed and consistent key-value store.
Every resource has a ``resourceVersion`` field representing the version of that resource as stored in ``etcd``. 

Example:

.. code-block:: bash

   kubectl get datacatalog-589586b67f-l6v58 -n flyte -o yaml

Sample output (excerpt):

.. code-block:: yaml

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

Every time a resource (e.g. a pod, a flyteworkflow CR, etc.) is modified, this counter is incremented.
As ``etcd`` is a distributed key-value store, it needs to manage writes from multiple clients (controllers in this case)
in a way that maintains consistency and performance.
That's why, in addition to using ``Revisions`` (implemented in Kubernetes as ``Resource Version``), ``etcd`` also prevents clients from writing if they're using
an outdated ``ResourceVersion``, which could happen after a temporary client disconnection or whenever a status replication from the Kubernetes API to 
the Informer cache hasn't completed yet. Poorly handled by a controller, this could result in kube-server and FlytePropeller worker overload by repeatedly attempting to perform outdated (or "stale") writes.

FlytePropeller handles these situations by keeping a record of the last known ``ResourceVersion``. In the event that ``etcd`` denies a write operation due to an outdated version, FlytePropeller continues the workflow
evaluation loop, waiting for the Informer cache to become consistent. This mechanism, enabled by default and known as ``ResourceVersionCache``, avoids both overloading the K8s API and wasting ``workers`` resources on invalid operations.
It also mitigates the impact of cache propagation latency, which can be on the order of seconds.

If ``max-streak-length`` is enabled, instead of waiting for the Informer cache to become consistent during the evaluation loop, FlytePropeller runs multiple evaluation loops using its in-memory copy of the ``ResourceVersion`` and corresponding Resource state, as long 
as there are mutations in any of the resources associated with that particular workflow. When the ``max-streak-length`` limit is reached, the evaluation loop is done and, if further evaluation is required, the cycle will start again by trying to get the most recent ``Resource Version`` as stored in ``etcd``.

Other supported options for ``workflowStore.policy`` are described below:

- ``InMemory``: utilizes an in-memory store for workflows, primarily for testing purposes.
- ``PassThrough``: directly interacts with the underlying Kubernetes clientset or shared informer cache for workflow operations.
- ``TrackTerminated``: specifically tracks terminated workflows.

5. Report status to the control plane
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. list-table:: Important Properties
   :widths: 25 50 50
   :header-rows: 1

   * - Property
     - Description
     - Impact on performance
   * - ``admin-launcher.tps``, ``admin-launcher.cacheSize``, ``admin-launcher.workers`` 
     - Configure the maximum rate and number of launchplans that FlytePropeller can launch against FlyteAdmin.
     - It is important to limit the number of writes from FlytePropeller to FlyteAdmin to prevent brown-outs or request throttling at the server. Also a bigger cache size, reduces number of calls to the server.

Concurrency vs parallelism
==========================

While FlytePropeller is designed to efficiently handle concurrency using the mechanisms described in this section, parallel executions (not only concurrent, but evaluated at the same time) pose an additional challenge, especially with workflows that have an extremely large fan-out. 
This is because FlytePropeller implements a greedy traversal algorithm, that tries to evaluate all unblocked nodes within a workflow in every round.
A way to mitigate the potential performance impact is to limit the maximum number of nodes that can be evaluated simultaneously. This can be done by setting ``max-parallelism`` using any of the following methods:

a. Platform default: This allows to set platform-wide defaults for maximum parallelism within a Workflow execution evaluation loop. This can be overridden per launch plan or per execution.
   The default `maxParallelism is configured to be 25 <https://github.com/flyteorg/flyteadmin/blob/master/pkg/runtime/application_config_provider.go#L40>`_.
   It can be overridden with this config block in flyteadmin

   .. code-block:: yaml

       flyteadmin:
          maxParallelism: 25

b. Default for a specific launch plan. For any launch plan, the ``max_parallelism`` value can be changed using :py:meth:`flytekit.LaunchPlan.get_or_create` or the :std:ref:`ref_flyteidl.admin.LaunchPlanCreateRequest`
   **Flytekit Example**

   .. code-block:: python

       LaunchPlan.get_or_create(
         name="my_cron_scheduled_lp",
         workflow=date_formatter_wf,
         max_parallelism=30,
       )

#. Specify for an execution. ``max-parallelism`` can be overridden using ``pyflyte run --max-parallelism`` or by setting it in the UI.


Scaling out FlyteAdmin
=======================
FlyteAdmin is a stateless service. Often, before needing to scale FlyteAdmin, you need to scale the backing database. 
Check the `FlyteAdmin Dashboard <https://github.com/flyteorg/flyte/blob/master/deployment/stats/prometheus/flyteadmin-dashboard.json>`__  for signs of database or API latency degradation.
PostgreSQL scaling techniques like connection pooling can help alleviate pressure on the database instance.
If needed, change the number of replicas of the FlyteAdmin K8s deployment to allow higher throughput.

Scaling out Datacatalog
========================
Datacatalog is a stateless service that connects to the same database as FlyteAdmin, so the recommendation to scale out the backing PostgreSQL database also applies here.

Scaling out FlytePropeller
===========================

Sharded scale-out
-------------------
FlytePropeller Manager facilitates horizontal scaling of FlytePropeller through sharding. Effectively, the Manager is responsible for maintaining liveness and proper configuration over a collection of FlytePropeller instances. This scheme uses K8s label selectors to deterministically assign FlyteWorkflow CRD responsibilities to FlytePropeller instances, effectively distributing load processing over the shards.

Deployment of FlytePropeller Manager requires K8s configuration updates including a modified FlytePropeller deployment and a new PodTemplate defining managed FlytePropeller instances. The easiest way to apply these updates is to set the ``flytepropeller.manager`` value to ``true`` in the Helm values and set the manager config at ``configmap.core.manager``.

Flyte provides a variety of shard strategies to configure how FlyteWorkflows are sharded among managed FlytePropeller instances. These include ``hash``, which uses consistent hashing to load balance evaluation over shards, and ``project`` / ``domain``, which map the respective IDs to specific managed FlytePropeller instances. Below we include examples of Helm configurations for each of the existing shard strategies.

The hash shard Strategy, denoted by ``type: Hash`` in the configuration below, uses consistent hashing to evenly distribute Flyte workflows over managed FlytePropeller instances. This configuration requires a ``shard-count`` variable, which defines the number of managed FlytePropeller instances. You may change the shard count without impacting existing workflows. Note that changing the ``shard-count`` is a manual step; it is not auto-scaling.

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
 
The project and domain shard strategies, denoted by ``type: project`` and ``type: domain`` respectively, use the Flyte workflow project and domain metadata to shard Flyte workflows. These shard strategies are configured using a ``per-shard-mapping`` option, which is a list of IDs. Each element in the ``per-shard-mapping`` list defines a new shard, and the ID list assigns responsibility for the specified IDs to that shard. A shard configured as a single wildcard ID (i.e. ``*``) is responsible for all IDs that are not covered by other shards. Only a single shard may be configured with a wildcard ID and, on that shard, there must be only one ID, namely the wildcard.

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

If the K8s cluster itself becomes a performance bottleneck, Flyte supports adding multiple K8s dataplane clusters by default. Each dataplane cluster has one or more FlytePropellers running in it, and flyteadmin manages the routing and assigning of workloads to these clusters.


Improving etcd Performance
===========================

Offloading Static Workflow Information from CRD
-----------------------------------------------

Flyte uses a K8s CRD (Custom Resource Definition) to store and track workflow executions. This resource includes the workflow definition, the tasks and subworkflows that are involved, and the dependencies between nodes. It also includes the execution status of the workflow. The latter information (i.e. runtime status) is dynamic, and changes during the workflow's execution as nodes transition phases and the workflow execution progresses. However, the former information (i.e. workflow definition) remains static, meaning it will never change and is only consulted to retrieve node definitions and workflow dependencies.

CRDs are stored within ``etcd``, which requires a complete rewrite of the value data every time a single field changes. Consequently, the read / write performance of ``etcd``, as with all key-value stores, is strongly correlated with the size of the data. In Flyte's case, to guarantee only-once execution of nodes, we need to persist workflow state by updating the CRD at every node phase change. As the size of a workflow increases this means we are frequently rewriting a large CRD. In addition to poor read / write performance in ``etcd``, these updates may be restricted by a hard limit on the overall CRD size.

To counter the challenges of large FlyteWorkflow CRDs, Flyte includes a configuration option to offload the static portions of the CRD (ie. workflow / task / subworkflow definitions and node dependencies) to the S3-compliant blobstore. This functionality can be enabled by setting the ``useOffloadedWorkflowClosure`` option to ``true`` in the `FlyteAdmin configuration <https://docs.flyte.org/en/latest/deployment/cluster_config/flyteadmin_config.html#useoffloadedworkflowclosure-bool>`_. When set, the FlyteWorkflow CRD will populate a ``WorkflowClosureReference`` field on the CRD with the location of the static data and FlytePropeller will read this information (through a cache) during each workflow evaluation. One important note is that currently this setting requires FlyteAdmin and FlytePropeller to have access to the same blobstore since FlyteAdmin only specifies a blobstore location in the CRD.
