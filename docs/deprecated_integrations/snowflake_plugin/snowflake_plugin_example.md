# Querying data in Snowflake

This example shows how to use the `SnowflakeTask` to execute a query in Snowflake.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/snowflake_plugin/snowflake_plugin/snowflake_plugin_example.py
:caption: snowflake_plugin/snowflake_plugin_example.py
:lines: 1-120
```

Instantiate a {py:class}`~flytekitplugins.snowflake.SnowflakeTask` to execute a query.
Incorporate {py:class}`~flytekitplugins.snowflake.SnowflakeConfig` within the task to define the appropriate configuration.

:::{note}
Please give your Snowflake task different names.
:::

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/snowflake_plugin
