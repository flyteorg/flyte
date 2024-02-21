import pytest
from flytekitplugins.kfmpi import CleanPodPolicy, HorovodJob, Launcher, MPIJob, RestartPolicy, RunPolicy, Worker
from flytekitplugins.kfmpi.task import MPIFunctionTask

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


def test_mpi_task(serialization_settings: SerializationSettings):
    @task(
        task_config=MPIJob(num_workers=10, num_launcher_replicas=10, slots=1),
        requests=Resources(cpu="1"),
        cache=True,
        cache_version="1",
    )
    def my_mpi_task(x: int, y: str) -> int:
        return x

    assert my_mpi_task(x=10, y="hello") == 10

    assert my_mpi_task.task_config is not None

    assert my_mpi_task.get_custom(serialization_settings) == {
        "launcherReplicas": {"replicas": 10, "resources": {}},
        "workerReplicas": {"replicas": 10, "resources": {}},
        "slots": 1,
    }
    assert my_mpi_task.task_type == "mpi"
    assert my_mpi_task.task_type_version == 1


def test_mpi_task_with_default_config(serialization_settings: SerializationSettings):
    task_config = MPIJob(
        worker=Worker(replicas=1),
        launcher=Launcher(replicas=1),
    )

    @task(
        task_config=task_config,
        cache=True,
        requests=Resources(cpu="1"),
        cache_version="1",
    )
    def my_mpi_task(x: int, y: str) -> int:
        return x

    assert my_mpi_task(x=10, y="hello") == 10

    assert my_mpi_task.task_config is not None
    assert my_mpi_task.task_type == "mpi"
    assert my_mpi_task.resources.limits == Resources()
    assert my_mpi_task.resources.requests == Resources(cpu="1")
    assert " ".join(my_mpi_task.get_command(serialization_settings)).startswith(
        " ".join(MPIFunctionTask._MPI_BASE_COMMAND + ["-np", "1"])
    )

    expected_dict = {
        "launcherReplicas": {
            "replicas": 1,
            "resources": {},
        },
        "workerReplicas": {
            "replicas": 1,
            "resources": {},
        },
        "slots": 1,
    }
    assert my_mpi_task.get_custom(serialization_settings) == expected_dict


def test_mpi_task_with_custom_config(serialization_settings: SerializationSettings):
    task_config = MPIJob(
        launcher=Launcher(
            replicas=1,
            requests=Resources(cpu="1"),
            limits=Resources(cpu="2"),
            image="launcher:latest",
        ),
        worker=Worker(
            replicas=5,
            requests=Resources(cpu="2", mem="2Gi"),
            limits=Resources(cpu="4", mem="2Gi"),
            image="worker:latest",
            restart_policy=RestartPolicy.NEVER,
        ),
        run_policy=RunPolicy(
            clean_pod_policy=CleanPodPolicy.ALL,
        ),
        slots=2,
    )

    @task(
        task_config=task_config,
        cache=True,
        requests=Resources(cpu="1"),
        cache_version="1",
    )
    def my_mpi_task(x: int, y: str) -> int:
        return x

    assert my_mpi_task(x=10, y="hello") == 10

    assert my_mpi_task.task_config is not None
    assert my_mpi_task.task_type == "mpi"
    assert my_mpi_task.task_type_version == 1
    assert my_mpi_task.resources.limits == Resources()
    assert my_mpi_task.resources.requests == Resources(cpu="1")
    assert " ".join(my_mpi_task.get_command(serialization_settings)).startswith(
        " ".join(MPIFunctionTask._MPI_BASE_COMMAND + ["-np", "1"])
    )

    expected_custom_dict = {
        "launcherReplicas": {
            "replicas": 1,
            "image": "launcher:latest",
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
        },
        "slots": 2,
        "runPolicy": {"cleanPodPolicy": "CLEANPOD_POLICY_ALL"},
    }
    assert my_mpi_task.get_custom(serialization_settings) == expected_custom_dict


def test_horovod_task(serialization_settings):
    @task(
        task_config=HorovodJob(
            launcher=Launcher(
                replicas=1,
                requests=Resources(cpu="1"),
                limits=Resources(cpu="2"),
            ),
            worker=Worker(
                replicas=1,
                command=["/usr/sbin/sshd", "-De", "-f", "/home/jobuser/.sshd_config"],
                restart_policy=RestartPolicy.NEVER,
            ),
            slots=2,
            verbose=False,
            log_level="INFO",
            run_policy=RunPolicy(
                clean_pod_policy=CleanPodPolicy.NONE,
                backoff_limit=5,
                active_deadline_seconds=100,
                ttl_seconds_after_finished=100,
            ),
        ),
    )
    def my_horovod_task():
        ...

    cmd = my_horovod_task.get_command(serialization_settings)
    assert "horovodrun" in cmd
    assert "--verbose" not in cmd
    assert "--log-level" in cmd
    assert "INFO" in cmd
    # CleanPodPolicy.NONE is the default, so it should not be in the output dictionary
    expected_dict = {
        "launcherReplicas": {
            "replicas": 1,
            "resources": {
                "requests": [
                    {"name": "CPU", "value": "1"},
                ],
                "limits": [
                    {"name": "CPU", "value": "2"},
                ],
            },
        },
        "workerReplicas": {
            "replicas": 1,
            "resources": {},
            "command": ["/usr/sbin/sshd", "-De", "-f", "/home/jobuser/.sshd_config"],
        },
        "slots": 2,
        "runPolicy": {
            "backoffLimit": 5,
            "activeDeadlineSeconds": 100,
            "ttlSecondsAfterFinished": 100,
        },
    }
    assert my_horovod_task.task_type_version == 1
    assert my_horovod_task.get_custom(serialization_settings) == expected_dict
