(migrating_from_airflow_to_flyte)=

# Migrating from Airflow to Flyte

Flyte can compile Airflow tasks into Flyte tasks without changing code, which allows you 
to migrate your Airflow DAGs to Flyte with minimal effort.

In addition to migration capabilities, Flyte users can seamlessly integrate Airflow tasks into their workflows, leveraging the ecosystem of Airflow operators and sensors.
By combining the robust Airflow ecosystem with Flyte's capabilities such as caching, versioning, and reproducibility, users can run more complex data and machine learning workflows with ease.

## Prerequisites

- Install `flytekitplugins-airflow` in your Python environment.
- Enable an {ref}`Airflow agent<deployment-agent-setup-airflow>` in your Flyte cluster.

## Steps

### 1. Define your Airflow tasks in a Flyte workflow

Flytekit compiles Airflow tasks into Flyte tasks, so you can use
any Airflow sensor or operator in a Flyte workflow.


```python
from flytekit import task, workflow
from airflow.operators.bash import BashOperator

@task
def say_hello() -> str:
    return "Hello, World!"

@workflow
def airflow_wf():
    flyte_task = say_hello()
    airflow_task = BashOperator(task_id=f"airflow_bash_operator", bash_command="echo hello")
    airflow_task >> flyte_task

if __name__ == "__main__":
    print(f"Running airflow_wf() {airflow_wf()}")
```

### 2. Test your workflow locally

:::{note}
Before running your workflow locally, you must configure the [Airflow connection](https://airflow.apache.org/docs/apache-airflow/stable/howto/connection.html) by setting the `AIRFLOW_CONN_{CONN_ID}` environment variable.
For example,
```bash
export AIRFLOW_CONN_MY_PROD_DATABASE='my-conn-type://login:password@host:port/schema?param1=val1&param2=val2'
```
:::

Although Airflow doesn't support local execution, you can run your workflow that contains Airflow tasks locally, which is helpful for testing and debugging your tasks before moving to production.

```bash
pyflyte run workflows.py airflow_wf
```

:::{warning}
Some Airflow operators may require certain permissions to execute. For instance, `DataprocCreateClusterOperator` requires the `dataproc.clusters.create` permission. 
When running Airflow tasks locally, you may need to set the necessary permissions locally for the task to execute successfully.
:::

### 3. Move your workflow to production

:::{note}
In production, we recommend storing connections in a [secrets backend](https://airflow.apache.org/docs/apache-airflow/stable/security/secrets/secrets-backend/index.html).
Make sure the agent pod has the right permission (IAM role) to access the secret from the external secrets backend.
:::

After you have tested your workflow locally, you can execute it on a Flyte cluster using the `--remote` flag.
In this case, Flyte creates a pod in the Kubernetes cluster to run the `say_hello` task, and then runs
your Airflow `BashOperator` task on the Airflow agent.

```bash
pyflyte run --remote workflows.py airflow_wf
```

:::{note}
Many Airflow operators and sensors have been tested on Flyte, but some may not work as expected.
If you encounter any issues, please file an [issue](https://github.com/flyteorg/flyte/issues) or reach out to the Flyte community on [Slack](https://slack.flyte.org/).
:::