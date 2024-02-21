import os
import typing
from typing import Type

import pandas

from flytekit import FlyteContext
from flytekit.core.type_engine import T, TypeEngine, TypeTransformer
from flytekit.models.literals import Literal, Scalar, Schema
from flytekit.models.types import LiteralType, SchemaType
from flytekit.types.schema import LocalIOSchemaReader, LocalIOSchemaWriter, SchemaEngine, SchemaFormat, SchemaHandler


class ParquetIO(object):
    PARQUET_ENGINE = "pyarrow"

    def _read(self, chunk: os.PathLike, columns: typing.Optional[typing.List[str]], **kwargs) -> pandas.DataFrame:
        return pandas.read_parquet(chunk, columns=columns, engine=self.PARQUET_ENGINE, **kwargs)

    def read(
        self, *files: os.PathLike, columns: typing.Optional[typing.List[str]] = None, **kwargs
    ) -> pandas.DataFrame:
        frames = [self._read(chunk=f, columns=columns, **kwargs) for f in files if os.path.getsize(f) > 0]
        if len(frames) == 1:
            return frames[0]
        elif len(frames) > 1:
            return pandas.concat(frames, copy=True)
        return pandas.DataFrame()

    def write(
        self,
        df: pandas.DataFrame,
        to_file: os.PathLike,
        coerce_timestamps: str = "us",
        allow_truncated_timestamps: bool = False,
        **kwargs,
    ):
        """
        Writes data frame as a chunk to the local directory owned by the Schema object.  Will later be uploaded to s3.
        :param df: data frame to write as parquet
        :param to_file: Sink file to write the dataframe to
        :param coerce_timestamps: format to store timestamp in parquet. 'us', 'ms', 's' are allowed values.
            Note: if your timestamps will lose data due to the coercion, your write will fail!  Nanoseconds are
            problematic in the Parquet format and will not work. See allow_truncated_timestamps.
        :param allow_truncated_timestamps: default False. Allow truncation when coercing timestamps to a coarser
            resolution.
        """
        # TODO @ketan validate and remove this comment, as python 3 all strings are unicode
        # Convert all columns to unicode as pyarrow's parquet reader can not handle mixed strings and unicode.
        # Since columns from Hive are returned as unicode, if a user wants to add a column to a dataframe returned from
        # Hive, then output the new data, the user would have to provide a unicode column name which is unnatural.
        df.to_parquet(
            to_file,
            coerce_timestamps=coerce_timestamps,
            allow_truncated_timestamps=allow_truncated_timestamps,
            **kwargs,
        )


class PandasSchemaReader(LocalIOSchemaReader[pandas.DataFrame]):
    def __init__(self, local_dir: str, cols: typing.Optional[typing.Dict[str, type]], fmt: SchemaFormat):
        super().__init__(local_dir, cols, fmt)
        self._parquet_engine = ParquetIO()

    def _read(self, *path: os.PathLike, **kwargs) -> pandas.DataFrame:
        return self._parquet_engine.read(*path, columns=self.column_names, **kwargs)


class PandasSchemaWriter(LocalIOSchemaWriter[pandas.DataFrame]):
    def __init__(self, local_dir: str, cols: typing.Optional[typing.Dict[str, type]], fmt: SchemaFormat):
        super().__init__(local_dir, cols, fmt)
        self._parquet_engine = ParquetIO()

    def _write(self, df: T, path: os.PathLike, **kwargs):
        return self._parquet_engine.write(df, to_file=path, **kwargs)


class PandasDataFrameTransformer(TypeTransformer[pandas.DataFrame]):
    """
    Transforms a pd.DataFrame to Schema without column types.
    """

    def __init__(self):
        super().__init__("PandasDataFrame<->GenericSchema", pandas.DataFrame)
        self._parquet_engine = ParquetIO()

    @staticmethod
    def _get_schema_type() -> SchemaType:
        return SchemaType(columns=[])

    def get_literal_type(self, t: Type[pandas.DataFrame]) -> LiteralType:
        return LiteralType(schema=self._get_schema_type())

    def to_literal(
        self,
        ctx: FlyteContext,
        python_val: pandas.DataFrame,
        python_type: Type[pandas.DataFrame],
        expected: LiteralType,
    ) -> Literal:
        local_dir = ctx.file_access.get_random_local_directory()
        w = PandasSchemaWriter(local_dir=local_dir, cols=None, fmt=SchemaFormat.PARQUET)
        w.write(python_val)
        remote_path = ctx.file_access.join(
            ctx.file_access.raw_output_prefix,
            ctx.file_access.get_random_string(),
        )
        remote_path = ctx.file_access.put_data(local_dir, remote_path, is_multipart=True)
        return Literal(scalar=Scalar(schema=Schema(remote_path, self._get_schema_type())))

    def to_python_value(
        self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[pandas.DataFrame]
    ) -> pandas.DataFrame:
        if not (lv and lv.scalar and lv.scalar.schema):
            return pandas.DataFrame()
        local_dir = ctx.file_access.get_random_local_directory()
        ctx.file_access.get_data(lv.scalar.schema.uri, local_dir, is_multipart=True)
        r = PandasSchemaReader(local_dir=local_dir, cols=None, fmt=SchemaFormat.PARQUET)
        return r.all()

    def to_html(self, ctx: FlyteContext, python_val: pandas.DataFrame, expected_python_type: Type[T]):
        return python_val.describe().to_html()


SchemaEngine.register_handler(
    SchemaHandler("pandas-dataframe-schema", pandas.DataFrame, PandasSchemaReader, PandasSchemaWriter)
)
TypeEngine.register(PandasDataFrameTransformer())
