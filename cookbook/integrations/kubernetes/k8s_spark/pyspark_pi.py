"""
.. _intermediate_using_spark_tasks:

Spark Tasks 
------------

Flyte has an optional plugin that makes it possible to run `Apache Spark <https://spark.apache.org/>`_ jobs natively on your kubernetes cluster. This plugin has been used extensively at Lyft and is battle tested.
It makes it extremely easy to run your pyspark (coming soon to scala/java) code as a task. The plugin creates a new virtual cluster for the spark execution dynamically and Flyte will manage the execution, auto-scaling
for the spark job.

.. NOTE::

   This has been tested at scale and more than 100k Spark Jobs run through Flyte at Lyft. This still needs a large capacity on Kubernetes and careful configuration.
   We recommend using Multi-cluster mode - :ref:`howto-multi-cluster`, and enabling :ref:`howto-resource-quota` for large and extremely frequent Spark Jobs.
   For extremely short running jobs, this is still not a recommended approach, and it might be better to use a pre-spawned cluster.

Why Use K8s Spark?
===================
Managing Python dependencies is hard. Flyte makes it easy to version and manage dependencies using containers. The K8s Spark plugin brings all the benefits of containerization
to spark without needing to manage special spark clusters.

Pros:
^^^^^
#. Extremely easy to get started and get complete isolation between workloads
#. Every job runs in isolation and has its own virtual cluster - no more nightmarish dependency management
#. Flyte manages everything for you!

Cons:
^^^^^
#. Short running, bursty jobs are not a great fit - because of the container overhead
#. No interactive spark capabilities available with Flyte K8s spark, which is more suited for running, adhoc and/or scheduled jobs


How to Enable Spark in the Flyte Backend
========================================
Flyte Spark uses the `Spark On K8s Operator <https://github.com/GoogleCloudPlatform/spark-on-k8s-operator>`_ and a custom built `Flyte Spark Plugin <https://pkg.go.dev/github.com/flyteorg/flyteplugins@v0.5.25/go/tasks/plugins/k8s/spark>`_.
The plugin is a backend plugin which has to be enabled in your deployment. To enable a plugin follow the steps in :ref:`howto-enable-backend-plugins`.

You can optionally configure the Plugin as per the - `backend Config Structure <https://pkg.go.dev/github.com/flyteorg/flyteplugins@v0.5.25/go/tasks/plugins/k8s/spark#Config>`_ and an example Config is defined
`here <https://github.com/flyteorg/flyte/blob/master/kustomize/overlays/sandbox/config/propeller/plugins/spark.yaml>`_, which looks like:

.. rli:: https://raw.githubusercontent.com/flyteorg/flyte/master/kustomize/overlays/sandbox/config/propeller/plugins/spark.yaml
   :language: yaml


Spark in Flytekit
=================
For a more complete example refer to the :std:ref:`User Guide <cookbook:sphx_glr_auto_integrations_kubernetes_k8s_spark>`

#. Ensure you have ``flytekit>=0.16.0``
#. Enable Spark in backend, following the previous section.
#. Install the `flytekit spark plugin <https://pypi.org/project/flytekitplugins-spark/>`_ ::

    pip install flytekitplugins-spark

#. Write regular pyspark code - with one change in ``@task`` decorator. Refer to the example below:

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

Examples
========

How Flytekit Simplifies Usage of Pyspark in a Users Code
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

The task ``hello_spark`` runs a new spark cluster, which when run locally runs a single node client only cluster,
but when run remote spins up a arbitrarily sized cluster depending on the specified spark configuration. ``spark_conf``

This Example also shows how a user can simply create 2 tasks, that use different Docker images. For more information refer to :any:`hosted_multi_images`

"""
import datetime
import random
from operator import add

import flytekit
from flytekit import task, workflow

# %%
# The following import is required to configure a Spark Server in Flyte:
from flytekitplugins.spark import Spark


# %%
# Spark Task Sample 
# ^^^^^^^^^^^^^^^^^^
# This example shows how a spark task can be written simply by adding a ``@task(task_config=Spark(...)...)`` decorator.
# Refer to :py:class:`flytekit.Spark` class to understand the various configuration options.
@task(
    task_config=Spark(
        # this configuration is applied to the spark cluster
        spark_conf={
            "spark.driver.memory": "1000M",
            "spark.executor.memory": "1000M",
            "spark.executor.cores": "1",
            "spark.executor.instances": "2",
            "spark.driver.cores": "1",
        }
    ),
    cache_version="1",
)
def hello_spark(partitions: int) -> float:
    print("Starting Spark with Partitions: {}".format(partitions))

    n = 100000 * partitions
    sess = flytekit.current_context().spark_session
    count = (
        sess.sparkContext.parallelize(range(1, n + 1), partitions).map(f).reduce(add)
    )
    pi_val = 4.0 * count / n
    print("Pi val is :{}".format(pi_val))
    return pi_val


def f(_):
    x = random.random() * 2 - 1
    y = random.random() * 2 - 1
    return 1 if x ** 2 + y ** 2 <= 1 else 0


# %%
# This is a regular python function task. This will not execute on the spark cluster
@task(cache_version="1")
def print_every_time(value_to_print: float, date_triggered: datetime.datetime) -> int:
    print("My printed value: {} @ {}".format(value_to_print, date_triggered))
    return 1


# %%
# The Workflow shows that a spark task and any python function (or any other task type) can be chained together as long as they match the parameter specifications.
@workflow
def my_spark(triggered_date: datetime.datetime) -> float:
    """
    Using the workflow is still as any other workflow. As image is a property of the task, the workflow does not care
    about how the image is configured.
    """
    pi = hello_spark(partitions=50)
    print_every_time(value_to_print=pi, date_triggered=triggered_date)
    return pi


# %%
# Workflows with spark tasks can be executed locally. Some aspects of spark, like links to plugins_hive metastores may not work, but these are limitations of using Spark and are not introduced by Flyte.
if __name__ == "__main__":
    """
    NOTE: To run a multi-image workflow locally, all dependencies of all the tasks should be installed, ignoring which
    may result in local runtime failures.
    """
    print(f"Running {__file__} main...")
    print(
        f"Running my_spark(triggered_date=datetime.datetime.now()){my_spark(triggered_date=datetime.datetime.now())}"
    )
