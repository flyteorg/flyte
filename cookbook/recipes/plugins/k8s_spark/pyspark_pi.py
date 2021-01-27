"""
.. _intermediate_using_spark_tasks:

Creating spark tasks as part of your workflow OR running spark jobs
------------------------------------------------------------------------

This example shows how flytekit simplifies usage of pyspark in a users code.
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
# The follow import is required to configure a Spark Server in Flyte.
from flytekitplugins.spark import Spark

from recipes.plugins.k8s_spark.dataframe_passing import CONTAINER_TEMPLATE


# %%
# Spark Task sample. This example shows how a spark task can be written simply by adding a ``@task(task_config=Spark(...)...)`` decorator.
# Refer to :py:class:`flytekit.Spark` class to understand the various configuration options.
# Also important to note here that the container_image is a special image that is built as part of the samples repo. To understand how to configure
# different containers per task refer to :any:`hosted_multi_images`.
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
    # a separate image is used
    container_image=CONTAINER_TEMPLATE,
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
@task(cache_version="1", container_image=CONTAINER_TEMPLATE)
def print_every_time(value_to_print: float, date_triggered: datetime.datetime) -> int:
    print("My printed value: {} @ {}".format(value_to_print, date_triggered))
    return 1


# %%
# The Workflow shows that a spark task and any python function (or any other task type) can be chained together as long as they match the parameter specifications
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
# Workflows with spark tasks can be executed locally. Some aspects of spark, like links to plugins_hive metastores etc may not work, but these are limitations of using Spark and are not introduced by Flyte.
if __name__ == "__main__":
    """
    NOTE: To run a multi-image workflow locally, all dependencies of all the tasks should be installed, ignoring which
    may result in local runtime failures.
    """
    print(f"Running {__file__} main...")
    print(
        f"Running my_spark(triggered_date=datetime.datetime.now()){my_spark(triggered_date=datetime.datetime.now())}"
    )
