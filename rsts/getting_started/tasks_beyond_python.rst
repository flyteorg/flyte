.. tasks-beyond-python:

###############
Tasks Beyond Python
###############

So far, there have been examples which show exectution of a Python task. But Flyte workflows are not limited to just Python tasks. Flyte can also execute other type of tasks. 
For example, On a Kubernetes Cluster, which controls the lifecycle, spin-up, and tear-down of virtual clusters, Flyte can run Spark jobs natively. It uses the Spark On K8s Operator, which is open-sourced, and may be used without registering for any services. This is comparable to using a transitory Spark cluster, which is a sort of cluster that is spun up for a particular Spark job and destroyed after it is finished. These clusters have higher setup and teardown costs but are better for production workloads.

With flytekit, PySpark code may be written directly as a task, and the Spark cluster will be set up automatically using the decorated SparkConf. 

With the help of an optional Flyte plugin, you can natively run Apache Spark tasks on your Kubernetes cluster. It makes running your PySpark code as a task really simple. Flyte will manage the execution, automatically scalability for the spark task, and the plugin dynamically builds a new virtual cluster for the spark execution.

Let's set it up: 

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