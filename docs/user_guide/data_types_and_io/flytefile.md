(file)=

# FlyteFile

```{eval-rst}
.. tags:: Data, Basic
```

Files are one of the most fundamental entities that users of Python work with,
and they are fully supported by Flyte. In the IDL, they are known as
[Blob](https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/core/literals.proto#L33)
literals which are backed by the
[blob type](https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/core/types.proto#L47).

Let's assume our mission here is pretty simple. We download a few CSV file
links, read them with the python built-in {py:class}`csv.DictReader` function,
normalize some pre-specified columns, and output the normalized columns to
another csv file.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

First, import the libraries:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/file.py
:caption: data_types_and_io/file.py
:lines: 1-8
```

Define a task that accepts {py:class}`~flytekit.types.file.FlyteFile` as an input.
The following is a task that accepts a `FlyteFile`, a list of column names,
and a list of column names to normalize. The task then outputs a CSV file
containing only the normalized columns. For this example, we use z-score normalization,
which involves mean-centering and standard-deviation-scaling.

:::{note}
The `FlyteFile` literal can be scoped with a string, which gets inserted
into the format of the Blob type ("jpeg" is the string in
`FlyteFile[typing.TypeVar("jpeg")]`). The format is entirely optional,
and if not specified, defaults to `""`.
Predefined aliases for commonly used flyte file formats are also available.
You can find them [here](https://github.com/flyteorg/flytekit/blob/master/flytekit/types/file/__init__.py).
:::

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/file.py
:caption: data_types_and_io/file.py
:pyobject: normalize_columns
```

When the image URL is sent to the task, the Flytekit engine translates it into a `FlyteFile` object on the local drive (but doesn't download it). The act of calling the `download()` method should trigger the download, and the `path` attribute enables to `open` the file.

If the `output_location` argument is specified, it will be passed to the `remote_path` argument of `FlyteFile`, which will use that path as the storage location instead of a random location (Flyte's object store).

When this task finishes, the Flytekit engine returns the `FlyteFile` instance, uploads the file to the location, and creates a blob literal pointing to it.

Lastly, define a workflow. The `normalize_csv_files` workflow has an `output_location` argument which is passed to the `location` input of the task. If it's not an empty string, the task attempts to upload its file to that location.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/0ec8388759d34566a0ffc0c3c2d7443fd4a3a46f/examples/data_types_and_io/data_types_and_io/file.py
:caption: data_types_and_io/file.py
:pyobject: normalize_csv_file
```

You can run the workflow locally as follows:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/0ec8388759d34566a0ffc0c3c2d7443fd4a3a46f/examples/data_types_and_io/data_types_and_io/file.py
:caption: data_types_and_io/file.py
:lines: 75-95
```

You can enable type validation if you have the [python-magic](https://pypi.org/project/python-magic/) package installed.

```{eval-rst}
.. tabs::

  .. group-tab:: Mac OS

    .. code-block:: bash

      brew install libmagic

  .. group-tab:: Linux

    .. code-block:: bash

      sudo apt-get install libmagic1
```

:::{note}
Currently, type validation is only supported on the `Mac OS` and `Linux` platforms.
:::

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/data_types_and_io/
