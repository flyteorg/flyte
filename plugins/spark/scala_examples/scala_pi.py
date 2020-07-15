from __future__ import absolute_import
from __future__ import division
from __future__ import print_function

from flytekit.sdk.tasks import generic_spark_task
from flytekit.sdk.spark_types import SparkType
from flytekit.sdk.workflow import workflow_class

# TODO: Add example with Input support.
scala_spark = generic_spark_task(
    spark_type=SparkType.SCALA,
#    inputs=inputs(partitions=Types.Integer),
    main_class="org.apache.spark.examples.SparkPi",
    main_application_file="local:///opt/spark/examples/jars/spark-examples_2.11-2.4.3.jar",
    spark_conf={
        'spark.driver.memory': "1000M",
        'spark.executor.memory': "1000M",
        'spark.executor.cores': '1',
        'spark.executor.instances': '2',
        'spark.driver.cores': '1',
            },
    cache_version='1'
)

@workflow_class
class SparkTasksWorkflow(object):
    spark_task = scala_spark()
