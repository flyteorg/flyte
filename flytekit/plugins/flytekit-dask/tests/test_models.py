import pytest
from flytekitplugins.dask import models

from flytekit.models import task as _task


@pytest.fixture
def image() -> str:
    return "foo:latest"


@pytest.fixture
def resources() -> _task.Resources:
    return _task.Resources(
        requests=[
            _task.Resources.ResourceEntry(name=_task.Resources.ResourceName.CPU, value="3"),
        ],
        limits=[],
    )


@pytest.fixture
def default_resources() -> _task.Resources:
    return _task.Resources(requests=[], limits=[])


@pytest.fixture
def scheduler(image: str, resources: _task.Resources) -> models.Scheduler:
    return models.Scheduler(image=image, resources=resources)


@pytest.fixture
def workers(image: str, resources: _task.Resources) -> models.WorkerGroup:
    return models.WorkerGroup(number_of_workers=123, image=image, resources=resources)


def test_create_scheduler_to_flyte_idl_no_optional(image: str, resources: _task.Resources):
    scheduler = models.Scheduler(image=image, resources=resources)
    idl_object = scheduler.to_flyte_idl()
    assert idl_object.image == image
    assert idl_object.resources == resources.to_flyte_idl()


def test_create_scheduler_to_flyte_idl_all_optional(default_resources: _task.Resources):
    scheduler = models.Scheduler(image=None, resources=None)
    idl_object = scheduler.to_flyte_idl()
    assert idl_object.image == ""
    assert idl_object.resources == default_resources.to_flyte_idl()


def test_create_scheduler_spec_property_access(image: str, resources: _task.Resources):
    scheduler = models.Scheduler(image=image, resources=resources)
    assert scheduler.image == image
    assert scheduler.resources == resources


def test_worker_group_to_flyte_idl_no_optional(image: str, resources: _task.Resources):
    n_workers = 1234
    worker_group = models.WorkerGroup(number_of_workers=n_workers, image=image, resources=resources)
    idl_object = worker_group.to_flyte_idl()
    assert idl_object.number_of_workers == n_workers
    assert idl_object.image == image
    assert idl_object.resources == resources.to_flyte_idl()


def test_worker_group_to_flyte_idl_all_optional(default_resources: _task.Resources):
    worker_group = models.WorkerGroup(number_of_workers=1, image=None, resources=None)
    idl_object = worker_group.to_flyte_idl()
    assert idl_object.image == ""
    assert idl_object.resources == default_resources.to_flyte_idl()


def test_worker_group_property_access(image: str, resources: _task.Resources):
    n_workers = 1234
    worker_group = models.WorkerGroup(number_of_workers=n_workers, image=image, resources=resources)
    assert worker_group.image == image
    assert worker_group.number_of_workers == n_workers
    assert worker_group.resources == resources


def test_worker_group_fails_for_less_than_one_worker():
    with pytest.raises(ValueError, match=r"Each worker group needs to"):
        models.WorkerGroup(number_of_workers=0, image=None, resources=None)


def test_dask_job_to_flyte_idl_no_optional(scheduler: models.Scheduler, workers: models.WorkerGroup):
    job = models.DaskJob(scheduler=scheduler, workers=workers)
    idl_object = job.to_flyte_idl()
    assert idl_object.scheduler == scheduler.to_flyte_idl()
    assert idl_object.workers == workers.to_flyte_idl()


def test_dask_job_property_access(scheduler: models.Scheduler, workers: models.WorkerGroup):
    job = models.DaskJob(scheduler=scheduler, workers=workers)
    assert job.scheduler == scheduler
    assert job.workers == workers
