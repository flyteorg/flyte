from dataclasses import dataclass, field
from typing import Any, Callable, Dict, Optional

from flytekitplugins.dask import models
from google.protobuf.json_format import MessageToDict

from flytekit import PythonFunctionTask, Resources
from flytekit.configuration import SerializationSettings
from flytekit.core.resources import convert_resources_to_resource_model
from flytekit.core.task import TaskPlugins


@dataclass
class Scheduler:
    """
    Configuration for the scheduler pod

    :param image: Custom image to use. If ``None``, will use the same image the task was registered with. Optional,
        defaults to ``None``. The image must have ``dask[distributed]`` installed and should have the same Python
        environment as the rest of the cluster (job runner pod + worker pods).
    :param requests: Resources to request for the scheduler pod. If ``None``, the requests passed into the task will be
        used. Optional, defaults to ``None``.
    :param limits: Resource limits for the scheduler pod. If ``None``, the limits passed into the task will be used.
        Optional, defaults to ``None``.
    """

    image: Optional[str] = None
    requests: Optional[Resources] = None
    limits: Optional[Resources] = None


@dataclass
class WorkerGroup:
    """
    Configuration for a group of dask worker pods

    :param number_of_workers: Number of workers to use. Optional, defaults to 1.
    :param image: Custom image to use. If ``None``, will use the same image the task was registered with. Optional,
        defaults to ``None``. The image must have ``dask[distributed]`` installed. The provided image should have the
        same Python environment as the job runner/driver as well as the scheduler.
    :param requests: Resources to request for the worker pods. If ``None``, the requests passed into the task will be
        used. Optional, defaults to ``None``.
    :param limits: Resource limits for the worker pods. If ``None``, the limits passed into the task will be used.
        Optional, defaults to ``None``.
    """

    number_of_workers: Optional[int] = 1
    image: Optional[str] = None
    requests: Optional[Resources] = None
    limits: Optional[Resources] = None


@dataclass
class Dask:
    """
    Configuration for the dask task

    :param scheduler: Configuration for the scheduler pod. Optional, defaults to ``Scheduler()``.
    :param workers: Configuration for the pods of the default worker group. Optional, defaults to ``WorkerGroup()``.
    """

    scheduler: Scheduler = field(default_factory=lambda: Scheduler())
    workers: WorkerGroup = field(default_factory=lambda: WorkerGroup())


class DaskTask(PythonFunctionTask[Dask]):
    """
    Actual Plugin that transforms the local python code for execution within a dask cluster
    """

    _DASK_TASK_TYPE = "dask"

    def __init__(self, task_config: Dask, task_function: Callable, **kwargs):
        super(DaskTask, self).__init__(
            task_config=task_config,
            task_type=self._DASK_TASK_TYPE,
            task_function=task_function,
            **kwargs,
        )

    def get_custom(self, settings: SerializationSettings) -> Optional[Dict[str, Any]]:
        """
        Serialize the `dask` task config into a dict.

        :param settings: Current serialization settings
        :return: Dictionary representation of the dask task config.
        """
        scheduler = models.Scheduler(
            image=self.task_config.scheduler.image,
            resources=convert_resources_to_resource_model(
                requests=self.task_config.scheduler.requests,
                limits=self.task_config.scheduler.limits,
            ),
        )
        workers = models.WorkerGroup(
            number_of_workers=self.task_config.workers.number_of_workers,
            image=self.task_config.workers.image,
            resources=convert_resources_to_resource_model(
                requests=self.task_config.workers.requests,
                limits=self.task_config.workers.limits,
            ),
        )
        job = models.DaskJob(scheduler=scheduler, workers=workers)
        return MessageToDict(job.to_flyte_idl())


# Inject the `dask` plugin into flytekits dynamic plugin loading system
TaskPlugins.register_pythontask_plugin(Dask, DaskTask)
