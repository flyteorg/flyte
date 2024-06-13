(spark_on_databricks_plugin)=

# Running Spark on Databricks

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the required dependencies:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/databricks_plugin/databricks_plugin/databricks_plugin_example.py
:caption: databricks_plugin/databricks_plugin_example.py
:lines: 1-7
```

To run a Spark job on the Databricks platform, simply include Databricks configuration in the task config.
The Databricks config is the same as the Databricks job request. For more details, please refer to the
[Databricks job request](https://docs.databricks.com/dev-tools/api/2.0/jobs.html#request-structure) documentation.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/databricks_plugin/databricks_plugin/databricks_plugin_example.py
:caption: databricks_plugin/databricks_plugin_example.py
:pyobject: hello_spark
```

For this particular example,
we define a function that executes the map-reduce operation within the Spark cluster.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/databricks_plugin/databricks_plugin/databricks_plugin_example.py
:caption: databricks_plugin/databricks_plugin_example.py
:pyobject: f
```

Additionally, we define a standard Flyte task that won't be executed on the Spark cluster.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/databricks_plugin/databricks_plugin/databricks_plugin_example.py
:caption: databricks_plugin/databricks_plugin_example.py
:pyobject: print_every_time
```

Finally, define a workflow that connects your tasks in a sequence.
Remember, Spark and non-Spark tasks can be chained together as long as their parameter specifications match.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/databricks_plugin/databricks_plugin/databricks_plugin_example.py
:caption: databricks_plugin/databricks_plugin_example.py
:pyobject: my_databricks_job
```

You can execute the workflow locally.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/databricks_plugin/databricks_plugin/databricks_plugin_example.py
:caption: databricks_plugin/databricks_plugin_example.py
:lines: 79-83
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/databricks_plugin
