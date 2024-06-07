(advanced_custom_types)=

# Custom types

```{eval-rst}
.. tags:: Extensibility, Contribute, Intermediate
```

Flyte is a strongly-typed framework for authoring tasks and workflows. But there are situations when the existing types do not directly work. This is true with any programming language!

Similar to a programming language enabling higher-level concepts to describe user-specific objects such as classes in Python/Java/C++, struct in C/Golang, etc.,
Flytekit allows modeling user classes. The idea is to make an interface that is more productive for the
use case, while writing a transformer that converts the user-defined type into one of the generic constructs in Flyte's type system.

This example will try to model an example user-defined dataset and show how it can be seamlessly integrated with Flytekit's type engine.

The example is demonstrated in the video below:

```{eval-rst}
..  youtube:: 1xExpRzz8Tw

```

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

First, we import the dependencies:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/0ec8388759d34566a0ffc0c3c2d7443fd4a3a46f/examples/extending/extending/custom_types.py
:caption: extending/custom_types.py
:lines: 1-7
```

:::{note}
`FlyteContext` is used to access a random local directory.
:::

Defined type here represents a list of files on the disk. We will refer to it as `MyDataset`.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/0ec8388759d34566a0ffc0c3c2d7443fd4a3a46f/examples/extending/extending/custom_types.py
:caption: extending/custom_types.py
:pyobject: MyDataset
```

`MyDataset` represents a set of files locally. However, when a workflow consists of multiple steps, we want the data to
flow between different steps. To achieve this, it is necessary to explain how the data will be transformed to
Flyte's remote references. To do this, we create a new instance of
{py:class}`~flytekit:flytekit.extend.TypeTransformer`, for the type `MyDataset` as follows:

:::{note}
The `TypeTransformer` is a Generic abstract base class. The `Generic` type argument refers to the actual object
that we want to work with. In this case, it is the `MyDataset` object.
:::

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/0ec8388759d34566a0ffc0c3c2d7443fd4a3a46f/examples/extending/extending/custom_types.py
:caption: extending/custom_types.py
:pyobject: MyDatasetTransformer
```

Before we can use MyDataset in our tasks, we need to let Flytekit know that `MyDataset` should be considered as a valid type.
This is done using {py:class}`~flytekit:flytekit.extend.TypeEngine`'s `register` method.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/0ec8388759d34566a0ffc0c3c2d7443fd4a3a46f/examples/extending/extending/custom_types.py
:caption: extending/custom_types.py
:lines: 87
```

The new type should be ready to use! Let us write an example generator and consumer for this new datatype.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/0ec8388759d34566a0ffc0c3c2d7443fd4a3a46f/examples/extending/extending/custom_types.py
:caption: extending/custom_types.py
:lines: 91-114
```

This workflow can be executed and tested locally. Flytekit will exercise the entire path even if you run it locally.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/0ec8388759d34566a0ffc0c3c2d7443fd4a3a46f/examples/extending/extending/custom_types.py
:caption: extending/custom_types.py
:lines: 119-120
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/0ec8388759d34566a0ffc0c3c2d7443fd4a3a46f/examples/extending/
