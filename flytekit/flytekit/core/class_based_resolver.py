from typing import List

from flytekit.configuration import SerializationSettings
from flytekit.core.base_task import TaskResolverMixin
from flytekit.core.python_auto_container import PythonAutoContainerTask
from flytekit.core.tracker import TrackedInstance


class ClassStorageTaskResolver(TrackedInstance, TaskResolverMixin):
    """
    Stores tasks inside a class variable. The class must be inherited from at the point of usage because the task
    loading process basically relies on the same sequence of things happening.
    """

    def __init__(self, *args, **kwargs):
        self.mapping = []
        super().__init__(*args, **kwargs)

    def name(self) -> str:
        return "ClassStorageTaskResolver"

    def get_all_tasks(self) -> List[PythonAutoContainerTask]:  # type:ignore
        return self.mapping

    def add(self, t: PythonAutoContainerTask):
        self.mapping.append(t)

    def load_task(self, loader_args: List[str]) -> PythonAutoContainerTask:
        if len(loader_args) != 1:
            raise RuntimeError(f"Unable to load task, received ambiguous loader args {loader_args}, expected only one")

        # string should be parseable as an int
        idx = int(loader_args[0])
        return self.mapping[idx]

    def loader_args(self, settings: SerializationSettings, t: PythonAutoContainerTask) -> List[str]:  # type: ignore
        """
        This is responsible for turning an instance of a task into args that the load_task function can reconstitute.
        """
        if t not in self.mapping:
            raise Exception("no such task")

        return [f"{self.mapping.index(t)}"]
