import collections
import typing

from typing_extensions import get_args

from flytekit import FlyteContext, Literal, LiteralType
from flytekit.core.type_engine import TypeEngine, TypeTransformer, TypeTransformerFailedError
from flytekit.models import types as _type_models
from flytekit.models.literals import LiteralCollection

T = typing.TypeVar("T")


class FlyteIterator:
    def __init__(self, ctx: FlyteContext, lv: Literal, expected_python_type: typing.Type[T], length: int):
        self._ctx = ctx
        self._lv = lv
        self._expected_python_type = expected_python_type
        self._length = length
        self._index = 0

    def __len__(self):
        return self._length

    def __iter__(self):
        self._index = 0
        return self

    def __next__(self):
        if self._index < self._length:
            lits = self._lv.collection.literals
            st = get_args(self._expected_python_type)[0]
            lt = TypeEngine.to_python_value(self._ctx, lits[self._index], st)
            self._index += 1
            return lt

        else:
            raise StopIteration


class IteratorTransformer(TypeTransformer[typing.Iterator]):
    def __init__(self):
        super().__init__("Typed Iterator", typing.Iterator)

    def get_literal_type(self, t: typing.Type[T]) -> typing.Optional[LiteralType]:
        try:
            sub_type = TypeEngine.to_literal_type(get_args(t)[0])
            return _type_models.LiteralType(collection_type=sub_type)
        except Exception as e:
            raise ValueError(f"Type of Generic List type is not supported, {e}")

    def to_literal(
        self, ctx: FlyteContext, python_val: typing.Iterator[T], python_type: typing.Type[T], expected: LiteralType
    ) -> Literal:
        t = get_args(python_type)[0]
        lit_list = [TypeEngine.to_literal(ctx, x, t, expected.collection_type) for x in python_val]
        return Literal(collection=LiteralCollection(literals=lit_list))

    def to_python_value(self, ctx: FlyteContext, lv: Literal, expected_python_type: typing.Type[T]) -> FlyteIterator:
        try:
            lits = lv.collection.literals
        except AttributeError:
            raise TypeTransformerFailedError()
        return FlyteIterator(ctx, lv, expected_python_type, len(lits))


TypeEngine.register(IteratorTransformer(), [collections.abc.Iterator])
