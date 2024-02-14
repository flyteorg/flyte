---
jupytext:
  formats: md:myst
  text_representation:
    extension: .md
    format_name: myst
---

(testing_agents_locally)=
# Testing agents locally

Agents can be tested locally without configuring or running the backend server, which makes agent development easier.

The task inherited from `AsyncAgentExecutorMixin` can be executed locally, allowing flytekit to mimic FlytePropeller's behavior to call the agent.

:::{note}

In some cases, you will need to store credentials in your local environment when testing locally.
For example, you need to set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable when running BigQuery tasks to test the BigQuery agent.

:::

Flytekit will automatically call the agent to `create`, `get`, or `delete` the task.

```python
bigquery_doge_coin = BigQueryTask(
    name=f"bigquery.doge_coin",
    inputs=kwtypes(version=int),
    query_template="SELECT * FROM `bigquery-public-data.crypto_dogecoin.transactions` WHERE version = @version LIMIT 10;",
    output_structured_dataset_type=StructuredDataset,
    task_config=BigQueryConfig(ProjectID="flyte-test-340607")
)
```

You can run the above example task locally and test the agent with the following command:

```bash
pyflyte run wf.py bigquery_doge_coin --version 10
```
