from __future__ import absolute_import
from __future__ import division
from __future__ import print_function

import random
from operator import add

from six.moves import range

from flytekit.sdk.tasks import spark_task, inputs, outputs, python_task
from flytekit.sdk.types import Types
from flytekit.sdk.workflow import workflow_class, Input, Output


@inputs(partitions=Types.Integer)
@outputs(out=Types.Float)
@spark_task(
    spark_conf={
        'spark.driver.memory': "1000M",
        'spark.executor.memory': "1000M",
        'spark.executor.cores': '1',
        'spark.executor.instances': '2',
        'spark.driver.cores': '1',
    },
    cache_version='1')
def hello_spark(workflow_parameters, spark_context, partitions, out):
    print("Starting Spark with Partitions: {}".format(partitions))

    n = 100000 * partitions
    count = spark_context.parallelize(range(1, n + 1), partitions).map(f).reduce(add)
    pi_val = 4.0 * count / n
    print("Pi val is :{}".format(pi_val))
    out.set(pi_val)


@inputs(value_to_print=Types.Float, date_triggered=Types.Datetime)
@python_task(cache_version='1')
def print_every_time(workflow_parameters, value_to_print, date_triggered):
    print("My printed value: {} @ {}".format(value_to_print, date_triggered))

def f(_):
    x = random.random() * 2 - 1
    y = random.random() * 2 - 1
    return 1 if x ** 2 + y ** 2 <= 1 else 0


@workflow_class
class SparkTasksWorkflow(object):
    triggered_date = Input(Types.Datetime)
    spark_pi = hello_spark(partitions=50)
    print_always = print_every_time(
        value_to_print=spark_pi.outputs.out,
        date_triggered=triggered_date)
    pi = Output(spark_pi.outputs.out, sdk_type=Types.Float)