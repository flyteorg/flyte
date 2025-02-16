import typing

from fsspec import Callback
from rich.progress import Progress


class RichCallback(Callback):
    def __init__(self, rich_kwargs: typing.Optional[typing.Dict] = None, **kwargs):
        super().__init__(**kwargs)
        rich_kwargs = rich_kwargs or {}
        self._pb = Progress(**rich_kwargs)
        self._pb.start()
        self._task = None

    def set_size(self, size):
        self._task = self._pb.add_task("Downloading...", total=size)

    def relative_update(self, inc=1):
        self._pb.update(self._task, advance=inc)

    def __del__(self):
        self._pb.stop()
