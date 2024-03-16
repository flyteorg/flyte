from typing import Optional

from flyteidl.plugins import dask_pb2 as dask_task

from flytekit.models import common as common
from flytekit.models import task as task


class Scheduler(common.FlyteIdlEntity):
    """
    Configuration for the scheduler pod

    :param image: Optional image to use.
    :param resources: Optional resources to use.
    """

    def __init__(self, image: Optional[str] = None, resources: Optional[task.Resources] = None):
        self._image = image
        self._resources = resources

    @property
    def image(self) -> Optional[str]:
        """
        :return: The optional image for the scheduler pod
        """
        return self._image

    @property
    def resources(self) -> Optional[task.Resources]:
        """
        :return: Optional resources for the scheduler pod
        """
        return self._resources

    def to_flyte_idl(self) -> dask_task.DaskScheduler:
        """
        :return: The scheduler spec serialized to protobuf
        """
        return dask_task.DaskScheduler(
            image=self.image,
            resources=self.resources.to_flyte_idl() if self.resources else None,
        )


class WorkerGroup(common.FlyteIdlEntity):
    """
    Configuration for a dask worker group

    :param number_of_workers:Number of workers in the group
    :param image: Optional image to use for the pods of the worker group
    :param resources: Optional resources to use for the pods of the worker group
    """

    def __init__(
        self,
        number_of_workers: int,
        image: Optional[str] = None,
        resources: Optional[task.Resources] = None,
    ):
        if number_of_workers < 1:
            raise ValueError(
                f"Each worker group needs to have at least one worker, but {number_of_workers} have been specified."
            )

        self._number_of_workers = number_of_workers
        self._image = image
        self._resources = resources

    @property
    def number_of_workers(self) -> Optional[int]:
        """
        :return: Optional number of workers for the worker group
        """
        return self._number_of_workers

    @property
    def image(self) -> Optional[str]:
        """
        :return: The optional image to use for the worker pods
        """
        return self._image

    @property
    def resources(self) -> Optional[task.Resources]:
        """
        :return: Optional resources to use for the worker pods
        """
        return self._resources

    def to_flyte_idl(self) -> dask_task.DaskWorkerGroup:
        """
        :return: The dask cluster serialized to protobuf
        """
        return dask_task.DaskWorkerGroup(
            number_of_workers=self.number_of_workers,
            image=self.image,
            resources=self.resources.to_flyte_idl() if self.resources else None,
        )


class DaskJob(common.FlyteIdlEntity):
    """
    Configuration for the custom dask job to run

    :param scheduler: Configuration for the scheduler
    :param workers: Configuration of the default worker group
    """

    def __init__(self, scheduler: Scheduler, workers: WorkerGroup):
        self._scheduler = scheduler
        self._workers = workers

    @property
    def scheduler(self) -> Scheduler:
        """
        :return: Configuration for the scheduler pod
        """
        return self._scheduler

    @property
    def workers(self) -> WorkerGroup:
        """
        :return: Configuration of the default worker group
        """
        return self._workers

    def to_flyte_idl(self) -> dask_task.DaskJob:
        """
        :return: The dask job serialized to protobuf
        """
        return dask_task.DaskJob(
            scheduler=self.scheduler.to_flyte_idl(),
            workers=self.workers.to_flyte_idl(),
        )
