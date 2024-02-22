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

First, import the libraries.

```{code-cell}
import csv
import os
from collections import defaultdict
from typing import List

import flytekit
from flytekit import task, workflow
from flytekit.types.file import FlyteFile
```

+++ {"lines_to_next_cell": 0}

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

```{code-cell}
@task
def normalize_columns(
    csv_url: FlyteFile,
    column_names: List[str],
    columns_to_normalize: List[str],
    output_location: str,
) -> FlyteFile:
    # read the data from the raw csv file
    parsed_data = defaultdict(list)
    with open(csv_url, newline="\n") as input_file:
        reader = csv.DictReader(input_file, fieldnames=column_names)
        next(reader)  # Skip header
        for row in reader:
            for column in columns_to_normalize:
                parsed_data[column].append(float(row[column].strip()))

    # normalize the data
    normalized_data = defaultdict(list)
    for colname, values in parsed_data.items():
        mean = sum(values) / len(values)
        std = (sum([(x - mean) ** 2 for x in values]) / len(values)) ** 0.5
        normalized_data[colname] = [(x - mean) / std for x in values]

    # write to local path
    out_path = os.path.join(
        flytekit.current_context().working_directory,
        f"normalized-{os.path.basename(csv_url.path).rsplit('.')[0]}.csv",
    )
    with open(out_path, mode="w") as output_file:
        writer = csv.DictWriter(output_file, fieldnames=columns_to_normalize)
        writer.writeheader()
        for row in zip(*normalized_data.values()):
            writer.writerow({k: row[i] for i, k in enumerate(columns_to_normalize)})

    if output_location:
        return FlyteFile(path=out_path, remote_path=output_location)
    else:
        return FlyteFile(path=out_path)
```

+++ {"lines_to_next_cell": 0}

When the image URL is sent to the task, the Flytekit engine translates it into a `FlyteFile` object on the local
drive (but doesn't download it). The act of calling the `download()` method should trigger the download, and the `path`
attribute enables to `open` the file.

If the `output_location` argument is specified, it will be passed to the `remote_path` argument of `FlyteFile`,
which will use that path as the storage location instead of a random location (Flyte's object store).

When this task finishes, the Flytekit engine returns the `FlyteFile` instance, uploads the file to the location, and
creates a blob literal pointing to it.

Lastly, define a workflow. The `normalize_csv_files` workflow has an `output_location` argument which is passed
to the `location` input of the task. If it's not an empty string, the task attempts to
upload its file to that location.

```{code-cell}
@workflow
def normalize_csv_file(
    csv_url: FlyteFile,
    column_names: List[str],
    columns_to_normalize: List[str],
    output_location: str = "",
) -> FlyteFile:
    return normalize_columns(
        csv_url=csv_url,
        column_names=column_names,
        columns_to_normalize=columns_to_normalize,
        output_location=output_location,
    )
```

+++ {"lines_to_next_cell": 0}

You can run the workflow locally as follows:

```{code-cell}
if __name__ == "__main__":
    default_files = [
        (
            "https://people.sc.fsu.edu/~jburkardt/data/csv/biostats.csv",
            ["Name", "Sex", "Age", "Heights (in)", "Weight (lbs)"],
            ["Age"],
        ),
        (
            "https://people.sc.fsu.edu/~jburkardt/data/csv/faithful.csv",
            ["Index", "Eruption length (mins)", "Eruption wait (mins)"],
            ["Eruption length (mins)"],
        ),
    ]
    print(f"Running {__file__} main...")
    for index, (csv_url, column_names, columns_to_normalize) in enumerate(default_files):
        normalized_columns = normalize_csv_file(
            csv_url=csv_url,
            column_names=column_names,
            columns_to_normalize=columns_to_normalize,
        )
        print(f"Running normalize_csv_file workflow on {csv_url}: " f"{normalized_columns}")
```
