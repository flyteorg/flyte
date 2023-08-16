---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.14.7
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

(spark_on_databricks)=

# Writing a PySpark Task

The Spark plugin makes it extremely easy to run your PySpark code as a task.
The plugin creates a new ephemeral cluster for the Spark execution dynamically, and Flyte manages the execution and auto-scaling.

## Spark in Flytekit

1. Ensure you have `flytekit` installed.

2. Enable Spark in the backend by following the {ref}`flyte:deployment-plugin-setup-k8s` guide.

3. Install the [flytekit spark plugin](https://pypi.org/project/flytekitplugins-spark/).

   ```bash
   pip install flytekitplugins-spark
   ```

4. Run remotely: to correctly kick off an execution of this task, you'll need to use the following command.

   ```bash
   pyflyte --config ~/.flyte/your-config.yaml run --destination-dir /app --remote --image ghcr.io/flyteorg/flytekit:py3.8-sqlalchemy-latest integrations/flytekit_plugins/sql/sql_alchemy.py my_wf --min_length 3 --max_length 100 --limit 50
   ```

Note also we added the `destination-dir` argument, since by default `pyflyte run` copies code into `/root` which
is not what that image's workdir is set to.

## How Flytekit Simplifies Usage of Pyspark

The task `hello_spark` runs a new Spark cluster, which when run locally runs a single node client only cluster,
but when run remotely spins up an arbitrarily-sized cluster depending on the specified spark configuration.

+++ {"lines_to_next_cell": 0}

Let's get started by importing the libraries.

```{code-cell}
import datetime
import random
from operator import add

import flytekit
from flytekit import Resources, task, workflow
from flytekitplugins.spark import Databricks
```

+++ {"lines_to_next_cell": 0}

You can create a Spark task by adding a `@task(task_config=Spark(...)...)` decorator.
`spark_conf` can have configuration options that are typically used when configuring a Spark cluster.
To run a Spark job on Databricks platform, just add Databricks config to the task config. Databricks Config is the same as the databricks job request.
Refer to [Databricks job request](https://docs.databricks.com/dev-tools/api/2.0/jobs.html#request-structure)

```{code-cell}
@task(
    task_config=Databricks(
        # this configuration is applied to the spark cluster
        spark_conf={
            "spark.driver.memory": "1000M",
            "spark.executor.memory": "1000M",
            "spark.executor.cores": "1",
            "spark.executor.instances": "2",
            "spark.driver.cores": "1",
        },
        databricks_conf={
            "run_name": "flytekit databricks plugin example",
            "new_cluster": {
                "spark_version": "11.0.x-scala2.12",
                "node_type_id": "r3.xlarge",
                "aws_attributes": {
                    "availability": "ON_DEMAND",
                    "instance_profile_arn": "arn:aws:iam::1237657460:instance-profile/databricks-s3-role",
                },
                "num_workers": 4,
            },
            "timeout_seconds": 3600,
            "max_retries": 1,
        },
    ),
    limits=Resources(mem="2000M"),
    cache_version="1",
)
def hello_spark(partitions: int) -> float:
    print("Starting Spark with Partitions: {}".format(partitions))

    n = 100000 * partitions
    sess = flytekit.current_context().spark_session
    count = sess.sparkContext.parallelize(range(1, n + 1), partitions).map(f).reduce(add)
    pi_val = 4.0 * count / n
    print("Pi val is :{}".format(pi_val))
    return pi_val
```

+++ {"lines_to_next_cell": 0}

Let's define a function on which the map-reduce operation is called within the Spark cluster.

```{code-cell}
def f(_):
    x = random.random() * 2 - 1
    y = random.random() * 2 - 1
    return 1 if x**2 + y**2 <= 1 else 0
```

+++ {"lines_to_next_cell": 0}

Next, we define a regular Flyte task which will not execute on the Spark cluster.

```{code-cell}
@task(cache_version="1")
def print_every_time(value_to_print: float, date_triggered: datetime.datetime) -> int:
    print("My printed value: {} @ {}".format(value_to_print, date_triggered))
    return 1
```

+++ {"lines_to_next_cell": 0}

This workflow shows that a spark task and any python function (or a Flyte task) can be chained together as long as they match the parameter specifications.

```{code-cell}
@workflow
def my_databricks_job(
    triggered_date: datetime.datetime = datetime.datetime.now(),
) -> float:
    """
    Using the workflow is still as any other workflow. As image is a property of the task, the workflow does not care
    about how the image is configured.
    """
    pi = hello_spark(partitions=50)
    print_every_time(value_to_print=pi, date_triggered=triggered_date)
    return pi
```

+++ {"lines_to_next_cell": 0}

Workflows with spark tasks can be executed locally. Some aspects of spark, like links to {ref}`Hive <Hive>` meta stores may not work, but these are limitations of using Spark and are not introduced by Flyte.

```{code-cell}
if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(
        f"Running To run a Spark job on Databricks platform(triggered_date=datetime.datetime.now()){my_databricks_job(triggered_date=datetime.datetime.now())}"
    )
```
