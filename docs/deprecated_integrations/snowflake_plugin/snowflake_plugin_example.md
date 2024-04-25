# Querying data in Snowflake

This example shows how to use the `SnowflakeTask` to execute a query in Snowflake.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the required libraries:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/snowflake_plugin/snowflake_plugin/snowflake_plugin_example.py
:caption: snowflake_plugin/snowflake_plugin_example.py
:lines: 1-2
```

Instantiate a {py:class}`~flytekitplugins.snowflake.SnowflakeTask` to execute a query.
Incorporate {py:class}`~flytekitplugins.snowflake.SnowflakeConfig` within the task to define the appropriate configuration.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/snowflake_plugin/snowflake_plugin/snowflake_plugin_example.py
:caption: snowflake_plugin/snowflake_plugin_example.py
:lines: 4-15
```

:::{note}
For successful registration, ensure that your Snowflake task is assigned a unique
name within your project/domain for your Flyte installation.
:::

In practical applications, our primary focus is often on utilizing Snowflake to query datasets.
Here, we employ the `SNOWFLAKE_SAMPLE_DATA`, a default dataset in the Snowflake service.
You can find more details about it [here](https://docs.snowflake.com/en/user-guide/sample-data.html).
The data adheres to the following schema:

```{eval-rst}
+----------------------------------------------+
| C_CUSTKEY (int)                              |
+----------------------------------------------+
| C_NAME (string)                              |
+----------------------------------------------+
| C_ADDRESS (string)                           |
+----------------------------------------------+
| C_NATIONKEY (int)                            |
+----------------------------------------------+
| C_PHONE (string)                             |
+----------------------------------------------+
| C_ACCTBAL (float)                            |
+----------------------------------------------+
| C_MKTSEGMENT (string)                        |
+----------------------------------------------+
| C_COMMENT (string)                           |
+----------------------------------------------+
```

Let us explore how we can parameterize our query to filter results for a specific country.
This country will be provided as user input, using a nation key to specify it.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/snowflake_plugin/snowflake_plugin/snowflake_plugin_example.py
:caption: snowflake_plugin/snowflake_plugin_example.py
:lines: 18-34
```

To review the query results, access the Snowflake console at:
`https://<SNOWFLAKE_ACCOUNT_ID>.snowflakecomputing.com/console#/monitoring/queries/detail`.

You can also execute the task and workflow locally.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/snowflake_plugin/snowflake_plugin/snowflake_plugin_example.py
:caption: snowflake_plugin/snowflake_plugin_example.py
:lines: 41-43
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/snowflake_plugin
