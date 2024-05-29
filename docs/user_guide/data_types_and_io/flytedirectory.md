(folder)=

# FlyteDirectory

```{eval-rst}
.. tags:: Data, Basic
```

In addition to files, folders are another fundamental operating system primitive.
Flyte supports folders in the form of
[multi-part blobs](https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/core/types.proto#L73).

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the libraries:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/data_types_and_io/data_types_and_io/folder.py
:caption: data_types_and_io/folder.py
:lines: 1-10
```

Building upon the previous example demonstrated in the {std:ref}`file <file>` section,
let's continue by considering the normalization of columns in a CSV file.

The following task downloads a list of URLs pointing to CSV files
and returns the folder path in a `FlyteDirectory` object.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/data_types_and_io/data_types_and_io/folder.py
:caption: data_types_and_io/folder.py
:pyobject: download_files
```

:::{note}
You can annotate a `FlyteDirectory` when you want to download or upload the contents of the directory in batches.
For example,

```{code-block}
@task
def t1(directory: Annotated[FlyteDirectory, BatchSize(10)]) -> Annotated[FlyteDirectory, BatchSize(100)]:
    ...
    return FlyteDirectory(...)
```

Flytekit efficiently downloads files from the specified input directory in 10-file chunks.
It then loads these chunks into memory before writing them to the local disk.
The process repeats for subsequent sets of 10 files.
Similarly, for outputs, Flytekit uploads the resulting directory in chunks of 100.
:::

We define a helper function to normalize the columns in-place.

:::{note}
This is a plain Python function that will be called in a subsequent Flyte task. This example
demonstrates how Flyte tasks are simply entrypoints of execution, which can themselves call
other functions and routines that are written in pure Python.
:::

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/data_types_and_io/data_types_and_io/folder.py
:caption: data_types_and_io/folder.py
:pyobject: normalize_columns
```

We then define a task that accepts the previously downloaded folder, along with some metadata about the
column names of each file in the directory and the column names that we want to normalize.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/data_types_and_io/data_types_and_io/folder.py
:caption: data_types_and_io/folder.py
:pyobject: normalize_all_files
```

Compose all of the above tasks into a workflow. This workflow accepts a list
of URL strings pointing to a remote location containing a CSV file, a list of column names
associated with each CSV file, and a list of columns that we want to normalize.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/data_types_and_io/data_types_and_io/folder.py
:caption: data_types_and_io/folder.py
:pyobject: download_and_normalize_csv_files
```

You can run the workflow locally as follows:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/data_types_and_io/data_types_and_io/folder.py
:caption: data_types_and_io/folder.py
:lines: 94-114
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/data_types_and_io/
