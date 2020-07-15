# Simple PySpark Job as a Flyte Workflow

[Example: pyspark_pi.py](pyspark_pi.py)

## Introduction:

This example shows how multiple different tasks (spark and python) can be built using the same docker image. We run a simple PySpark job to calculate the value of Pi and sets it as the output of the task.
This output is then consumed by a dependent python task which forms the other task in the workflow.

For PySpark Tasks, currently Flyte provides an additional `spark_context` input in addition to the `workflow_parameters`. If needed, users can convert `spark_context` to a `spark_session` like:

<code> 
sql_context = SQLContext(spark_context)

spark_session = sql_context.sparkSession
</code>
 

In future, we are thinking of updating the Spark task to provide `spark_session` directly in-place of the deprecated `spark_context`.
