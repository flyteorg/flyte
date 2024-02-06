(getting_started_quickstart_guide)=
# Quickstart guide

In this guide, you will create and run a Flyte workflow in a local Python environment to generate the output "Hello, world!"

## Prerequisites

* [Install Python 3.8x or higher](https://www.python.org/downloads/)
* Install [Flytekit](https://github.com/flyteorg/flytekit) with `pip install -U flytekit`

## Steps

### 1. Create a "Hello, world!" file

To create an example workflow file, copy the following into a file called `example.py`:

```python
from flytekit import task, workflow


@task
def say_hello(name: str) -> str:
    return f"Hello, {name}!"


@workflow
def hello_world_wf(name: str = 'world') -> str:
    res = say_hello(name=name)
    return res


if __name__ == "__main__":
    print(f"Running wf() {hello_world_wf(name='passengers')}")
```

:::{note}
You can also use the `pyflyte init` command to initialize the "Hello, world!" Flyte project by running the following command:

```{prompt} bash $
pyflyte init --template hello-world hello-world
```

This will create a project directory that contains an `example.py` file with code above.
:::

### 2. Run the example workflow in a local Python environment

Next, run the workflow in the example workflow file with `pyflyte run`. The initial arguments of `pyflyte run` take the form of
`path/to/script.py <task_or_workflow_name>`, where `<task_or_workflow_name>`
refers to the function decorated with `@task` or `@workflow` that you wish to run:

```{prompt} bash $
pyflyte run example.py hello_world_wf
```

You can also provide a `name` argument to the workflow:
```{prompt} bash $
pyflyte run example.py hello_world_wf --name Ada
```

:::{note}
If you created a "Hello, world" project using `pyflyte init`, you will need to change directories before running the workflow:
```{prompt} bash $
cd hello-world
pyflyte run example.py hello_world_wf
```
:::

## The @task and @workflow decorators

In this example, the file `example.py` contains a task and a workflow, decorated with the `@task` and `@workflow` decorators, respectively. You can invoke tasks and workflows like regular Python methods, and even import and use them in other Python modules or scripts.

```python
@task
def say_hello(name: str) -> str:
    return f"Hello, {name}!"


@workflow
def hello_world_wf(name: str = 'world') -> str:
    res = say_hello(name=name)
    return res
```

To learn more about tasks and workflows, see the {ref}`"Workflow code" section<getting_started_workflow_code>` of {doc}`"Flyte project components"<getting_started_with_workflow_development/flyte_project_components>`.

## Next steps

To create a productionizable Flyte project to structure your code according to software engineering best practices, and that can be used to package your code for deployment to a Flyte cluster, see {doc}`"Getting started with workflow development" <getting_started_with_workflow_development/index>`.
