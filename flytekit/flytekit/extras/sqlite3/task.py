import contextlib
import os
import shutil
import sqlite3
import tempfile
import typing
from dataclasses import dataclass

from flytekit import FlyteContext, kwtypes, lazy_module
from flytekit.configuration import DefaultImages, SerializationSettings
from flytekit.core.base_sql_task import SQLTask
from flytekit.core.python_customized_container_task import PythonCustomizedContainerTask
from flytekit.core.shim_task import ShimTaskExecutor
from flytekit.models import task as task_models

if typing.TYPE_CHECKING:
    import pandas as pd
else:
    pd = lazy_module("pandas")


def unarchive_file(local_path: str, to_dir: str):
    """
    Unarchive given archive and returns the unarchived file name. It is expected that only one file is unarchived.
    More than one file or 0 files will result in a ``RuntimeError``
    """
    archive_dir = os.path.join(to_dir, "_arch")
    shutil.unpack_archive(local_path, archive_dir)
    # file gets uncompressed into to_dir/_arch/*.*
    files = os.listdir(archive_dir)
    if not files or len(files) == 0 or len(files) > 1:
        raise RuntimeError(f"Uncompressed archive should contain only one file - found {files}!")
    return os.path.join(archive_dir, files[0])


@dataclass
class SQLite3Config(object):
    """
    Use this configuration to configure if sqlite3 files that should be loaded by the task. The file itself is
    considered as a database and hence is treated like a configuration
    The path to a static sqlite3 compatible database file can be

    -  within the container
    - or from a publicly downloadable source

    Args:
        uri: default FlyteFile that will be downloaded on execute
        compressed: Boolean that indicates if the given file is a compressed archive. Supported file types are
            [zip, tar, gztar, bztar, xztar]
    """

    uri: str
    compressed: bool = False


class SQLite3Task(PythonCustomizedContainerTask[SQLite3Config], SQLTask[SQLite3Config]):
    """
    Run client side SQLite3 queries that optionally return a FlyteSchema object.

    .. note::

       This is a pre-built container task. That is, your user container will not be used at task execution time.
       Instead the image defined in this task definition will be used instead.

    .. literalinclude:: ../../../tests/flytekit/unit/extras/sqlite3/test_task.py
       :start-after: # sqlite3_start
       :end-before: # sqlite3_end
       :language: python
       :dedent: 4

    See the :ref:`integrations guide <cookbook:integrations_sql_sqlite3>` for additional usage examples and
    the base class :py:class:`flytekit.extend.PythonCustomizedContainerTask` as well.
    """

    _SQLITE_TASK_TYPE = "sqlite"

    def __init__(
        self,
        name: str,
        query_template: str,
        inputs: typing.Optional[typing.Dict[str, typing.Type]] = None,
        task_config: typing.Optional[SQLite3Config] = None,
        output_schema_type: typing.Optional[typing.Type["FlyteSchema"]] = None,  # type: ignore
        container_image: typing.Optional[str] = None,
        **kwargs,
    ):
        if task_config is None or task_config.uri is None:
            raise ValueError("SQLite DB uri is required.")
        from flytekit.types.schema import FlyteSchema

        outputs = kwtypes(results=output_schema_type if output_schema_type else FlyteSchema)
        super().__init__(
            name=name,
            task_config=task_config,
            # if you use your own image, keep in mind to specify the container image here
            container_image=container_image or DefaultImages.default_image(),
            executor_type=SQLite3TaskExecutor,
            task_type=self._SQLITE_TASK_TYPE,
            # Sanitize query by removing the newlines at the end of the query. Keep in mind
            # that the query can be a multiline string.
            query_template=query_template,
            inputs=inputs,
            outputs=outputs,
            **kwargs,
        )

    @property
    def output_columns(self) -> typing.Optional[typing.List[str]]:
        c = self.python_interface.outputs["results"].column_names()
        return c if c else None

    def get_custom(self, settings: SerializationSettings) -> typing.Dict[str, typing.Any]:
        return {
            "query_template": self.query_template,
            "uri": self.task_config.uri,
            "compressed": self.task_config.compressed,
        }


class SQLite3TaskExecutor(ShimTaskExecutor[SQLite3Task]):
    def execute_from_model(self, tt: task_models.TaskTemplate, **kwargs) -> typing.Any:
        with tempfile.TemporaryDirectory() as temp_dir:
            ctx = FlyteContext.current_context()
            file_ext = os.path.basename(tt.custom["uri"])
            local_path = os.path.join(temp_dir, file_ext)
            ctx.file_access.get_data(tt.custom["uri"], local_path)
            if tt.custom["compressed"]:
                local_path = unarchive_file(local_path, temp_dir)

            print(f"Connecting to db {local_path}")
            interpolated_query = SQLite3Task.interpolate_query(tt.custom["query_template"], **kwargs)
            print(f"Interpolated query {interpolated_query}")
            with contextlib.closing(sqlite3.connect(local_path)) as con:
                df = pd.read_sql_query(interpolated_query, con)
                return df
