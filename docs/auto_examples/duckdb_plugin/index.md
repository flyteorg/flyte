(duckdb)=

# DuckDB

```{eval-rst}
.. tags:: Integration, Data, Analytics, Beginner
```

[DuckDB](https://duckdb.org/) is an in-process SQL OLAP database management system that is explicitly designed to achieve high performance in analytics.

The Flytekit DuckDB plugin facilitates the efficient execution of intricate analytical queries within your workflow.

To install the Flytekit DuckDB plugin, run the following command:

```
pip install flytekitplugins-duckdb
```

The Flytekit DuckDB plugin includes the {py:class}`~flytekitplugins:flytekitplugins.duckdb.DuckDBQuery` task, which allows you to specify the following parameters:

- `query`: The DuckDB query to execute.
- `inputs`: The query parameters to be used during query execution. This can be a StructuredDataset, a string or a list.

```{auto-examples-toc}
duckdb_example
```
