import pytest
from flytekitplugins.dask import Dask, Scheduler, WorkerGroup

from flytekit import PythonFunctionTask, Resources, task
from flytekit.configuration import Image, ImageConfig, SerializationSettings


@pytest.fixture
def serialization_settings() -> SerializationSettings:
    default_img = Image(name="default", fqn="test", tag="tag")
    settings = SerializationSettings(
        project="project",
        domain="domain",
        version="version",
        env={"FOO": "baz"},
        image_config=ImageConfig(default_image=default_img, images=[default_img]),
    )
    return settings


def test_dask_task_with_default_config(serialization_settings: SerializationSettings):
    task_config = Dask()

    @task(task_config=task_config)
    def dask_task():
        pass

    # Helping type completion in PyCharm
    dask_task: PythonFunctionTask[Dask]

    assert dask_task.task_config == task_config
    assert dask_task.task_type == "dask"

    expected_dict = {
        "scheduler": {
            "resources": {},
        },
        "workers": {
            "numberOfWorkers": 1,
            "resources": {},
        },
    }
    assert dask_task.get_custom(serialization_settings) == expected_dict


def test_dask_task_get_custom(serialization_settings: SerializationSettings):
    task_config = Dask(
        scheduler=Scheduler(
            image="scheduler:latest",
            requests=Resources(cpu="1"),
            limits=Resources(cpu="2"),
        ),
        workers=WorkerGroup(
            number_of_workers=123,
            image="dask_cluster:latest",
            requests=Resources(cpu="3"),
            limits=Resources(cpu="4"),
        ),
    )

    @task(task_config=task_config)
    def dask_task():
        pass

    # Helping type completion in PyCharm
    dask_task: PythonFunctionTask[Dask]

    expected_custom_dict = {
        "scheduler": {
            "image": "scheduler:latest",
            "resources": {
                "requests": [{"name": "CPU", "value": "1"}],
                "limits": [{"name": "CPU", "value": "2"}],
            },
        },
        "workers": {
            "numberOfWorkers": 123,
            "image": "dask_cluster:latest",
            "resources": {
                "requests": [{"name": "CPU", "value": "3"}],
                "limits": [{"name": "CPU", "value": "4"}],
            },
        },
    }
    custom_dict = dask_task.get_custom(serialization_settings)
    assert custom_dict == expected_custom_dict
