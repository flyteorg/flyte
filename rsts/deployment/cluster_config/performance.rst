.. _deployment-cluster-config-performance:

######################################################
How do I optimize performance of my Flyte Deployment?
######################################################

.. tip:: Before getting started, it is always important to measure the performance. Flyte project publishes and manages some grafana templates as described in - :ref:`deployment-cluster-config-monitoring`.

The video below contains an overview of the Flyte architecture, what is meant by "performance", details of one loop in FlytePropeller, and a demo of the Grafana Labs dashboard.

..  youtube:: FJ-rG9lZDhY 

Scaling up FlytePropeller
==========================
`FlytePropeller <https://pkg.go.dev/github.com/flyteorg/FlytePropeller>`_ is the core engine of Flyte that executes the workflows for Flyte.
It is designed as a `Kubernetes Controller <https://kubernetes.io/docs/concepts/architecture/controller/>`_, where the desired state is specified as a FlyteWorkflow `Custom Resource <https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/>`_.

One of the base assumptions of FlytePropeller is that every workflow is independent and can be executed by a completely distinct process, without a need for communication with other processes. Meanwhile, one workflow tracks the dependencies between tasks using a DAG structure and hence constantly needs synchronization.
Currently, FlytePropeller executes Workflows by using an event loop to periodically track and amend the execution status. Within each iteration, a single thread requests the state of Workflow nodes and performs operations (i.e., scheduling node executions, handling failures, etc) to gracefully transition a Workflow from the observed state to the desired state (i.e., Success). Consequently, actual node executions are performed by various FlytePlugins, a diverse collection of operations spanning k8s and other remote services, and FlytePropeller is only responsible for effectively monitoring and managing these executions.

FlytePropeller has a lot of knobs that can be tweaked for performance. The default configuration is usually adequate for small to medium sized installations of Flyte, that are running about 500 workflows concurrently with no noticeable overhead. In the case when the number of workflows increases,
FlytePropeller will automatically slow down, without losing correctness.

Signs of slowdown
------------------

.. list-table:: Important metrics to look out for in Prometheus dashboard
   :widths: 25 25 50 100
   :header-rows: 1

   * - Metric
     - Dashboard
     - Alerting factor
     - Effect
   * - flyte:propeller:all:free_workers_count
     - Flyte Propeller Dashboard
     - Number of free-workers in all FlytePropeller instances are very low.
     - This will increase overall latency for Each workflow propagation
   * - sum(rate(flyte:propeller:all:round:raw_ms[5m])) by (wf)
     - Flyte Propeller Dashboard
     - Round Latency for Each workflow increases
     - Flyte propeller is taking more time to process each workflow
   * - sum(rate(flyte:propeller:all:main_depth[5m]))
     - Flyte Propeller Dashboard
     - Workflows take longer to start or slow
     - The processing queue depth is long and is taking long to drain

For each of the metrics above you can dig deeper into the cause for this spike in latency. All of them are mostly related to one latency and should be correlated - ``The Round latency!``.
The round-latency is the time it takes for FlytePropeller to to perform a single iteration of workflow evaluation. To understand this, you have to understand the :ref:`divedeep-execution-timeline`. Once you understand this, continue reading on.

Optimizing round latency
-------------------------

Optimize FlytePropeller configuration
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Usually round-latency can be resolved by adjusting FlytePropeller config specified `here <https://pkg.go.dev/github.com/flyteorg/FlytePropeller@v0.10.3/pkg/controller/config>`_ or sometimes adjusting the global defaults for your deployment or per workflow-execution.
Let us first look at various config properties that can be set and would impact the performance of one round in Flyte and should be tweaked

.. list-table:: Important Properties
   :widths: 25 25 25 50
   :header-rows: 1

   * - Property
     - Section
     - Rule of thumb
     - Description
   * - workers
     - propeller
     - Larger the number, implies more workflows can be evaluated in parallel. But it should depend on number of CPU cores assigned to FlytePropeller and evaluated against the cost of context swtiching. A number usually < 500 - 800 with 4-8 cpu cores works fine.
     - Number of `logical threads` workers, that can work concurrently. Also implies number of workflows that can be executed in parallel. Since FlytePropeller uses go-routines, it can run way  more than number of physical cores.
   * - workflow-reeval-duration
     - propeller
     - lower the number - lower latency, lower throughput (low throughput is because the same workflow will be evaluated more times)
     - frequency at which, given no external signal, a workflow should be re-evaluated by Flyte propellers reval loop
   * - downstream-eval-duration
     - propeller
     - lower the number - lower latency and lower throughput (low throughput is because the same workflow will be evaluated more times)
     - This indicates how often are external events like pods completion etc recorded.
   * - max-streak-length
     - propeller
     - higher the number lower the latency for end to end workflow, especially for cached workflows
     - number of consecutive rounds to try with one workflow - prioritize a hot workflow over others.
   * - kube-client-config
     - propeller
     - This is how you can control the number of requests ceiling that FlytePropeller can initiate to KubeAPI. This is usual the #1 bottle neck
     - this configures the kubernetes client used by FlytePropeller
   * - workflowStore.policy
     - propeller
     - This config uses a trick in etcD to minimize number of redundant loops in FlytePropeller, thus improving free slots
     - Use this to configure how FlytePropeller should evaluate workflows, the default is usually a good choice
   * - storage.cache
     - propeller
     - This config is used to configure the write-through cache used by FlytePropeller on top of the metastore
     - FlytePropeller uses the configure blob-store (can be changed to something more performant in the future) to optimize read and write latency, for all metadata IO operations. Metadata refers to the input and output pointers
   * - admin-launcher.tps, admin-launcher.cacheSize, admin-launcher.workers
     - propeller
     - This config is used to configure the max rate and launch-plans that FlytePropeller can launch against FlyteAdmin
     - It is essential to limit the number of writes from FlytePropeller to flyteadmin to prevent brown-outs or request throttling at the server. Also the cache reduces number of calls to the server.
   * - tasks.backoff.max-duration
     - propeller
     - This config is used to configure the maximum back-off interval incase of resource-quota errors
     - FlytePropeller will automatically back-off when k8s or other services request it to slowdown or when desired quotas are met.
   * - max-parallelism
     - admin, per workflow, per execution
     - Refer to examples and documentation below
     - docs below


In the above table the 2 most important configs are ``workers`` and ``kube-client-config``.
Kubeclient config is usually configured poorly and leads to bad performance. This is especially noticeable if your workload involves spawning a lot of pods or other CRDs. When the workload is a good mix of K8s local resources and external resources, the default for this configuration should suffice.
FlytePropeller configures a default version, which is better than the actual default. This configuration is critical, as this improves the number of requests that Flyte can send to KubeAPI, which is critical. An example kube-client-config is as follows



.. code-block:: yaml

    propeller:
        kube-client-config:
            qps: 100 # Refers to max rate of requests to KubeAPI server
            burst: 50 # refers to max burst rate to Kube API server
            timeout: 30s # Refers to timeout when talking with kubeapi server


.. note:: As you increase the number of workers in FlytePropeller it is important to increase the number of cpu's given to FlytePropeller pod.

Another area of slowdown could be the size of the input-output cache that FlytePropeller maintains in-memory. This can be configured, while configuring
the storage for FlytePropeller. Rule of thumb, for FlytePropeller with x memory limit, allocate x/2 to the cache

Learn: max-streak-length & ResourceVersionCaching
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Kubernetes controllers often use Informer caches, rather than reading data directly from KubeAPI. This is to prevent excessive requests to KubeAPI server. The caches are eventually consistent, i.e., every write by the controller is eventually replicated to the cache, but there can be time periods, when the cache lags.
Since FlytePropeller, runs Workflow evaluations as an event loop, which is triggered by any changes to one of the resources that a workflow spawned.
It is possible that a Workflow will be evaluated, even when the last write has not yet propagated to the Informer cache. EtcD also does not allow stale writes, i.e., writes with an object that is older than the object that was written. This is maintained using a server side vector-clock - called the resource version.
Stale writes are writes when the evaluation resulted in a mutation of an object that is older than the object recorded in etcD.
These stale writes often lead to conflicts and hence increase load on the KubeAPI server and on FlytePropeller as the workers are busy writing stale objects repeatedly.

To prevent this duplication and redundancy, FlytePropeller employs a trick. For every write, it records the last known version number in the database and then tries to wait for the change to propagate to the informer cache.

If `max-streaks` are enabled then instead of waiting for the informer cache to be refreshed, FlytePropeller uses its own inmemory copy to run multiple rounds as long as mutations occur or the max-streak-length configuration is met. This reduces the latency of cache propagation, which can be order of seconds.


Worst case workflows: Poison Pills & max-parallelism
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
The worst case for FlytePropeller is workflows that have an extremely large fan-out. This is because FlytePropeller implements a greedy traversal algorithm, that tries to evaluate the entire unblocked nodes within a workflow in every round.
A solution for this is to limit the maximum number of nodes that can be evaluated. This can be done by setting max-parallelism for an execution.
This can done in multiple ways

#. Platform default
#. Default for a specific launch plan
#. Specify for an execution


Manual Scale out FlytePropeller
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
FlytePropeller can be run manually per namespace. This is not a recommended solution as it is harder to deploy, but if your organization can deploy and maintain multiple copies of FlytePropeller, you can use this.

Automatic scale-out: coming soon
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

TODO link to RFC

Scaling out FlyteAdmin
=======================
FlyteAdmin is a stateless service. Often time before needing to scale FlyteAdmin, you need to scale the backing database. Check out the FlyteAdmin Dashboard to see signs of latency degredation and increase the size of backing postgres instance.
FlyteAdmin is a stateless service and its replicas (in the kubernetes deployment) can be simply increased to allow higher throughput.

Scaling out Datacatalog
========================
Datacatalog is a stateless service. Often time before needing to scale Datacatalog, you need to scale the backing database. Check out the Datacatalog Dashboard to see signs of latency degredation and increase the size of backing postgres instance.
Datacatalog is a stateless service and its replicas (in the kubernetes deployment) can be simply increased to allow higher throughput.

Scaling out FlytePropeller
===========================
FlytePropeller can be sharded to work on a specific namespace of use consistent hashing to allow workflows to be handled by different instances.

.. caution:: Coming soon!

Multi-Cluster mode
===================
In our experience at Lyft, we saw that the Kubernetes cluster would have problems before FlytePropeller or FlyteAdmin would have impact. Thus Flyte supports adding multiple dataplane clusters by default. Each dataplane cluster, has one or more FlytePropellers running in them, and flyteadmin manages the routing and assigning of workloads to these clusters.
