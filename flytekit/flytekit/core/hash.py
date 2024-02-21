from typing import Callable, Generic, TypeVar

T = TypeVar("T")


class HashOnReferenceMixin(object):
    def __hash__(self):
        return hash(id(self))


class HashMethod(Generic[T]):
    """
    Flyte-specific object used to wrap the hash function for a specific type
    """

    def __init__(self, function: Callable[[T], str]):
        self._function = function

    def calculate(self, obj: T) -> str:
        """
        Calculate hash for `obj`.
        """
        return self._function(obj)
