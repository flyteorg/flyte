.. _spark-task-type:

Spark Task
==========

Spark Task Type allows users to write a Spark Job which are then executed on the Kubernetes cluster via the `SparkOperator`_.

#######
 Setup 
#######

Spark on Kubernetes is containerized i.e. your own repo image is run as Spark driver and executors hence we need to have Spark and it's dependencies installed in the user's container. 

One way to achieve this is to build relevant images for your use-case using `Dockerfile`_ which already exist as part of the Spark distribution. `Flytekit`_ also provides *flytekit_install_spark.sh* and *flytekit_spark_entrypoint.sh* which can be used to install spark dependencies/set the correct entrypoint for Spark Tasks in an user image.

Please refer to `Spark Documentation`_ for details on Spark on K8s.

#######
Usage
#######

The following is an example of a simple Spark job which calculates the value of Pi.

.. code-block:: python
   :caption: Simple Spark Job example 
   
   @spark_task(
       spark_conf={
         'spark.driver.memory': "1000M",
         'spark.executor.memory': "1000M",
         'spark.executor.cores': '1',
         'spark.executor.instances': '2',
       },
    )
    def hello_spark(workflow_parameters, spark_context, partitions, out):
        print("Starting Spark with Partitions: {}".format(partitions))
        n = 100000 * partitions
        count = spark_context.parallelize(range(1, n + 1), partitions).map(f).reduce(add)
        pi_val = 4.0 * count / n
        print("Pi val is :{}".format(pi_val))
        out.set(pi_val)

This is a pretty simple job. It calculates the value of Pi and sets it as the output for the task. The task takes in `spark_conf` and `hadoop_conf` where the user can provide any custom spark and hadoop config required for the Spark Task.

`Flyte Workflow Demo - Spark Workflow`_ is a Spark Workflow example which uses the Spark Task defined above.

.. _SparkOperator: https://github.com/GoogleCloudPlatform/spark-on-k8s-operator
.. _Dockerfile: https://github.com/apache/spark/blob/master/resource-managers/kubernetes/docker/src/main/dockerfiles/spark/Dockerfile#L22
.. _Spark Documentation: https://spark.apache.org/docs/latest/running-on-kubernetes.html
.. _Flytekit: https://github.com/lyft/flytekit/tree/master/scripts
.. _Flyte Workflow Demo - Spark Workflow: https://github.com/lyft/flytekit/blob/master/tests/flytekit/common/workflows/spark.py
