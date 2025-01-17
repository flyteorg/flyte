from typing import Optional, TypeVar

from flytekit import FlyteContextManager
from flytekit.sensor.base_sensor import BaseSensor

T = TypeVar("T")


class FileSensor(BaseSensor):
    def __init__(self, name: str, config: Optional[T] = None, **kwargs):
        super().__init__(name=name, sensor_config=config, **kwargs)

    async def poke(self, path: str) -> bool:
        file_access = FlyteContextManager.current_context().file_access
        fs = file_access.get_filesystem_for_path(path, asynchronous=True)
        if file_access.is_remote(path):
            return await fs._exists(path)
        return fs.exists(path)
