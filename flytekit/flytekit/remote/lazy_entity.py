import typing
from threading import Lock

from flytekit import FlyteContext
from flytekit.remote.remote_callable import RemoteEntity

T = typing.TypeVar("T", bound=RemoteEntity)


class LazyEntity(RemoteEntity, typing.Generic[T]):
    """
    Fetches the entity when the entity is called or when the entity is retrieved.
    The entity is derived from RemoteEntity so that it behaves exactly like the mimicked entity.
    """

    def __init__(self, name: str, getter: typing.Callable[[], T], *args, **kwargs):
        super().__init__(*args, **kwargs)
        self._entity = None
        self._getter = getter
        self._name = name
        if not self._getter:
            raise ValueError("getter method is required to create a Lazy loadable Remote Entity.")
        self._mutex = Lock()

    @property
    def name(self) -> str:
        return self._name

    def entity_fetched(self) -> bool:
        with self._mutex:
            return self._entity is not None

    @property
    def entity(self) -> T:
        """
        If not already fetched / available, then the entity will be force fetched.
        """
        with self._mutex:
            if self._entity is None:
                try:
                    self._entity = self._getter()
                except AttributeError as e:
                    raise RuntimeError(
                        f"Error downloading the entity {self._name}, (check original exception...)"
                    ) from e
            return self._entity

    def __getattr__(self, item: str) -> typing.Any:
        """
        Forwards all other attributes to entity, causing the entity to be fetched!
        """
        return getattr(self.entity, item)

    def compile(self, ctx: FlyteContext, *args, **kwargs):
        return self.entity.compile(ctx, *args, **kwargs)

    def __call__(self, *args, **kwargs):
        """
        Forwards the call to the underlying entity. The entity will be fetched if not already present
        """
        return self.entity(*args, **kwargs)

    def __repr__(self) -> str:
        return str(self)

    def __str__(self) -> str:
        return f"Promise for entity [{self._name}]"
