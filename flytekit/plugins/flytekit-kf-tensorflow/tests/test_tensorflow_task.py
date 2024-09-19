import pytest
from flytekitplugins.kftensorflow import PS, Chief, CleanPodPolicy, Evaluator, RestartPolicy, RunPolicy, TfJob, Worker

from flytekit import Resources, task
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


def test_tensorflow_task_with_default_config(serialization_settings: SerializationSettings):
    task_config = TfJob(
        worker=Worker(replicas=1),
        chief=Chief(replicas=0),
        ps=PS(replicas=0),
        evaluator=Evaluator(replicas=0),
    )

    @task(
        task_config=task_config,
        cache=True,
        requests=Resources(cpu="1"),
        cache_version="1",
    )
    def my_tensorflow_task(x: int, y: str) -> int:
        return x

    assert my_tensorflow_task(x=10, y="hello") == 10

    assert my_tensorflow_task.task_config is not None
    assert my_tensorflow_task.task_type == "tensorflow"
    assert my_tensorflow_task.resources.limits == Resources()
    assert my_tensorflow_task.resources.requests == Resources(cpu="1")

    expected_dict = {
        "chiefReplicas": {
            "resources": {},
        },
        "workerReplicas": {
            "replicas": 1,
            "resources": {},
        },
        "psReplicas": {
            "resources": {},
        },
        "evaluatorReplicas": {
            "resources": {},
        },
    }
    assert my_tensorflow_task.get_custom(serialization_settings) == expected_dict


def test_tensorflow_task_with_custom_config(serialization_settings: SerializationSettings):
    task_config = TfJob(
        chief=Chief(
            replicas=1,
            requests=Resources(cpu="1"),
            limits=Resources(cpu="2"),
            image="chief:latest",
        ),
        worker=Worker(
            replicas=5,
            requests=Resources(cpu="2", mem="2Gi"),
            limits=Resources(cpu="4", mem="2Gi"),
            image="worker:latest",
            restart_policy=RestartPolicy.FAILURE,
        ),
        ps=PS(
            replicas=2,
            restart_policy=RestartPolicy.ALWAYS,
        ),
        evaluator=Evaluator(
            replicas=5,
            requests=Resources(cpu="2", mem="2Gi"),
            limits=Resources(cpu="4", mem="2Gi"),
            image="evaluator:latest",
            restart_policy=RestartPolicy.FAILURE,
        ),
    )

    @task(
        task_config=task_config,
        cache=True,
        requests=Resources(cpu="1"),
        cache_version="1",
    )
    def my_tensorflow_task(x: int, y: str) -> int:
        return x

    assert my_tensorflow_task(x=10, y="hello") == 10

    assert my_tensorflow_task.task_config is not None
    assert my_tensorflow_task.task_type == "tensorflow"
    assert my_tensorflow_task.resources.limits == Resources()
    assert my_tensorflow_task.resources.requests == Resources(cpu="1")

    expected_custom_dict = {
        "chiefReplicas": {
            "replicas": 1,
            "image": "chief:latest",
            "resources": {
                "requests": [{"name": "CPU", "value": "1"}],
                "limits": [{"name": "CPU", "value": "2"}],
            },
        },
        "workerReplicas": {
            "replicas": 5,
            "image": "worker:latest",
            "resources": {
                "requests": [
                    {"name": "CPU", "value": "2"},
                    {"name": "MEMORY", "value": "2Gi"},
                ],
                "limits": [
                    {"name": "CPU", "value": "4"},
                    {"name": "MEMORY", "value": "2Gi"},
                ],
            },
            "restartPolicy": "RESTART_POLICY_ON_FAILURE",
        },
        "psReplicas": {
            "resources": {},
            "replicas": 2,
            "restartPolicy": "RESTART_POLICY_ALWAYS",
        },
        "evaluatorReplicas": {
            "replicas": 5,
            "image": "evaluator:latest",
            "resources": {
                "requests": [
                    {"name": "CPU", "value": "2"},
                    {"name": "MEMORY", "value": "2Gi"},
                ],
                "limits": [
                    {"name": "CPU", "value": "4"},
                    {"name": "MEMORY", "value": "2Gi"},
                ],
            },
            "restartPolicy": "RESTART_POLICY_ON_FAILURE",
        },
    }

    assert my_tensorflow_task.get_custom(serialization_settings) == expected_custom_dict


def test_tensorflow_task_with_run_policy(serialization_settings: SerializationSettings):
    task_config = TfJob(
        worker=Worker(replicas=1),
        ps=PS(replicas=0),
        chief=Chief(replicas=0),
        evaluator=Evaluator(replicas=0),
        run_policy=RunPolicy(
            clean_pod_policy=CleanPodPolicy.RUNNING,
            backoff_limit=5,
            active_deadline_seconds=100,
            ttl_seconds_after_finished=100,
        ),
    )

    @task(
        task_config=task_config,
        cache=True,
        requests=Resources(cpu="1"),
        cache_version="1",
    )
    def my_tensorflow_task(x: int, y: str) -> int:
        return x

    assert my_tensorflow_task(x=10, y="hello") == 10

    assert my_tensorflow_task.task_config is not None
    assert my_tensorflow_task.task_type == "tensorflow"
    assert my_tensorflow_task.resources.limits == Resources()
    assert my_tensorflow_task.resources.requests == Resources(cpu="1")

    expected_dict = {
        "chiefReplicas": {
            "resources": {},
        },
        "workerReplicas": {
            "replicas": 1,
            "resources": {},
        },
        "psReplicas": {
            "resources": {},
        },
        "evaluatorReplicas": {
            "resources": {},
        },
        "runPolicy": {
            "cleanPodPolicy": "CLEANPOD_POLICY_RUNNING",
            "backoffLimit": 5,
            "activeDeadlineSeconds": 100,
            "ttlSecondsAfterFinished": 100,
        },
    }

    assert my_tensorflow_task.get_custom(serialization_settings) == expected_dict


def test_tensorflow_task():
    @task(
        task_config=TfJob(num_workers=10, num_ps_replicas=1, num_chief_replicas=1, num_evaluator_replicas=1),
        cache=True,
        requests=Resources(cpu="1"),
        cache_version="1",
    )
    def my_tensorflow_task(x: int, y: str) -> int:
        return x

    assert my_tensorflow_task(x=10, y="hello") == 10

    assert my_tensorflow_task.task_config is not None

    default_img = Image(name="default", fqn="test", tag="tag")
    settings = SerializationSettings(
        project="project",
        domain="domain",
        version="version",
        env={"FOO": "baz"},
        image_config=ImageConfig(default_image=default_img, images=[default_img]),
    )

    expected_dict = {
        "chiefReplicas": {
            "replicas": 1,
            "resources": {},
        },
        "workerReplicas": {
            "replicas": 10,
            "resources": {},
        },
        "psReplicas": {
            "replicas": 1,
            "resources": {},
        },
        "evaluatorReplicas": {
            "replicas": 1,
            "resources": {},
        },
    }

    assert my_tensorflow_task.get_custom(settings) == expected_dict
    assert my_tensorflow_task.resources.limits == Resources()
    assert my_tensorflow_task.resources.requests == Resources(cpu="1")
    assert my_tensorflow_task.task_type == "tensorflow"
