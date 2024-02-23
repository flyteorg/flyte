---
jupytext:
  formats: md:myst
  text_representation:
    extension: .md
    format_name: myst
---

(testing_agents_locally)=
# Testing agents locally

You can test agents locally without running the backend server, making agent development easier.

To test an agent locally, create a class for the agent task that inherits from [AsyncAgentExecutorMixin](https://github.com/flyteorg/flytekit/blob/master/flytekit/extend/backend/base_agent.py#L155). This mixin can handle both asynchronous tasks and synchronous tasks and allows flytekit to mimic FlytePropeller's behavior in calling the agent.

## BigQuery example

To test the BigQuery example, copy the following code to a file called `wf.py`, modifying as needed.

```{note}

In some cases, you will need to store credentials in your local environment when testing locally.
For example, you need to set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable when running BigQuery tasks to test the BigQuery agent.

```

```python
class BigQueryTask(AsyncAgentExecutorMixin, SQLTask[BigQueryConfig]):
    def __init__(self, name: str, **kwargs):
        ...


# Instantiate the task class. Flytekit will automatically call the agent
# to `create`, `get`, or `delete` the job.
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
