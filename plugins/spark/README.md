# Spark Jobs as Flyte Tasks/Workflows

This section provides examples of how to author spark tasks and workflows using FlyteKit as well as additional setup required to run Spark Jobs via Flyte.

## Setup
1. For Spark, the image must contain spark dependencies as well as the correct entrypoint for the Spark driver/executors. This can be achieved by using
the [flytekit_install_spark.sh](https://github.com/lyft/flytekit/blob/master/scripts/flytekit_install_spark.sh) script provided as referenced in the Dockerfile included here.
2. In-addition, Flyte uses the SparkOperator to run Spark Jobs as well as separate `K8s Service Account/Role` per namespace. All of these are created as part of the standard Flyte deploy.
Please refer to [Getting Started guide](https://lyft.github.io/flyte/administrator/install/getting_started.html)for more details on how to deploy Flyte.

3. Based on the resources required for your spark job (across driver/executors), you might have to tweak `resourcequotas` for the namespace.


## Spark Example Workflows
Flyte supports both `python` and `scala/java` spark tasks:

1. [PySpark Workflow](pyspark_examples/README.md)
2. [Scala/Java Spark Workflow](scala_examples/README.md)
