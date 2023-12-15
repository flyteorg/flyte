---
jupytext:
  formats: md:myst
  text_representation:
    extension: .md
    format_name: myst
kernelspec:
  display_name: Python 3
  language: python
  name: python3

# override the toc-determined page navigation order
next-page: getting_started/flyte_fundamentals
next-page-title: Flyte Fundamentals
---

(getting_started_index)=

# Getting Started

## Introduction to Flyte

Flyte is a workflow orchestrator that seamlessly unifies data,
machine learning, and analytics stacks for building robust and reliable
applications.

This introduction provides a quick overview of how to get Flyte up and running
on your local machine.

````{dropdown} Want to try Flyte on the browser?
:title: text-muted
:animate: fade-in-slide-down

The introduction below is also available on a hosted sandbox environment, where
you can get started with Flyte without installing anything locally.

```{link-button} https://sandbox.union.ai/
---
classes: try-hosted-flyte btn-warning btn-block
text: Try Hosted Flyte Sandbox
---
```

```{div} text-muted
*Courtesy of [Union.ai](https://www.union.ai/)*
```

````

(getting_started_installation)=

## Installation

```{admonition} Prerequisites
:class: important

[Install Docker](https://docs.docker.com/get-docker/) and ensure that you
have the Docker daemon running.

Flyte supports any [OCI-compatible](https://opencontainers.org/) container
technology (like [Podman](https://podman.io/),
[LXD](https://linuxcontainers.org/lxd/introduction/), and
[Containerd](https://containerd.io/)) when running tasks on a Flyte cluster, but
for the purpose of this guide, `flytectl` uses Docker to spin up a local
Kubernetes cluster so that you can interact with it on your machine.
```

First install [flytekit](https://pypi.org/project/flytekit/), Flyte's Python SDK and [Scikit-learn](https://scikit-learn.org/stable).

```{prompt} bash $
pip install flytekit flytekitplugins-deck-standard scikit-learn
```

Then install [flytectl](https://docs.flyte.org/projects/flytectl/en/latest/),
which the command-line interface for interacting with a Flyte backend.

````{tabbed} Homebrew

```{prompt} bash $
brew install flyteorg/homebrew-tap/flytectl
```

````

````{tabbed} Curl

```{prompt} bash $
curl -sL https://ctl.flyte.org/install | sudo bash -s -- -b /usr/local/bin
```

````

## Creating a Workflow

The first workflow we'll create is a simple model training workflow that consists
of three steps that will:

1. 🍷 Get the classic [wine dataset](https://scikit-learn.org/stable/datasets/toy_dataset.html#wine-recognition-dataset)
   using [sklearn](https://scikit-learn.org/stable/).
2. 📊 Process the data that simplifies the 3-class prediction problem into a
   binary classification problem by consolidating class labels `1` and `2` into
   a single class.
3. 🤖 Train a `LogisticRegression` model to learn a binary classifier.

First, we'll define three tasks for each of these steps. Create a file called
`example.py` and copy the following code into it.

```{code-cell} python
:tags: [remove-output]

import pandas as pd
from sklearn.datasets import load_wine
from sklearn.linear_model import LogisticRegression

import flytekit.extras.sklearn
from flytekit import task, workflow


@task
def get_data() -> pd.DataFrame:
    """Get the wine dataset."""
    return load_wine(as_frame=True).frame

@task
def process_data(data: pd.DataFrame) -> pd.DataFrame:
    """Simplify the task from a 3-class to a binary classification problem."""
    return data.assign(target=lambda x: x["target"].where(x["target"] == 0, 1))

@task
def train_model(data: pd.DataFrame, hyperparameters: dict) -> LogisticRegression:
    """Train a model on the wine dataset."""
    features = data.drop("target", axis="columns")
    target = data["target"]
    return LogisticRegression(max_iter=3000, **hyperparameters).fit(features, target)
```

As we can see in the code snippet above, we defined three tasks as Python
functions: `get_data`, `process_data`, and `train_model`.

In Flyte, **tasks** are the most basic unit of compute and serve as the building
blocks 🧱 for more complex applications. A task is a function that takes some
inputs and produces an output. We can use these tasks to define a simple model
training workflow:

```{code-cell} python
@workflow
def training_workflow(hyperparameters: dict) -> LogisticRegression:
    """Put all of the steps together into a single workflow."""
    data = get_data()
    processed_data = process_data(data=data)
    return train_model(
        data=processed_data,
        hyperparameters=hyperparameters,
    )
```

```{note}
A task can also be an isolated piece of compute that takes no inputs and
produces no output, but for the most part to do something useful a task
is typically written with inputs and outputs.
```

A **workflow** is also defined as a Python function, and it specifies the flow
of data between tasks and, more generally, the dependencies between tasks 🔀.

::::{dropdown} {fa}`info-circle` The code above looks like Python, but what do `@task` and `@workflow` do exactly?
:title: text-muted
:animate: fade-in-slide-down

Flyte `@task` and `@workflow` decorators are designed to work seamlessly with
your code-base, provided that the *decorated function is at the top-level scope
of the module*.

This means that you can invoke tasks and workflows as regular Python methods and
even import and use them in other Python modules or scripts.

:::{note}
A {func}`task <flytekit.task>` is a pure Python function, while a {func}`workflow <flytekit.workflow>`
is actually a [DSL](https://en.wikipedia.org/wiki/Domain-specific_language) that
only supports a subset of Python's semantics. Learn more in the
{ref}`Flyte Fundamentals <workflows_versus_task_syntax>` section.
:::

::::

(intro_running_flyte_workflows)=

## Running Flyte Workflows in Python

You can run the workflow in ``example.py`` on a local Python by using `pyflyte`,
the CLI that ships with `flytekit`.

```{prompt} bash $
pyflyte run example.py training_workflow \
    --hyperparameters '{"C": 0.1}'
```

:::::{dropdown} {fa}`info-circle`  Running into shell issues?
:title: text-muted
:animate: fade-in-slide-down

If you're using Bash, you can ignore this 🙂
You may need to add .local/bin to your PATH variable if it's not already set,
as that's not automatically added for non-bourne shells like fish or xzsh.

To use pyflyte, make sure to set the /.local/bin directory in PATH

:::{code-block} fish
set -gx PATH $PATH ~/.local/bin
:::
:::::



:::::{dropdown} {fa}`info-circle` Why use `pyflyte run` rather than `python example.py`?
:title: text-muted
:animate: fade-in-slide-down

`pyflyte run` enables you to execute a specific workflow using the syntax
`pyflyte run <path/to/script.py> <workflow_or_task_function_name>`.

Keyword arguments can be supplied to ``pyflyte run`` by passing in options in
the format ``--kwarg value``, and in the case of ``snake_case_arg`` argument
names, you can pass in options in the form of ``--snake-case-arg value``.

::::{note}
If you want to run a workflow with `python example.py`, you would have to write
a `main` module conditional at the end of the script to actually run the
workflow:

:::{code-block} python
if __name__ == "__main__":
    training_workflow(hyperparameters={"C": 0.1})
:::

This becomes even more verbose if you want to pass in arguments:

:::{code-block} python
if __name__ == "__main__":
    import json
    from argparse import ArgumentParser

    parser = ArgumentParser()
    parser.add_argument("--hyperparameters", type=json.loads)
    ...  # add the other options

    args = parser.parse_args()
    training_workflow(hyperparameters=args.hyperparameters)
:::

::::

:::::

(getting_started_flyte_cluster)=

## Running Workflows in a Flyte Cluster

You can also use `pyflyte run` to execute workflows on a Flyte cluster.
To do so, first spin up a local demo cluster. `flytectl` uses Docker to create
a local Kubernetes cluster and minimal Flyte backend that you can use to run
the example above:

```{important}
Before you start the local cluster, make sure that you allocate a minimum of
`4 CPUs` and `3 GB` of memory in your Docker daemon. If you're using the
[Docker Desktop](https://www.docker.com/products/docker-desktop/), you can
do this easily by going to:

`Settings > Resources > Advanced`

Then set the **CPUs** and **Memory** sliders to the appropriate levels.
```

```{prompt} bash $
flytectl demo start
```

````{div} shadow p-3 mb-8 rounded
**Expected Output:**

```{code-block}
👨‍💻 Flyte is ready! Flyte UI is available at http://localhost:30080/console 🚀 🚀 🎉
❇️ Run the following command to export sandbox environment variables for accessing flytectl
	export FLYTECTL_CONFIG=~/.flyte/config-sandbox.yaml
🐋 Flyte sandbox ships with a Docker registry. Tag and push custom workflow images to localhost:30000
📂 The Minio API is hosted on localhost:30002. Use http://localhost:30080/minio/login for Minio console
```

```{important}
Make sure to export the `FLYTECTL_CONFIG=~/.flyte/config-sandbox.yaml` environment
variable in your shell.
```
````

Then, run the workflow on the Flyte cluster with `pyflyte run` using the
`--remote` flag:

```{prompt} bash $
pyflyte run --remote example.py training_workflow \
    --hyperparameters '{"C": 0.1}'
```

````{div} shadow p-3 mb-8 rounded

**Expected Output:** A URL to the workflow execution on your demo Flyte cluster:

```{code-block}
Go to http://localhost:30080/console/projects/flytesnacks/domains/development/executions/<execution_name> to see execution in the console.
```

Where ``<execution_name>`` is a unique identifier for the workflow execution.

````


## Inspect the Results

Navigate to the URL produced by `pyflyte run`. This will take you to
FlyteConsole, the web UI used to manage Flyte entities such as tasks,
workflows, and executions.

![getting started console](https://github.com/flyteorg/static-resources/raw/main/flytesnacks/getting_started/getting_started_console.gif)


```{note}
There are a few features about FlyteConsole worth pointing out in the GIF above:

- The default execution view shows the list of tasks executing in sequential order.
- The right-hand panel shows metadata about the task execution, including logs, inputs, outputs, and task metadata.
- The **Graph** view shows the execution graph of the workflow, providing visual information about the topology
  of the graph and the state of each node as the workflow progresses.
- On completion, you can inspect the outputs of each task, and ultimately, the overarching workflow.
```

## Summary

🎉  **Congratulations! In this introductory guide, you:**

1. 📜 Created a Flyte script, which trains a binary classification model.
2. 🚀 Spun up a demo Flyte cluster on your local system.
3. 👟 Ran a workflow locally and on a demo Flyte cluster.


## What's Next?

Follow the rest of the sections in the documentation to get a better
understanding of the key constructs that make Flyte such a powerful
orchestration tool 💪.

```{admonition} Recommendation
:class: tip

If you're new to Flyte we recommend that you go through the
{ref}`Flyte Fundamentals <getting_started_fundamentals>` and
{ref}`Core Use Cases <getting_started_core_use_cases>` section before diving
into the other sections of the documentation.
```

```{list-table}
:header-rows: 0
:widths: 10 30

* - {ref}`🔤 Flyte Fundamentals <getting_started_fundamentals>`
  - A brief tour of the Flyte's main concepts and development lifecycle
* - {ref}`🌟 Core Use Cases <getting_started_core_use_cases>`
  - An overview of core uses cases for data, machine learning, and analytics
    practitioners.
* - {ref}`📖 User Guide <userguide>`
  - A comprehensive view of Flyte's functionality for data scientists,
    ML engineers, data engineers, and data analysts.
* - {ref}`📚 Tutorials <tutorials>`
  - End-to-end examples of Flyte for data/feature engineering, machine learning,
    bioinformatics, and more.
* - {ref}`🚀 Deployment Guide <deployment>`
  - Guides for platform engineers to deploy a Flyte cluster on your own
    infrastructure.
```
