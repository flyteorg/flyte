# Flyte project components

A Flyte project is a directory containing task and workflow code, internal Python source code, configuration files, and other artifacts required to package up your code so that it can be run on a Flyte cluster.

## Directory structure

If you look at the project you created with `pyflyte init` in {doc}`"Creating a Flyte project" <creating_a_flyte_project>`, you'll see the following directory structure:

```{code-block} bash
my_project
├── LICENSE
├── README.md
├── requirements.txt  # Python dependencies
└── workflows
    ├── __init__.py
    └── example.py    # Example Flyte workflow code
```

(getting_started_python_dependencies)=

## `requirements.txt` Python dependencies

Most Flyte projects contain a `requirements.txt` file that you can modify to suit the needs of your project.

You can specify pip-installable Python dependencies in your project by adding them to the
`requirements.txt` file.

```{note}
We recommend using [pip-compile](https://pip-tools.readthedocs.io/en/latest/) to
manage your project's Python requirements.
```

````{dropdown} See requirements.txt

```{rli} https://raw.githubusercontent.com/flyteorg/flytekit-python-template/main/simple-example/%7B%7Bcookiecutter.project_name%7D%7D/requirements.txt
:caption: requirements.txt
```

````

(getting_started_workflow_code)=

## `example.py` workflow code

Flyte projects initialized with `pyflyte init` contain a `workflows` directory, inside of which is a Python file that holds the workflow code for the application.

The workflow code may contain ImageSpec configurations, and one or more task and workflow functions, decorated with the `@task` and `@workflow` decorators, respectively.

```{note}
The workflow directory also contains an `__init__.py` file to indicate that the workflow code is part of a Python package. For more information, see the [Python documentation](https://docs.python.org/3/reference/import.html#regular-packages).
```

### ImageSpec

The workflow code file in the basic template includes an optional ImageSpec configuration. ImageSpec is a Flyte feature that enables you to build a custom container image without having to write a Dockerfile. To learn more, see the [ImageSpec documentation](https://docs.flyte.org/projects/cookbook/en/latest/auto_examples/customizing_dependencies/image_spec.html#image-spec-example)

```python
# basic_image = ImageSpec(
#    name="flytekit",  # rename this to your docker image name
#    base_image="ghcr.io/flyteorg/flytekit:py3.11-1.10.2",
#    # the base image that flytekit will use to build your image
#    packages=["example-package"],  # packages to add to the base image
#    # remove "example-package" before using.
#    registry="ghcr.io/unionai-oss",
#    # the registry your image will be pushed to
#    python_version="3.11"
#    # the python version; optional if not different from the base image
# )
```

```{note}
If you need to use a Dockerfile instead of ImageSpec, you will need to add a Dockerfile and a `docker_build.sh` script to the top-level directory of your project, and either remove any ImageSpec configurations from the workflow code file or leave them commented out.
```

### The `@task` and `@workflow` decorators

* The `@task` and `@workflow` decorators can be parsed by Python provided that they are used only on functions at the top-level scope of the module.
* Task and workflow function signatures must be type-annotated with Python type hints.
* Tasks and workflows can be invoked like regular Python methods, and even imported and used in other Python modules or scripts.
* Task and workflow functions must be invoked with keyword arguments.

#### `@task`

The `@task` decorator indicates a Python function that defines a task.

* A task is a Python function that takes some inputs and produces an output.
* Tasks are assembled into workflows.
* When deployed to a Flyte cluster, each task runs in its own [Kubernetes Pod](https://kubernetes.io/docs/concepts/workloads/pods/), where Flyte orchestrates what task runs at what time in the context of a workflow.

```python
@task()
def say_hello(name: str) -> str:
    return f"Hello, {name}!"
```

#### `@workflow`

The `@workflow` decorator indicates a function-esque construct that defines a workflow.

* Workflows specify the flow of data between tasks, and the dependencies between tasks.
* A workflow appears to be a Python function but is actually a [domain-specific language (DSL)](https://en.wikipedia.org/wiki/Domain-specific_language) that only supports a subset of Python syntax and semantics.
* When deployed to a Flyte cluster, the workflow function is "compiled" to construct the directed acyclic graph (DAG) of tasks, defining the order of execution of task pods and the data flow dependencies between them.

```python
@workflow
def wf(name: str = "world") -> typing.Tuple[str, int]:
    greeting = say_hello(name=name)
    greeting_len = greeting_length(greeting=greeting)
    return greeting, greeting_len
```
