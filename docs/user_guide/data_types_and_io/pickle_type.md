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
    jupytext_version: 1.16.1
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

+++ {"lines_to_next_cell": 0}

(pickle_type)=

# Pickle type

```{eval-rst}
.. tags:: Basic
```

Flyte enforces type safety by utilizing type information for compiling tasks and workflows,
enabling various features such as static analysis and conditional branching.

However, we also strive to offer flexibility to end-users so they don't have to invest heavily
in understanding their data structures upfront before experiencing the value Flyte has to offer.

Flyte supports the `FlytePickle` transformer, which converts any unrecognized type hint into `FlytePickle`,
enabling the serialization/deserialization of Python values to/from a pickle file.

:::{important}
Pickle can only be used to send objects between the exact same Python version.
For optimal performance, it's advisable to either employ Python types that are supported by Flyte
or register a custom transformer, as using pickle types can result in lower performance.
:::

This example demonstrates how you can utilize custom objects without registering a transformer.

```{code-cell}
from flytekit import task, workflow
```

+++ {"lines_to_next_cell": 0}

`Superhero` represents a user-defined complex type that can be serialized to a pickle file by Flytekit
and transferred between tasks as both input and output data.

:::{note}
Alternatively, you can {ref}`turn this object into a dataclass <dataclass>` for improved performance.
We have used a simple object here for demonstration purposes.
:::

```{code-cell}
class Superhero:
    def __init__(self, name, power):
        self.name = name
        self.power = power


@task
def welcome_superhero(name: str, power: str) -> Superhero:
    return Superhero(name, power)


@task
def greet_superhero(superhero: Superhero) -> str:
    return f"👋 Hello {superhero.name}! Your superpower is {superhero.power}."


@workflow
def superhero_wf(name: str = "Thor", power: str = "Flight") -> str:
    superhero = welcome_superhero(name=name, power=power)
    return greet_superhero(superhero=superhero)
```

+++ {"lines_to_next_cell": 0}

## Batch size

By default, if the list subtype is unrecognized, a single pickle file is generated.
To optimize serialization and deserialization performance for scenarios involving a large number of items
or significant list elements, you can specify a batch size.
This feature allows for the processing of each batch as a separate pickle file.
The following example demonstrates how to set the batch size.

```{code-cell}
from typing import Iterator

from flytekit.types.pickle.pickle import BatchSize
from typing_extensions import Annotated


@task
def welcome_superheroes(names: list[str], powers: list[str]) -> Annotated[list[Superhero], BatchSize(3)]:
    return [Superhero(name, power) for name, power in zip(names, powers)]


@task
def greet_superheroes(superheroes: list[Superhero]) -> Iterator[str]:
    for superhero in superheroes:
        yield f"👋 Hello {superhero.name}! Your superpower is {superhero.power}."


@workflow
def superheroes_wf(
    names: list[str] = ["Thor", "Spiderman", "Hulk"],
    powers: list[str] = ["Flight", "Surface clinger", "Shapeshifting"],
) -> Iterator[str]:
    superheroes = welcome_superheroes(names=names, powers=powers)
    return greet_superheroes(superheroes=superheroes)
```

+++ {"lines_to_next_cell": 0}

:::{note}
The `welcome_superheroes` task will generate two pickle files: one containing two superheroes and the other containing one superhero.
:::

You can run the workflows locally as follows:

```{code-cell}
if __name__ == "__main__":
    print(f"Superhero wf: {superhero_wf()}")
    print(f"Superhero(es) wf: {superheroes_wf()}")
```
