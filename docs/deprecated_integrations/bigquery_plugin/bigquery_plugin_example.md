(bigquery_plugin_example)=
# BigQuery example query

This example shows how to use a Flyte BigQueryTask to execute a query.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/bigquery_plugin/bigquery_plugin/bigquery_plugin_example.py
:caption: bigquery_plugin/bigquery_plugin_example.py
:lines: 1-8
```

This is the world's simplest query. Note that in order for registration to work properly, you'll need to give your BigQuery task a name that's unique across your project/domain for your Flyte installation.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/bigquery_plugin/bigquery_plugin/bigquery_plugin_example.py
:caption: bigquery_plugin/bigquery_plugin_example.py
:lines: 12-22
```

Of course, in real world applications we are usually more interested in using BigQuery to query a dataset.
In this case we use crypto_dogecoin data which is public dataset in BigQuery.
[here](https://console.cloud.google.com/bigquery?project=bigquery-public-data&page=table&d=crypto_dogecoin&p=bigquery-public-data&t=transactions)

Let's look out how we can parameterize our query to filter results for a specific transaction version, provided as a user input
specifying a version.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/bigquery_plugin/bigquery_plugin/bigquery_plugin_example.py
:caption: bigquery_plugin/bigquery_plugin_example.py
:lines: 25-34
```

StructuredDataset transformer can convert query result to pandas dataframe here.
We can also change "pandas.dataframe" to "pyarrow.Table", and convert result to Arrow table.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/bigquery_plugin/bigquery_plugin/bigquery_plugin_example.py
:caption: bigquery_plugin/bigquery_plugin_example.py
:lines: 37-45
```

Check query result on bigquery console: `https://console.cloud.google.com/bigquery`

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/bigquery_plugin