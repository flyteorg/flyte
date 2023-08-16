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

+++ {"lines_to_next_cell": 0}

# dbt example

In this example we're going to execute dbt commands supported by Flyte: `dbt run`,
`dbt test`, and `dbt source freshness`.

First, let's import what we need from `flytekit` and `flytekitplugins-dbt`.

```{code-cell}
import subprocess
from typing import Tuple

from flytekit import task, workflow
from flytekitplugins.dbt.schema import (
    DBTFreshnessInput,
    DBTFreshnessOutput,
    DBTRunInput,
    DBTRunOutput,
    DBTTestInput,
    DBTTestOutput,
)
from flytekitplugins.dbt.task import DBTFreshness, DBTRun, DBTTest
```

+++ {"lines_to_next_cell": 0}

We're going to use the well-known jaffle shop example, which can be found
[here](https://github.com/dbt-labs/jaffle_shop).

```{code-cell}
:lines_to_next_cell: 1

DBT_PROJECT_DIR = "jaffle_shop"
DBT_PROFILES_DIR = "dbt-profiles"
DBT_PROFILE = "jaffle_shop"
```

+++ {"lines_to_next_cell": 0}

This task ensures that the jaffle_shop database is created and it also contains
some data before scheduling an execution of this workflow.

```{code-cell}
@task
def prepare_and_seed_database():
    # Ensure the jaffle_shop database is created
    subprocess.run(
        [
            "psql",
            "-h",
            "sandbox-postgresql.flyte.svc.cluster.local",
            "-p",
            "5432",
            "-U",
            "postgres",
            "-c",
            "CREATE DATABASE jaffle_shop;",
        ],
        env={"PGPASSWORD": "postgres"},
    )
    # Seed the database with some data
    subprocess.run(
        [
            "dbt",
            "seed",
            "--project-dir",
            DBT_PROJECT_DIR,
            "--profiles-dir",
            DBT_PROFILES_DIR,
        ]
    )
```

Define the dbt tasks, in this particular case, we're going to execute a DAG containing 3 tasks:

1. [dbt run](https://docs.getdbt.com/reference/commands/run)
2. [dbt test](https://docs.getdbt.com/reference/commands/test)
3. [dbt source freshness](https://docs.getdbt.com/reference/commands/source)

```{code-cell}
dbt_run_task = DBTRun(name="example-run-task")
dbt_test_task = DBTTest(name="example-test-task")
dbt_freshness_task = DBTFreshness(name="example-freshness-task")
```

+++ {"lines_to_next_cell": 0}

Define a workflow to run the dbt tasks.

```{code-cell}
:lines_to_next_cell: 2

@workflow
def wf() -> Tuple[DBTRunOutput, DBTTestOutput, DBTFreshnessOutput]:
    dbt_run_output = dbt_run_task(
        input=DBTRunInput(
            project_dir=DBT_PROJECT_DIR,
            profiles_dir=DBT_PROFILES_DIR,
            profile=DBT_PROFILE,
        )
    )
    dbt_test_output = dbt_test_task(
        input=DBTTestInput(
            project_dir=DBT_PROJECT_DIR,
            profiles_dir=DBT_PROFILES_DIR,
            profile=DBT_PROFILE,
        )
    )
    dbt_freshness_output = dbt_freshness_task(
        input=DBTFreshnessInput(
            project_dir=DBT_PROJECT_DIR,
            profiles_dir=DBT_PROFILES_DIR,
            profile=DBT_PROFILE,
        )
    )

    # Ensure the order of the tasks.
    prepare_and_seed_database() >> dbt_run_output
    dbt_run_output >> dbt_test_output
    dbt_test_output >> dbt_freshness_output

    return dbt_run_output, dbt_test_output, dbt_freshness_output
```

To run this example workflow, follow the instructions in the
{ref}`dbt integrations page <dbt-integration>`.
