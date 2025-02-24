import typing
from typing import Type

from flytekit import FlyteContext, lazy_module
from flytekit.extend import TypeEngine, TypeTransformer
from flytekit.models.literals import Literal, Scalar, Schema
from flytekit.models.types import LiteralType, SchemaType
from flytekit.types.schema import FlyteSchema, SchemaFormat, SchemaOpenMode
from flytekit.types.schema.types import FlyteSchemaTransformer
from flytekit.types.schema.types_pandas import PandasSchemaWriter

pandas = lazy_module("pandas")
pandera = lazy_module("pandera")

T = typing.TypeVar("T")


class PanderaTransformer(TypeTransformer[pandera.typing.DataFrame]):
    _SUPPORTED_TYPES: typing.Dict[
        type, SchemaType.SchemaColumn.SchemaColumnType
    ] = FlyteSchemaTransformer._SUPPORTED_TYPES

    def __init__(self):
        super().__init__("Pandera Transformer", pandera.typing.DataFrame)  # type: ignore

    def _pandera_schema(self, t: Type[pandera.typing.DataFrame]):
        try:
            type_args = typing.get_args(t)
        except AttributeError:
            # for python < 3.8
            type_args = getattr(t, "__args__", None)

        if type_args:
            schema_model, *_ = type_args
            schema = schema_model.to_schema()
        else:
            schema = pandera.DataFrameSchema()  # type: ignore
        return schema

    @staticmethod
    def _get_pandas_type(pandera_dtype: pandera.dtypes.DataType):
        return pandera_dtype.type.type

    def _get_col_dtypes(self, t: Type[pandera.typing.DataFrame]):
        return {k: self._get_pandas_type(v.dtype) for k, v in self._pandera_schema(t).columns.items()}

    def _get_schema_type(self, t: Type[pandera.typing.DataFrame]) -> SchemaType:
        converted_cols: typing.List[SchemaType.SchemaColumn] = []
        for k, col in self._pandera_schema(t).columns.items():
            pandas_type = self._get_pandas_type(col.dtype)
            if pandas_type not in self._SUPPORTED_TYPES:
                raise AssertionError(f"type {pandas_type} is currently not supported by the flytekit-pandera plugin")
            converted_cols.append(SchemaType.SchemaColumn(name=k, type=self._SUPPORTED_TYPES[pandas_type]))
        return SchemaType(columns=converted_cols)

    def get_literal_type(self, t: Type[pandera.typing.DataFrame]) -> LiteralType:
        return LiteralType(schema=self._get_schema_type(t))

    def assert_type(self, t: Type[T], v: T):
        if not hasattr(t, "__origin__") and not isinstance(v, (t, pandas.DataFrame)):
            raise TypeError(f"Type of Val '{v}' is not an instance of {t}")

    def to_literal(
        self,
        ctx: FlyteContext,
        python_val: pandas.DataFrame,
        python_type: Type[pandera.typing.DataFrame],
        expected: LiteralType,
    ) -> Literal:
        if isinstance(python_val, pandas.DataFrame):
            local_dir = ctx.file_access.get_random_local_directory()
            w = PandasSchemaWriter(
                local_dir=local_dir, cols=self._get_col_dtypes(python_type), fmt=SchemaFormat.PARQUET
            )
            w.write(self._pandera_schema(python_type)(python_val))
            remote_path = ctx.file_access.put_raw_data(local_dir)
            return Literal(scalar=Scalar(schema=Schema(remote_path, self._get_schema_type(python_type))))
        else:
            raise AssertionError(
                f"Only Pandas Dataframe object can be returned from a task, returned object type {type(python_val)}"
            )

    def to_python_value(
        self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[pandera.typing.DataFrame]
    ) -> pandera.typing.DataFrame:
        if not (lv and lv.scalar and lv.scalar.schema):
            raise AssertionError("Can only convert a literal schema to a pandera schema")

        def downloader(x, y):
            ctx.file_access.get_data(x, y, is_multipart=True)

        df = FlyteSchema(
            local_path=ctx.file_access.get_random_local_directory(),
            remote_path=lv.scalar.schema.uri,
            downloader=downloader,
            supported_mode=SchemaOpenMode.READ,
        )
        return self._pandera_schema(expected_python_type)(df.open().all())


TypeEngine.register(PanderaTransformer())
