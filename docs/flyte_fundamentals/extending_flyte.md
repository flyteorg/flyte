---
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

(getting_started_extending_flyte)=

# Extending Flyte

Once you have a hang of the fundamentals of Flyte, you may find that you want
to use it in ways that aren't supported out-of-the-box. Fortunately, Flyte
provides multiple extension points that enable you make it more powerful
for your specific use cases.

This guide will walk you through two of the many different ways you can extend
the Flyte type system and Flyte tasks.

(customizing_flyte_types)=

## Customizing Flyte Types

Flyte has a {ref}`rich type system <data_types_and_io>` that automatically
handles the serialization and deserialization of objects so that when you pass
data from one task to the next, you don't have to write a bunch of boilerplate
code.

However, the types that ship with Flyte or one of Flyte's
{ref}`first-party integrations <integrations>` may not fulfill your needs. In
this case, you'll need to create your own.

The easiest way to do this is with the {py:mod}`dataclasses` module, which
lets you compose several Flyte-supported types into a single object. For
example, suppose you want to support a coordinates data type with arbitrary
metadata:

```{code-cell} ipython3
import typing

from dataclasses import dataclass
from mashumaro.mixins.json import DataClassJSONMixin

@dataclass
class Coordinate(DataClassJSONMixin):
    """A custom type for coordinates with metadata attached."""
    x: float
    y: float
    metadata: typing.Dict[str, float]
```

You can then use this as a new type in your tasks and workflows:

```{code-cell} ipython3
from flytekit import task

@task
def generate_coordinates(num: int) -> typing.List[Coordinate]:
    """Generate some coordinates."""
    ...

@task
def subset_coordinates(
    coordinates: typing.List[Coordinate], x_min: float, x_max: float,
) -> typing.List[Coordinate]:
    """Select coordinates within a certain x-axis range."""
    ...
```

```{important}
The limitation of using the approach above is that you can only compose types
that are already supported by Flyte.

To create entirely new types, you'll need to use the
{py:class}`~flytekit.extend.TypeTransformer` interface to explicitly handle
the way in which the object is (a) serialized as a task output and (b)
deserialized when passed into a task as an input.

See the {ref}`User Guide <advanced_custom_types>` for an example of a custom
type.
```

## Customizing Flyte Tasks

The easiest way to extend Flyte tasks is to use Python decorators. Since Flyte
tasks are simply functions, you can wrap the task function in a custom
decorator _before_ wrapping the entire function in the `@task` decorator.

For example, if we want to do something before and after the actual task function
is invoked, we can do the following:

```{code-cell} ipython3
from functools import partial, wraps

def decorator(fn):

    @wraps(fn)
    def wrapper(*args, **kwargs):
        print("do something before")
        out = fn(*args, **kwargs)
        print("do something after")
        return out

    return wrapper
```

Then, making sure `@task` is the outermost decorator, we can modify the
behavior of the task:

```{code-cell} ipython3
@task
@decorator
def add_one(x: int) -> int:
    return x + 1

add_one(x=10)
```

This approach allows you to call out to some external service or library before
and after your task function body is executed. For example, this pattern is used
by the MLFlow integration via the {py:func}`~flytekitplugins.mlflow.mlflow_autolog`
decorator to auto-log metrics during a model-training task.

```{note}
You can stack multiple decorators on top of each other. Learn more in the
{ref}`User Guide <stacking_decorators>`.

Flyte also supports a setup-teardown pattern at the workflow level, which
allows you to enable/disable services at the beginning/end of your workflows.
See the {ref}`User Guide <decorating_workflows>` for more details.
```

## The Plugin Hierarchy of Needs

The decorator approach is great for many surface-level use cases, but there are
many more ways to customize Flyte tasks:

```{list-table}
:header-rows: 0
:widths: 10 30

* - {ref}`Pre-built Container Task Plugins <prebuilt_container>`
  - Task extensions that use pre-built containers, useful for tasks that don't
    require user-defined code and simply rely on input parameters.
* - {ref}`User Container Task Plugins <user_container>`
  - Task extensions that require user-built containers when the task also
    requires user-defined code.
* - {ref}`Raw Container Tasks <raw_container>`
  - These tasks can be implemented in other programming languages like R,
    Julia, etc. Useful for leveraging highly optimized domain-specific libraries
    in other languages outside of the `flytekit` SDK language.
* - {ref}`Backend Plugins <extend-plugin-flyte-backend>`
  - These tasks plugins require implementing a backend plugin to leverage
    external services like Sagemaker, Snowflake, BigQuery, etc.
```

## What's Next?

Congratulations! 🎉 You've just completed the Flyte Fundamentals tour.

The final section in the getting started section of the docs will provide you
with some {ref}`core use cases <getting_started_core_use_cases>` for implementing
your first workflows, whether you're a data scientist, data analyst, data engineer,
or machine learning engineer.
