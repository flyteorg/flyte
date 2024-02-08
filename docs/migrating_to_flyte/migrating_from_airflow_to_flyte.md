(migrating_from_airflow_to_flyte)=

# Migrating from Airflow to Flyte

Flyte can compile Airflow tasks into Flyte tasks without changing code, which allows you 
to migrate your Airflow DAGs to Flyte with minimal effort. This guide will walk you through
the process of migrating Airflow to Flyte.

## Prerequisites

- Install `flytekitplugins-airflow` in your python environment.
- Deploy an Airflow agent to your flyte cluster.

## Use Airflow tasks inside Flyte workflow
flytekit compiles Airflow tasks into Flyte tasks under the hood, so you can use
any Airflow sensor or operator inside a Flyte workflow.


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

## Run your Airflow tasks locally
Although Airflow doesn't support local execution, you can run your Airflow tasks locally using Flyte.

```bash
pyflyte run workflows.py airflow_wf
```

:::{warning}
Some Airflow operators may require certain permissions to execute. For instance, `DataprocCreateClusterOperator` requires the `dataproc.clusters.create` permission. 
When running Airflow tasks locally, you may need to set up the necessary permissions locally for the task to execute successfully.
:::

## Move to production
Airflow workflows can be executed on a Flyte cluster using the `--remote` flag.
In this case, Flyte creates a pod in the Kubernetes cluster to run `say_hello` task, and then run
your Airflow `BashOperator` on the Airflow agent.

```bash
pyflyte run --remote workflows.py airflow_wf
```

## Configure Airflow connection
In the local execution, you can configure the [Airflow connection](https://airflow.apache.org/docs/apache-airflow/stable/howto/connection.html) by setting the `AIRFLOW_CONN_{CONN_ID}` environment variable.
For example,
```bash
export AIRFLOW_CONN_MY_PROD_DATABASE='my-conn-type://login:password@host:port/schema?param1=val1&param2=val2'
```

In production, we recommend storing connections in a [secret Backend](https://airflow.apache.org/docs/apache-airflow/stable/security/secrets/secrets-backend/index.html).
Make sure agent pod has the right permission (IAM role) to access the secret from external secrets backends.