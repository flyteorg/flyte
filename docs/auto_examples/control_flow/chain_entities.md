---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.14.7
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

(chain_flyte_entities)=

# Chain Flyte Entities

```{eval-rst}
.. tags:: Basic
```

Data passing between tasks or workflows need not necessarily happen through parameters.
In such a case, if you want to explicitly construct the dependency, flytekit provides a mechanism to chain Flyte entities using the `>>` operator.

## Tasks

Let's enforce an order for `read()` to happen after `write()`, and for `write()` to happen after `create_bucket()`.

:::{note}
To run the example locally, spin up the demo cluster using `flytectl demo start`.
:::

+++ {"lines_to_next_cell": 0}

Import the necessary dependencies.

```{code-cell}
import logging
from io import StringIO

import pandas as pd
from botocore import session
from flytekit import task, workflow
from flytekit.configuration import S3Config

logger = logging.getLogger(__file__)

CSV_FILE = "iris.csv"
BUCKET_NAME = "chain-flyte-entities"


def s3_client():
    cfg = S3Config.auto()
    sess = session.get_session()
    return sess.create_client(
        "s3",
        aws_access_key_id=cfg.access_key_id,
        aws_secret_access_key=cfg.secret_access_key,
        use_ssl=False,
        endpoint_url=cfg.endpoint,
    )
```

+++ {"lines_to_next_cell": 0}

Create an s3 bucket.
This task exists just for the sandbox case.

```{code-cell}
@task(cache=True, cache_version="1.0")
def create_bucket():
    client = s3_client()
    try:
        client.create_bucket(Bucket=BUCKET_NAME)
    except client.exceptions.BucketAlreadyOwnedByYou:
        logger.info(f"Bucket {BUCKET_NAME} has already been created by you.")
```

+++ {"lines_to_next_cell": 0}

Define a `read()` task to read from the s3 bucket.

```{code-cell}
@task
def read() -> pd.DataFrame:
    data = pd.read_csv(s3_client().get_object(Bucket=BUCKET_NAME, Key=CSV_FILE)["Body"])
    return data
```

+++ {"lines_to_next_cell": 0}

Define a `write()` task to write the dataframe to a CSV file in the s3 bucket.

```{code-cell}
@task(cache=True, cache_version="1.0")
def write():
    df = pd.DataFrame(  # noqa : F841
        data={
            "sepal_length": [5.3],
            "sepal_width": [3.8],
            "petal_length": [0.1],
            "petal_width": [0.3],
            "species": ["setosa"],
        }
    )
    csv_buffer = StringIO()
    df.to_csv(csv_buffer)
    s3_client().put_object(Body=csv_buffer.getvalue(), Bucket=BUCKET_NAME, Key=CSV_FILE)
```

+++ {"lines_to_next_cell": 0}

We want to enforce an order here: `create_bucket()` followed by `write()` followed by `read()`.
Since no data-passing happens between the tasks, use `>>` operator on the tasks.

```{code-cell}
@workflow
def chain_tasks_wf() -> pd.DataFrame:
    create_bucket_promise = create_bucket()
    write_promise = write()
    read_promise = read()

    create_bucket_promise >> write_promise
    write_promise >> read_promise

    return read_promise
```

+++ {"lines_to_next_cell": 0}

## Chain SubWorkflows

Similar to tasks, you can chain {ref}`subworkflows <subworkflows>`.

```{code-cell}
@workflow
def write_sub_workflow():
    write()


@workflow
def read_sub_workflow() -> pd.DataFrame:
    return read()
```

+++ {"lines_to_next_cell": 0}

Use `>>` to chain the subworkflows.

```{code-cell}
@workflow
def chain_workflows_wf() -> pd.DataFrame:
    create_bucket_promise = create_bucket()
    write_sub_wf = write_sub_workflow()
    read_sub_wf = read_sub_workflow()

    create_bucket_promise >> write_sub_wf
    write_sub_wf >> read_sub_wf

    return read_sub_wf
```

+++ {"lines_to_next_cell": 0}

Run the workflows locally.

```{code-cell}
if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(f"Running chain_tasks_wf()... {chain_tasks_wf()}")
    print(f"Running chain_workflows_wf()... {chain_workflows_wf()}")
```
