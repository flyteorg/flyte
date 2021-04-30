.. _plugins-spark-k8s:

########################################
Run Spark on your Kubernetes Cluster
########################################

.. tip:: If you just looking for examples of spark on flyte - refer to :std:ref:`Cookbook Spark Plugin <sphx_glr_auto_plugins_k8s_spark>`


Flyte has an optional plugin that makes it possible to run `Apache Spark <https://spark.apache.org/>`_ jobs native on your kubernetes cluster. This plugin has been used extensively at Lyft and is battle tested.
This makes it extremely easy to run your pyspark (coming soon scala/java) code as a task. The plugin creates a new virtual cluster for the spark execution dynamically and Flyte will manage the execution, auto-scaling
for the spark job.

.. NOTE::

   This has been tested at scale and more than 100k Spark Jobs run through Flyte at Lyft. This still needs a large capacity on Kubernetes and careful configuration.
   We recommend using Multi-cluster mode - :ref:`howto-multi-cluster`, and enabling :ref:`howto-resource-quota` for large and extremely frequent Spark Jobs.
   For extremely short running jobs, this is still not a recommended approach, and it might be better to use a pre-spawned cluster.

Why use K8s Spark?
===================
Managing Python dependencies is hard. Flyte makes it easy to version and manage dependencies using Containers. K8s Spark plugin brings all the benefits of containerization
to spark and without needing to manage special spark clusters.

Pros:
------
#. Extremely easy to get started and get complete isolation between workloads
#. Every job runs in isolation and has its own virtual cluster - no more nightmarish dependency management
#. Flyte manages everything for you!

Cons:
-----
#. Short running, bursty jobs are not a great fit - because of the container overhead
#. No interactive spark capabilities available with Flyte K8s spark which is more suited for running, adhoc and/or scheduled jobs


How to enable Spark in flyte backend?
======================================
Flyte Spark uses the `Spark On K8s Operator <https://github.com/GoogleCloudPlatform/spark-on-k8s-operator>`_ and a custom built `Flyte Spark Plugin <https://pkg.go.dev/github.com/flyteorg/flyteplugins@v0.5.25/go/tasks/plugins/k8s/spark>`_.
The plugin is a backend plugin and you have to enable it in your deployment. To enable a plugin follow the steps in :ref:`howto-enable-backend-plugins`.

You can optionally configure the Plugin as per the - `backend Config Structure <https://pkg.go.dev/github.com/flyteorg/flyteplugins@v0.5.25/go/tasks/plugins/k8s/spark#Config>`_ and an example Config is defined
`here <https://github.com/flyteorg/flyte/blob/master/kustomize/overlays/sandbox/config/propeller/plugins/spark.yaml>`_, which looks like,

.. rli:: https://raw.githubusercontent.com/flyteorg/flyte/master/kustomize/overlays/sandbox/config/propeller/plugins/spark.yaml
   :language: yaml


Spark in Flytekit
========================
For a more complete example refer to :std:ref:`Cookbook Spark Plugin <sphx_glr_auto_plugins_k8s_spark>`

#. Ensure you have ``flytekit>=0.16.0``
#. Enable Spark in backend, following the previous section.
#. Install the `flytekit spark plugin <https://pypi.org/project/flytekitplugins-spark/>`_ ::

    pip install flytekitplugins-spark

#. Write regular pyspark code - with one change in ``@task`` decorator. Refer to the example

    .. code-block:: python

        @task(
            task_config=Spark(
                # this configuration is applied to the spark cluster
                spark_conf={
                    "spark.driver.memory": "1000M",
                    "spark.executor.instances": "2",
                    "spark.driver.cores": "1",
                }
            ),
            cache_version="1",
            cache=True,
        )
        def hello_spark(partitions: int) -> float:
            ...
            sess = flytekit.current_context().spark_session
            # Regular Pypsark code
            ...


#. Run it locally

    .. code-block:: python

        hello_spark(partitions=10)

#. Use it in a workflow (check cookbook)
#. Run it on a remote cluster - To do this, you have to build the correct dockerfile, as explained here :std:ref:`spark-docker-image`. You can also you the `Standard Dockerfile recommended by Spark <https://github.com/apache/spark/blob/master/resource-managers/kubernetes/docker/src/main/dockerfiles/spark/Dockerfile#L22>`_.
