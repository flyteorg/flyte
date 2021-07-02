.. _deployment-cluster-config-performance:

######################################################
How do I optimize performance of my Flyte Deployment?
######################################################

.. tip:: Before getting started, it is always important to measure the performance. Flyte project publishes and manages some grafana templates as described in - :ref:`deployment-cluster-config-monitoring`.

Scaling up FlytePropeller
==========================
`FlytePropeller <https://pkg.go.dev/github.com/flyteorg/flytepropeller>`_ is the core engine of Flyte that executes the workflows for Flyte. It is implemented as a `controller <https://kubernetes.io/docs/concepts/architecture/controller/>`__ in Kubernetes.
FlytePropeller has a lot of knobs that can be tweaked for performance. The default configuration is good enough for small to medium sized installations of Flyte, that are running about 500 workflows concurrently with no noticeable overhead. In the case when the number of workflows increases,
FlytePropeller will automatically slow down, without losing correctness.

Here are signs of slowdown,

#. Round Latency for each workflow increases
#. Transition latency increases
#. Workflows take longer to start

This is usually because the number of threads in FlytePropeller are not enough to keep up with the number of workflows. This can be resolved by adjusting FlytePropeller config specified `here <https://pkg.go.dev/github.com/flyteorg/flytepropeller@v0.10.3/pkg/controller/config>`_.

.. list-table:: Important Properties
   :widths: 25 25 25 50
   :header-rows: 1

   * - Property
     - Section
     - Rule of thumb
     - Description
   * - workers
     - propeller
     - larger the number, more the performance (not too high)
     - Number of `logical threads` workers, that can work concurrently. Also implies number of workflows that can be executed in parallel. Since Flytepropeller uses go-routines, it can run way  more than number of physical cores.
   * - workflow-reeval-duration
     - propeller
     - lower the number - lower latency, lower throughput
     - frequency at which, given no external signal, a workflow should be re-evaluated by Flyte propellers reval loop
   * - downstream-eval-duration
     - propeller
     - lower the number - lower latency and throughput
     - This indicates how often are external events like pods completion etc recorded.
   * - max-streak-length
     - propeller
     - higher the number lower the latency for end to end workflow, epsecially for cached workflows
     - number of consecutive rounds to try with one workflow - prioritize a hot workflow over others.
   * - kube-client-config
     - propeller
     - this is a very important config
     - this configures the kubernetes client used by flytepropeller

In the above table the 2 most important configs are ``workers`` and ``kube-client-config``.

Kubeclient config is usually configured poorly and leads to bad performance. This is especially noticeable, if your workload involves spawning a lot of pods or other CRDs. For case in which your workload is a good mix of K8s local resources and external resources, the default for this configuration should suffice.
Flytepropeller configures a default version, which is better than the actual default. This configuration is critical, as this improves the number of requests that Flyte can send to KubeAPI, which is critical. An example kube-client-config is as follows

.. code-block:: yaml

    propeller:
        kube-client-config:
            qps: 100 # Refers to max rate of requests to KubeAPI server
            burst: 50 # refers to max burst rate to Kube API server
            timeout: 30s # Refers to timeout when talking with kubeapi server


.. note:: As you increase the number of workers in FlytePropeller it is important to increase the number of cpu's given to flytepropeller pod.


Another area of slowdown could be the size of the input-output cache that Flytepropeller maintains in-memory. This can be configured, while configuring
the storage for FlytePropeller.

Rule of thumb, for Flytepropeller that has 2GB or memory, should allocate about 512MB-756MB to the input/output cache.


Scaling out FlyteAdmin
=======================
Flyteadmin is a stateless service. Often time before needing to scale FlyteAdmin, you need to scale the backing database. Check out the FlyteAdmin Dashboard to see signs of latency degredation and increase the size of backing postgres instance.
FlyteAdmin is a stateless service and its replicas (in the kubernetes deployment) can be simply increased to allow higher throughput.

Scaling out Datacatalog
========================
Datacatalog is a stateless service. Often time before needing to scale Datacatalog, you need to scale the backing database. Check out the Datacatalog Dashboard to see signs of latency degredation and increase the size of backing postgres instance.
Datacatalog is a stateless service and its replicas (in the kubernetes deployment) can be simply increased to allow higher throughput.

Scaling out FlytePropeller
===========================
Flytepropeller can be sharded to work on a specific namespace of use consistent hashing to allow workflows to be handled by different instances.

.. caution:: Coming soon!

Multi-Cluster mode
===================
In our experience at Lyft, we saw that the Kubernetes cluster would have problems before Flytepropeller or Flyteadmin would have impact. Thus Flyte supports adding multiple dataplane clusters by default. Each dataplane cluster, has one or more flytepropellers running in them, and flyteadmin manages the routing and assigning of workloads to these clusters.