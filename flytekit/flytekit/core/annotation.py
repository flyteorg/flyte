from typing import Any, Dict


class FlyteAnnotation:
    """A core object to add arbitrary annotations to flyte types.

    This metadata is ingested as a python dictionary and will be serialized
    into fields on the flyteidl type literals. This data is not accessible at
    runtime but rather can be retrieved from flyteadmin for custom presentation
    of typed parameters.

    Flytekit expects to receive a maximum of one `FlyteAnnotation` object
    within each typehint.

    For a task definition:

    .. code-block:: python

        @task
        def x(a: typing.Annotated[int, FlyteAnnotation({"foo": {"bar": 1}})]):
            return

    """

    def __init__(self, data: Dict[str, Any]):
        self._data = data

    @property
    def data(self):
        return self._data
