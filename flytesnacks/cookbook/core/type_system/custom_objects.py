"""
.. _dataclass_type:

Using Custom Python Objects
---------------------------

.. tags:: Basic

Flyte supports passing JSON between tasks. But to simplify the usage for the users and introduce type-safety,
Flytekit supports passing custom data objects between tasks.

Currently, data classes decorated with ``@dataclass_json`` are supported.
One good use case of a data class would be when you want to wrap all input in a data class in the case of a map task
which can only accept one input and produce one output.

This example shows how users can serialize custom JSON-compatible dataclasses between successive tasks using the
excellent `dataclasses_json <https://pypi.org/project/dataclasses-json/>`__ library.
"""

# %%
# To get started, let's import the necessary libraries.
import os
import tempfile
import typing
from dataclasses import dataclass

import pandas as pd
from dataclasses_json import dataclass_json
from flytekit import task, workflow
from flytekit.types.directory import FlyteDirectory
from flytekit.types.file import FlyteFile
from flytekit.types.schema import FlyteSchema


# %%
# We define a simple data class that can be sent between tasks.
@dataclass_json
@dataclass
class Datum(object):
    """
    Example of a simple custom class that is modeled as a dataclass
    """

    x: int
    y: str
    z: typing.Dict[int, str]


# %%
# ``Datum`` is a user defined complex type that can be used to pass complex data between tasks.
# Interestingly, users can send this data between different tasks written in different languages and input it through the Flyte Console as raw JSON.
#
# .. note::
#
#   All variables in a data class should be **annotated with their type**. Failure to do should will result in an error.

# %%
# Next, we define a data class that accepts :std:ref:`FlyteSchema <typed_schema>`, :std:ref:`FlyteFile <sphx_glr_auto_core_flyte_basics_files.py>`,
# and :std:ref:`FlyteDirectory <sphx_glr_auto_core_flyte_basics_folders.py>`.
@dataclass_json
@dataclass
class Result:
    schema: FlyteSchema
    file: FlyteFile
    directory: FlyteDirectory


# %%
# .. note::
#
#   A data class supports the usage of data associated with Python types, data classes, FlyteFile, FlyteDirectory, and FlyteSchema.
#
# Once declared, dataclasses can be returned as outputs or accepted as inputs.
#
# 1. Datum Data Class
@task
def stringify(x: int) -> Datum:
    """
    A dataclass return will be regarded as a complex single json return.
    """
    return Datum(x=x, y=str(x), z={x: str(x)})


@task
def add(x: Datum, y: Datum) -> Datum:
    """
    Flytekit will automatically convert the passed in json into a DataClass. If the structures dont match, it will raise
    a runtime failure
    """
    x.z.update(y.z)
    return Datum(x=x.x + y.x, y=x.y + y.y, z=x.z)


# %%
# The ``stringify`` task outputs a data class, and the ``add`` task accepts data classes as inputs.
#
# 2. Result Data Class
@task
def upload_result() -> Result:
    """
    Flytekit will upload FlyteFile, FlyteDirectory, and FlyteSchema to blob store (GCP, S3)
    """
    df = pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [20, 22]})
    temp_dir = tempfile.mkdtemp(prefix="flyte-")

    schema_path = temp_dir + "/schema.parquet"
    df.to_parquet(schema_path)

    file_path = tempfile.NamedTemporaryFile(delete=False)
    file_path.write(b"Hello world!")
    fs = Result(
        schema=FlyteSchema(temp_dir),
        file=FlyteFile(file_path.name),
        directory=FlyteDirectory(temp_dir),
    )
    return fs


@task
def download_result(res: Result):
    """
    Flytekit will lazily load FlyteSchema. We download the schema only when users invoke open().
    """
    assert pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [20, 22]}).equals(
        res.schema.open().all()
    )
    f = open(res.file, "r")
    assert f.read() == "Hello world!"
    assert os.listdir(res.directory) == ["schema.parquet"]


# %%
# The ``upload_result`` task outputs a data class, and the ``download_result`` task accepts data classes as inputs.

# %%
# Lastly, we create a workflow.
@workflow
def wf(x: int, y: int) -> (Datum, Result):
    """
    Dataclasses (JSON) can be returned from a workflow as well.
    """
    res = upload_result()
    download_result(res=res)
    return add(x=stringify(x=x), y=stringify(x=y)), res


# %%
# We can run the workflow locally.
if __name__ == "__main__":
    """
    This workflow can be run locally. During local execution also, the dataclasses will be marshalled to and from json.
    """
    wf(x=10, y=20)
