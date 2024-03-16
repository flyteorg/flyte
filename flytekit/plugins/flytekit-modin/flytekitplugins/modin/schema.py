import os
import typing
from typing import Type

from flytekit import FlyteContext, lazy_module
from flytekit.extend import T, TypeEngine, TypeTransformer
from flytekit.models.literals import Literal, Scalar, Schema
from flytekit.models.types import LiteralType, SchemaType
from flytekit.types.schema import LocalIOSchemaReader, LocalIOSchemaWriter, SchemaEngine, SchemaFormat, SchemaHandler
from flytekit.types.schema.types import FlyteSchemaTransformer

modin = lazy_module("modin")
pandas = lazy_module("modin.pandas")


class ModinPandasSchemaReader(LocalIOSchemaReader[pandas.DataFrame]):
    """
    Implements how ModinPandasDataFrame should be read using the ``open`` method of FlyteSchema
    """

    def __init__(
        self,
        from_path: str,
        cols: typing.Optional[typing.Dict[str, type]],
        fmt: SchemaFormat,
    ):
        super().__init__(from_path, cols, fmt)

    def all(self, **kwargs) -> pandas.DataFrame:
        if self._fmt == SchemaFormat.PARQUET:
            return pandas.read_parquet(self.from_path + "/00000")
        raise AssertionError("Only Parquet type files are supported for modin pandas dataframe currently")


class ModinPandasSchemaWriter(LocalIOSchemaWriter[pandas.DataFrame]):
    """
    Implements how ModinPandasDataFrame should be written to using ``open`` method of FlyteSchema
    """

    def __init__(
        self,
        to_path: os.PathLike,
        cols: typing.Optional[typing.Dict[str, type]],
        fmt: SchemaFormat,
    ):
        super().__init__(str(to_path), cols, fmt)

    def write(self, *dfs: pandas.DataFrame, **kwargs):
        if dfs is None or len(dfs) == 0:
            return
        if len(dfs) > 1:
            raise AssertionError("Only a single pandas.DataFrame can be written per variable currently")
        if self._fmt == SchemaFormat.PARQUET:
            dfs[0].to_parquet(self.to_path + "/00000")
            return
        raise AssertionError("Only Parquet type files are supported for pandas dataframe currently")


class ModinPandasDataFrameTransformer(TypeTransformer[pandas.DataFrame]):
    """
    Transforms ModinPandas DataFrame's to and from a Schema (typed/untyped)
    """

    _SUPPORTED_TYPES: typing.Dict[
        type, SchemaType.SchemaColumn.SchemaColumnType
    ] = FlyteSchemaTransformer._SUPPORTED_TYPES

    def __init__(self):
        super().__init__("pandas-df-transformer", pandas.DataFrame)

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
        w = ModinPandasSchemaWriter(to_path=local_dir, cols=None, fmt=SchemaFormat.PARQUET)
        w.write(python_val)
        remote_path = ctx.file_access.join(
            ctx.file_access.raw_output_prefix,
            ctx.file_access.get_random_string(),
        )
        remote_path = ctx.file_access.put_data(local_dir, remote_path, is_multipart=True)
        return Literal(scalar=Scalar(schema=Schema(remote_path, self._get_schema_type())))

    def to_python_value(
        self,
        ctx: FlyteContext,
        lv: Literal,
        expected_python_type: Type[pandas.DataFrame],
    ) -> T:
        if not (lv and lv.scalar and lv.scalar.schema):
            return pandas.DataFrame()
        local_dir = ctx.file_access.get_random_local_directory()
        ctx.file_access.download_directory(lv.scalar.schema.uri, local_dir)
        r = ModinPandasSchemaReader(from_path=local_dir, cols=None, fmt=SchemaFormat.PARQUET)
        return r.all()


SchemaEngine.register_handler(
    SchemaHandler(
        "pandas.Dataframe-Schema",
        pandas.DataFrame,
        ModinPandasSchemaReader,
        ModinPandasSchemaWriter,
    )
)

TypeEngine.register(ModinPandasDataFrameTransformer())
