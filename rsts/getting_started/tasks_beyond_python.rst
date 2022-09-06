.. tasks-beyond-python:

###############
Tasks Beyond Python
###############

So far, there have been examples which show exectution of a Python task. But Flyte workflows are not limited to just Python tasks. Flyte can also execute other type of tasks. 
For example, On a Kubernetes Cluster, which controls the lifecycle, spin-up, and tear-down of virtual clusters, Flyte can run Spark jobs natively. It uses the Spark On K8s Operator, which is open-sourced, and may be used without registering for any services. This is comparable to using a transitory Spark cluster, which is a sort of cluster that is spun up for a particular Spark job and destroyed after it is finished. These clusters have higher setup and teardown costs but are better for production workloads.

With flytekit, PySpark code may be written directly as a task, and the Spark cluster will be set up automatically using the decorated SparkConf. 

With the help of an optional Flyte plugin, you can natively run Apache Spark tasks on your Kubernetes cluster. It makes running your PySpark code as a task really simple. Flyte will manage the execution, automatically scalability for the spark task, and the plugin dynamically builds a new virtual cluster for the spark execution.

Setup
^^^^^^^^^^^^

1. Ensure you have flytekit>=0.16.0
   
2. Enable and Deploy Spark by following this `guide <https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/kubernetes/k8s_spark/index.html>`__

3. Install the flytekit spark plugin
   
.. prompt:: bash

  pip install flytekit

4. Write standard PySpark code with the @task decorator changed. 

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
5. Run it locally

.. prompt:: bash

  hello_spark(partitions=10)

6. Use it in a workflow 
7. Run it on a remote cluster - To do this, you have to build the correct dockerfile, as explained here :std:ref:`spark-docker-image`. You can also use the `Standard Dockerfile recommended by Spark <https://github.com/apache/spark/blob/master/resource-managers/kubernetes/docker/src/main/dockerfiles/spark/Dockerfile#L22>`_.

Example
^^^^^^^^^^^^
The following import is required to configure a Spark Server in Flyte:

.. code-block:: python

  from flytekitplugins.spark import Spark

This example shows how a Spark task can be written simply by adding a ``@task(task_config=Spark(...)...)`` decorator.
Refer to `Spark <https://github.com/flyteorg/flytekit/blob/9e156bb0cf3d1441c7d1727729e8f9b4bbc3f168/plugins/flytekit-spark/flytekitplugins/spark/task.py#L18-L36>`__ class to understand the various configuration options.

The task ``hello_spark`` runs a new spark cluster, which when run locally runs a single node client only cluster, but when run remote spins up a arbitrarily sized cluster depending on the specified spark configuration. 

.. code-block:: python

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


    def f(_):
        x = random.random() * 2 - 1
        y = random.random() * 2 - 1
        return 1 if x**2 + y**2 <= 1 else 0

This is a regular python function task. This will not execute on the spark cluster

.. code-block:: python
    
    @task(cache_version="1")
    def print_every_time(value_to_print: float, date_triggered: datetime.datetime) -> int:
        print("My printed value: {} @ {}".format(value_to_print, date_triggered))
        return 1

The Workflow shows that a spark task and any python function (or any other task type) can be chained together as long as they match the parameter specifications.

.. code-block:: python
    
    @workflow
    def my_spark(triggered_date: datetime.datetime) -> float:
        """
        Using the workflow is still as any other workflow. As image is a property of the task, the workflow does not care
        about how the image is configured.
        """
        pi = hello_spark(partitions=50)
        print_every_time(value_to_print=pi, date_triggered=triggered_date)
        return pi

Workflows with spark tasks can be executed locally. Some aspects of spark, like links to plugins_hive metastores may not work, but these are limitations of using Spark and are not introduced by Flyte.

.. code-block:: python

    if __name__ == "__main__":
    """
    NOTE: To run a multi-image workflow locally, all dependencies of all the tasks should be installed, ignoring which
    may result in local runtime failures.
    """
    print(f"Running {__file__} main...")
    print(
        f"Running my_spark(triggered_date=datetime.datetime.now()){my_spark(triggered_date=datetime.datetime.now())}"
    )