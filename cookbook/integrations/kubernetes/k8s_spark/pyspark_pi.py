"""
.. _intermediate_using_spark_tasks:

Writing a PySpark Task
----------------------

The Spark plugin makes it extremely easy to run your PySpark code as a task.
The plugin creates a new ephemeral cluster for the Spark execution dynamically, and Flyte manages the execution and auto-scaling.

Spark in Flytekit
=================

#. Ensure you have ``flytekit`` installed.
#. Enable Spark in the backend by following the :ref:`flyte:deployment-plugin-setup-k8s` guide.
#. Install the `flytekit spark plugin <https://pypi.org/project/flytekitplugins-spark/>`__. ::

    pip install flytekitplugins-spark

#. Write regular PySpark code.

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
           # Regular PySpark code
           ...

#. Run it locally.

   .. code-block:: python

       hello_spark(partitions=10)

#. Use it in a Flyte workflow.
#. Run it on a remote cluster by building a Docker image using the Dockerfile in the README.

.. _example-spark:

How Flytekit Simplifies Usage of Pyspark
========================================

The task ``hello_spark`` runs a new Spark cluster, which when run locally runs a single node client only cluster,
but when run remotely spins up an arbitrarily-sized cluster depending on the specified spark configuration.
"""

# %%
# Let's get started by importing the libraries.
import datetime
import random
from operator import add

import flytekit
from flytekit import Resources, task, workflow

from flytekitplugins.spark import Spark

# %%
# You can create a Spark task by adding a ``@task(task_config=Spark(...)...)`` decorator.
# ``spark_conf`` can have configuration options that are typically used when configuring a Spark cluster.
# ``hadoop_conf`` can also be given as an input if required.
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
    limits=Resources(mem="2000M"),
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

# %%
# Let's define a function on which the map-reduce operation is called within the Spark cluster.
def f(_):
    x = random.random() * 2 - 1
    y = random.random() * 2 - 1
    return 1 if x**2 + y**2 <= 1 else 0


# %%
# Next, we define a regular Flyte task which will not execute on the Spark cluster.
@task(cache_version="1")
def print_every_time(value_to_print: float, date_triggered: datetime.datetime) -> int:
    print("My printed value: {} @ {}".format(value_to_print, date_triggered))
    return 1


# %%
# This workflow shows that a spark task and any python function (or a Flyte task) can be chained together as long as they match the parameter specifications.
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
# Workflows with spark tasks can be executed locally. Some aspects of spark, like links to :ref:`Hive <Hive>` metastores may not work, but these are limitations of using Spark and are not introduced by Flyte.
if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(
        f"Running my_spark(triggered_date=datetime.datetime.now()){my_spark(triggered_date=datetime.datetime.now())}"
    )
