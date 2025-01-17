import json
from collections import OrderedDict
from typing import List
from unittest.mock import MagicMock

import pytest
from flytekitplugins.pod.task import Pod, PodFunctionTask
from kubernetes.client import ApiClient
from kubernetes.client.models import V1Container, V1EnvVar, V1PodSpec, V1ResourceRequirements, V1VolumeMount

from flytekit import Resources, TaskMetadata, dynamic, map_task, task
from flytekit.configuration import FastSerializationSettings, Image, ImageConfig, SerializationSettings
from flytekit.core import context_manager
from flytekit.core.type_engine import TypeEngine
from flytekit.exceptions import user
from flytekit.extend import ExecutionState
from flytekit.tools.translator import get_serializable


def get_pod_spec(environment=[]):
    a_container = V1Container(
        name="a container",
        env=environment,
    )
    a_container.command = ["fee", "fi", "fo", "fum"]
    a_container.volume_mounts = [
        V1VolumeMount(
            name="volume mount",
            mount_path="some/where",
        )
    ]

    pod_spec = V1PodSpec(restart_policy="OnFailure", containers=[a_container, V1Container(name="another container")])
    return pod_spec


@pytest.mark.parametrize(
    "task_environment, podspec_env_vars, expected_environment",
    [
        ({"FOO": "bar"}, [], [V1EnvVar(name="FOO", value="bar")]),
        (
            {"FOO": "bar"},
            [V1EnvVar(name="AN_ENV_VAR", value="42")],
            [V1EnvVar(name="FOO", value="bar"), V1EnvVar(name="AN_ENV_VAR", value="42")],
        ),
        # We do not provide any validation for the duplication of env vars, neither does k8s.
        (
            {"FOO": "bar"},
            [V1EnvVar(name="FOO", value="another-bar")],
            [V1EnvVar(name="FOO", value="bar"), V1EnvVar(name="FOO", value="another-bar")],
        ),
    ],
)
def test_pod_task_deserialization(task_environment, podspec_env_vars, expected_environment):
    pod = Pod(pod_spec=get_pod_spec(podspec_env_vars), primary_container_name="a container")

    @task(task_config=pod, requests=Resources(cpu="10"), limits=Resources(gpu="2"), environment=task_environment)
    def simple_pod_task(i: int):
        pass

    assert isinstance(simple_pod_task, PodFunctionTask)
    assert simple_pod_task.task_config == pod

    default_img = Image(name="default", fqn="test", tag="tag")

    target = simple_pod_task.get_k8s_pod(
        SerializationSettings(
            project="project",
            domain="domain",
            version="version",
            env={"FOO": "baz"},
            image_config=ImageConfig(default_image=default_img, images=[default_img]),
        )
    )

    # Test that custom is correctly serialized by deserializing it with the python API client
    response = MagicMock()
    response.data = json.dumps(target.pod_spec)
    deserialized_pod_spec = ApiClient().deserialize(response, V1PodSpec)

    assert deserialized_pod_spec.restart_policy == "OnFailure"
    assert len(deserialized_pod_spec.containers) == 2
    primary_container = deserialized_pod_spec.containers[0]
    assert primary_container.name == "a container"
    assert primary_container.args == [
        "pyflyte-execute",
        "--inputs",
        "{{.input}}",
        "--output-prefix",
        "{{.outputPrefix}}",
        "--raw-output-data-prefix",
        "{{.rawOutputDataPrefix}}",
        "--checkpoint-path",
        "{{.checkpointOutputPrefix}}",
        "--prev-checkpoint",
        "{{.prevCheckpointPrefix}}",
        "--resolver",
        "flytekit.core.python_auto_container.default_task_resolver",
        "--",
        "task-module",
        "tests.test_pod",
        "task-name",
        "simple_pod_task",
    ]
    assert primary_container.volume_mounts[0].mount_path == "some/where"
    assert primary_container.volume_mounts[0].name == "volume mount"
    assert primary_container.resources == V1ResourceRequirements(limits={"gpu": "2"}, requests={"cpu": "10"})
    assert primary_container.env == expected_environment
    assert deserialized_pod_spec.containers[1].name == "another container"

    config = simple_pod_task.get_config(
        SerializationSettings(
            project="project",
            domain="domain",
            version="version",
            env={"FOO": "baz"},
            image_config=ImageConfig(default_image=default_img, images=[default_img]),
        )
    )
    assert config["primary_container_name"] == "a container"


def test_pod_task():
    pod = Pod(pod_spec=get_pod_spec(), primary_container_name="a container")

    @task(
        task_config=pod,
        requests=Resources(cpu="10"),
        limits=Resources(ephemeral_storage="1Gi", gpu="2"),
        environment={"FOO": "bar"},
    )
    def simple_pod_task(i: int):
        pass

    assert isinstance(simple_pod_task, PodFunctionTask)
    assert simple_pod_task.task_config == pod

    default_img = Image(name="default", fqn="test", tag="tag")

    pod_spec = simple_pod_task.get_k8s_pod(
        SerializationSettings(
            project="project",
            domain="domain",
            version="version",
            env={"FOO": "baz"},
            image_config=ImageConfig(default_image=default_img, images=[default_img]),
        )
    ).pod_spec

    assert pod_spec["restartPolicy"] == "OnFailure"
    assert len(pod_spec["containers"]) == 2
    primary_container = pod_spec["containers"][0]
    assert primary_container["name"] == "a container"
    assert primary_container["args"] == [
        "pyflyte-execute",
        "--inputs",
        "{{.input}}",
        "--output-prefix",
        "{{.outputPrefix}}",
        "--raw-output-data-prefix",
        "{{.rawOutputDataPrefix}}",
        "--checkpoint-path",
        "{{.checkpointOutputPrefix}}",
        "--prev-checkpoint",
        "{{.prevCheckpointPrefix}}",
        "--resolver",
        "flytekit.core.python_auto_container.default_task_resolver",
        "--",
        "task-module",
        "tests.test_pod",
        "task-name",
        "simple_pod_task",
    ]
    assert primary_container["volumeMounts"][0]["mountPath"] == "some/where"
    assert primary_container["volumeMounts"][0]["name"] == "volume mount"
    assert primary_container["resources"] == {
        "requests": {"cpu": "10"},
        "limits": {"ephemeral-storage": "1Gi", "gpu": "2"},
    }
    assert primary_container["env"] == [{"name": "FOO", "value": "bar"}]
    assert pod_spec["containers"][1]["name"] == "another container"


def test_dynamic_pod_task():
    dynamic_pod = Pod(pod_spec=get_pod_spec(), primary_container_name="a container")

    @task
    def t1(a: int) -> int:
        return a + 10

    @dynamic(
        task_config=dynamic_pod,
        requests=Resources(cpu="10"),
        limits=Resources(ephemeral_storage="1Gi", gpu="2"),
        environment={"FOO": "bar"},
    )
    def dynamic_pod_task(a: int) -> List[int]:
        s = []
        for i in range(a):
            s.append(t1(a=i))
        return s

    assert isinstance(dynamic_pod_task, PodFunctionTask)
    default_img = Image(name="default", fqn="test", tag="tag")

    pod_spec = dynamic_pod_task.get_k8s_pod(
        SerializationSettings(
            project="project",
            domain="domain",
            version="version",
            env={"FOO": "baz"},
            image_config=ImageConfig(default_image=default_img, images=[default_img]),
        )
    ).pod_spec
    assert len(pod_spec["containers"]) == 2
    primary_container = pod_spec["containers"][0]
    assert isinstance(dynamic_pod_task.task_config, Pod)
    assert primary_container["resources"] == {
        "requests": {"cpu": "10"},
        "limits": {"ephemeral-storage": "1Gi", "gpu": "2"},
    }

    config = dynamic_pod_task.get_config(
        SerializationSettings(
            project="project",
            domain="domain",
            version="version",
            env={"FOO": "baz"},
            image_config=ImageConfig(default_image=default_img, images=[default_img]),
        )
    )
    assert config["primary_container_name"] == "a container"

    with context_manager.FlyteContextManager.with_context(
        context_manager.FlyteContext.current_context().with_serialization_settings(
            SerializationSettings(
                project="test_proj",
                domain="test_domain",
                version="abc",
                image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
                env={},
            )
        )
    ) as ctx:
        with context_manager.FlyteContextManager.with_context(
            ctx.with_execution_state(ctx.execution_state.with_params(mode=ExecutionState.Mode.TASK_EXECUTION))
        ) as ctx:
            dynamic_job_spec = dynamic_pod_task.compile_into_workflow(ctx, dynamic_pod_task._task_function, a=5)
            assert len(dynamic_job_spec._nodes) == 5


def test_pod_task_undefined_primary():
    pod = Pod(pod_spec=get_pod_spec(), primary_container_name="an undefined container")

    @task(task_config=pod, requests=Resources(cpu="10"), limits=Resources(gpu="2"), environment={"FOO": "bar"})
    def simple_pod_task(i: int):
        pass

    assert isinstance(simple_pod_task, PodFunctionTask)
    assert simple_pod_task.task_config == pod

    default_img = Image(name="default", fqn="test", tag="tag")
    pod_spec = simple_pod_task.get_k8s_pod(
        SerializationSettings(
            project="project",
            domain="domain",
            version="version",
            env={"FOO": "baz"},
            image_config=ImageConfig(default_image=default_img, images=[default_img]),
        )
    ).pod_spec

    assert len(pod_spec["containers"]) == 3

    primary_container = pod_spec["containers"][2]
    assert primary_container["name"] == "an undefined container"

    config = simple_pod_task.get_config(
        SerializationSettings(
            project="project",
            domain="domain",
            version="version",
            env={"FOO": "baz"},
            image_config=ImageConfig(default_image=default_img, images=[default_img]),
        )
    )
    assert config["primary_container_name"] == "an undefined container"


def test_pod_task_serialized():
    pod = Pod(
        pod_spec=get_pod_spec(),
        primary_container_name="an undefined container",
        labels={"label": "foo"},
        annotations={"anno": "bar"},
    )

    @task(task_config=pod, requests=Resources(cpu="10"), limits=Resources(gpu="2"), environment={"FOO": "bar"})
    def simple_pod_task(i: int):
        pass

    assert isinstance(simple_pod_task, PodFunctionTask)
    assert simple_pod_task.task_config == pod

    default_img = Image(name="default", fqn="test", tag="tag")
    ssettings = SerializationSettings(
        project="project",
        domain="domain",
        version="version",
        env={"FOO": "baz"},
        image_config=ImageConfig(default_image=default_img, images=[default_img]),
    )
    serialized = get_serializable(OrderedDict(), ssettings, simple_pod_task)
    assert serialized.template.task_type_version == 2
    assert serialized.template.config["primary_container_name"] == "an undefined container"
    assert serialized.template.k8s_pod.metadata.labels == {"label": "foo"}
    assert serialized.template.k8s_pod.metadata.annotations == {"anno": "bar"}
    assert serialized.template.k8s_pod.pod_spec is not None


def test_map_pod_task_serialization():
    pod = Pod(
        pod_spec=V1PodSpec(restart_policy="OnFailure", containers=[V1Container(name="primary")]),
        primary_container_name="primary",
    )

    @task(task_config=pod, environment={"FOO": "bar"})
    def simple_pod_task(i: int):
        pass

    mapped_task = map_task(simple_pod_task, metadata=TaskMetadata(retries=1))
    default_img = Image(name="default", fqn="test", tag="tag")
    serialization_settings = SerializationSettings(
        project="project",
        domain="domain",
        version="version",
        env={"FOO": "baz"},
        image_config=ImageConfig(default_image=default_img, images=[default_img]),
    )

    # Test that target is correctly serialized with an updated command
    pod_spec = mapped_task.get_k8s_pod(serialization_settings).pod_spec

    assert len(pod_spec["containers"]) == 1
    assert pod_spec["containers"][0]["args"] == [
        "pyflyte-map-execute",
        "--inputs",
        "{{.input}}",
        "--output-prefix",
        "{{.outputPrefix}}",
        "--raw-output-data-prefix",
        "{{.rawOutputDataPrefix}}",
        "--checkpoint-path",
        "{{.checkpointOutputPrefix}}",
        "--prev-checkpoint",
        "{{.prevCheckpointPrefix}}",
        "--resolver",
        "MapTaskResolver",
        "--",
        "vars",
        "",
        "resolver",
        "flytekit.core.python_auto_container.default_task_resolver",
        "task-module",
        "tests.test_pod",
        "task-name",
        "simple_pod_task",
    ]
    assert {"primary_container_name": "primary"} == mapped_task.get_config(serialization_settings)


def test_fast_pod_task_serialization():
    pod = Pod(
        pod_spec=V1PodSpec(restart_policy="OnFailure", containers=[V1Container(name="primary")]),
        primary_container_name="primary",
    )

    @task(task_config=pod, environment={"FOO": "bar"})
    def simple_pod_task(i: int):
        pass

    default_img = Image(name="default", fqn="test", tag="tag")
    serialization_settings = SerializationSettings(
        project="project",
        domain="domain",
        version="version",
        env={"FOO": "baz"},
        image_config=ImageConfig(default_image=default_img, images=[default_img]),
        fast_serialization_settings=FastSerializationSettings(enabled=True),
    )
    serialized = get_serializable(OrderedDict(), serialization_settings, simple_pod_task)

    assert serialized.template.k8s_pod.pod_spec["containers"][0]["args"] == [
        "pyflyte-fast-execute",
        "--additional-distribution",
        "{{ .remote_package_path }}",
        "--dest-dir",
        "{{ .dest_dir }}",
        "--",
        "pyflyte-execute",
        "--inputs",
        "{{.input}}",
        "--output-prefix",
        "{{.outputPrefix}}",
        "--raw-output-data-prefix",
        "{{.rawOutputDataPrefix}}",
        "--checkpoint-path",
        "{{.checkpointOutputPrefix}}",
        "--prev-checkpoint",
        "{{.prevCheckpointPrefix}}",
        "--resolver",
        "flytekit.core.python_auto_container.default_task_resolver",
        "--",
        "task-module",
        "tests.test_pod",
        "task-name",
        "simple_pod_task",
    ]


def test_fast():
    REQUESTS_GPU = Resources(cpu="123m", mem="234Mi", ephemeral_storage="123M", gpu="1")
    LIMITS_GPU = Resources(cpu="124M", mem="235Mi", ephemeral_storage="124M", gpu="1")

    def get_minimal_pod_task_config() -> Pod:
        primary_container = V1Container(name="flytetask")
        pod_spec = V1PodSpec(containers=[primary_container])
        return Pod(pod_spec=pod_spec, primary_container_name="flytetask")

    @task(
        task_config=get_minimal_pod_task_config(),
        requests=REQUESTS_GPU,
        limits=LIMITS_GPU,
    )
    def pod_task_with_resources(dummy_input: str) -> str:
        return dummy_input

    @dynamic(requests=REQUESTS_GPU, limits=LIMITS_GPU)
    def dynamic_task_with_pod_subtask(dummy_input: str) -> str:
        pod_task_with_resources(dummy_input=dummy_input)
        return dummy_input

    default_img = Image(name="default", fqn="test", tag="tag")
    serialization_settings = SerializationSettings(
        project="project",
        domain="domain",
        version="version",
        env={"FOO": "baz"},
        image_config=ImageConfig(default_image=default_img, images=[default_img]),
        fast_serialization_settings=FastSerializationSettings(
            enabled=True,
            destination_dir="/User/flyte/workflows",
            distribution_location="s3://my-s3-bucket/fast/123",
        ),
    )

    with context_manager.FlyteContextManager.with_context(
        context_manager.FlyteContextManager.current_context().with_serialization_settings(serialization_settings)
    ) as ctx:
        with context_manager.FlyteContextManager.with_context(
            ctx.with_execution_state(
                ctx.execution_state.with_params(
                    mode=ExecutionState.Mode.TASK_EXECUTION,
                )
            )
        ) as ctx:
            input_literal_map = TypeEngine.dict_to_literal_map(ctx, {"dummy_input": "hi"})
            dynamic_job_spec = dynamic_task_with_pod_subtask.dispatch_execute(ctx, input_literal_map)
            # print(dynamic_job_spec)
            assert len(dynamic_job_spec._nodes) == 1
            assert len(dynamic_job_spec.tasks) == 1
            args = " ".join(dynamic_job_spec.tasks[0].k8s_pod.pod_spec["containers"][0]["args"])
            assert args.startswith(
                "pyflyte-fast-execute --additional-distribution s3://my-s3-bucket/fast/123 "
                "--dest-dir /User/flyte/workflows"
            )
            assert dynamic_job_spec.tasks[0].k8s_pod.pod_spec["containers"][0]["resources"]["limits"]["cpu"] == "124M"
            assert dynamic_job_spec.tasks[0].k8s_pod.pod_spec["containers"][0]["resources"]["requests"]["gpu"] == "1"

    assert context_manager.FlyteContextManager.size() == 1


def test_pod_config():
    with pytest.raises(user.FlyteValidationException):
        Pod(pod_spec=None)

    with pytest.raises(user.FlyteValidationException):
        Pod(pod_spec=V1PodSpec(containers=[]), primary_container_name=None)

    selector = {"node_group": "memory"}

    @task(
        task_config=Pod(
            pod_spec=V1PodSpec(
                containers=[],
                node_selector=selector,
            ),
        ),
        requests=Resources(
            mem="1G",
        ),
    )
    def my_pod_task():
        print("hello world")
        time.sleep(30000)

    assert my_pod_task.task_config
    assert isinstance(my_pod_task.task_config, Pod)
    assert my_pod_task.task_config.pod_spec.node_selector == selector
