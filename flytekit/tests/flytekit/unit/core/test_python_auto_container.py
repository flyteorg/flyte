from collections import OrderedDict
from typing import Any

import pytest
from kubernetes.client.models import V1Container, V1EnvVar, V1PodSpec, V1ResourceRequirements, V1Volume

from flytekit.configuration import Image, ImageConfig, SerializationSettings
from flytekit.core.base_task import TaskMetadata
from flytekit.core.pod_template import PodTemplate
from flytekit.core.python_auto_container import PythonAutoContainerTask, get_registerable_container_image
from flytekit.core.resources import Resources
from flytekit.image_spec.image_spec import ImageBuildEngine, ImageSpec
from flytekit.tools.translator import get_serializable_task


@pytest.fixture
def default_image_config():
    default_image = Image(name="default", fqn="docker.io/xyz", tag="some-git-hash")
    return ImageConfig(default_image=default_image)


@pytest.fixture
def default_serialization_settings(default_image_config):
    return SerializationSettings(
        project="p", domain="d", version="v", image_config=default_image_config, env={"FOO": "bar"}
    )


@pytest.fixture
def minimal_serialization_settings(default_image_config):
    return SerializationSettings(project="p", domain="d", version="v", image_config=default_image_config)


def test_image_name_interpolation(default_image_config):
    img_to_interpolate = "{{.image.default.fqn}}:{{.image.default.version}}-special"
    img = get_registerable_container_image(img=img_to_interpolate, cfg=default_image_config)
    assert img == "docker.io/xyz:some-git-hash-special"


class DummyAutoContainerTask(PythonAutoContainerTask):
    def execute(self, **kwargs) -> Any:
        pass


task = DummyAutoContainerTask(name="x", task_config=None, task_type="t")


def test_default_command(default_serialization_settings):
    cmd = task.get_default_command(settings=default_serialization_settings)
    assert cmd == [
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
        "tests.flytekit.unit.core.test_python_auto_container",
        "task-name",
        "task",
    ]


def test_get_container(default_serialization_settings):
    c = task.get_container(default_serialization_settings)
    assert c.image == "docker.io/xyz:some-git-hash"
    assert c.env == {"FOO": "bar"}

    ts = get_serializable_task(OrderedDict(), default_serialization_settings, task)
    assert ts.template.container.image == "docker.io/xyz:some-git-hash"
    assert ts.template.container.env == {"FOO": "bar"}


task_with_env_vars = DummyAutoContainerTask(name="x", environment={"HAM": "spam"}, task_config=None, task_type="t")


def test_get_container_with_task_envvars(default_serialization_settings):
    c = task_with_env_vars.get_container(default_serialization_settings)
    assert c.image == "docker.io/xyz:some-git-hash"
    assert c.env == {"FOO": "bar", "HAM": "spam"}

    ts = get_serializable_task(OrderedDict(), default_serialization_settings, task_with_env_vars)
    assert ts.template.container.image == "docker.io/xyz:some-git-hash"
    assert ts.template.container.env == {"FOO": "bar", "HAM": "spam"}


def test_get_container_without_serialization_settings_envvars(minimal_serialization_settings):
    c = task_with_env_vars.get_container(minimal_serialization_settings)
    assert c.image == "docker.io/xyz:some-git-hash"
    assert c.env == {"HAM": "spam"}

    ts = get_serializable_task(OrderedDict(), minimal_serialization_settings, task_with_env_vars)
    assert ts.template.container.image == "docker.io/xyz:some-git-hash"
    assert ts.template.container.env == {"HAM": "spam"}


task_with_pod_template = DummyAutoContainerTask(
    name="x",
    metadata=TaskMetadata(
        pod_template_name="podTemplateB",  # should be overwritten
        retries=3,  # ensure other fields still exists
    ),
    task_config=None,
    task_type="t",
    container_image="repo/image:0.0.0",
    requests=Resources(cpu="3", gpu="1"),
    limits=Resources(cpu="6", gpu="2"),
    environment={"eKeyA": "eValA", "eKeyB": "vKeyB"},
    pod_template=PodTemplate(
        primary_container_name="primary",
        labels={"lKeyA": "lValA", "lKeyB": "lValB"},
        annotations={"aKeyA": "aValA", "aKeyB": "aValB"},
        pod_spec=V1PodSpec(
            containers=[
                V1Container(
                    name="notPrimary",
                ),
                V1Container(
                    name="primary",
                    image="repo/primaryImage:0.0.0",
                    command="placeholderCommand",
                    args="placeholderArgs",
                    resources=V1ResourceRequirements(limits={"cpu": "999", "gpu": "999"}),
                    env=[V1EnvVar(name="eKeyC", value="eValC"), V1EnvVar(name="eKeyD", value="eValD")],
                ),
            ],
            volumes=[V1Volume(name="volume")],
        ),
    ),
    pod_template_name="podTemplateA",
)


def test_pod_template(default_serialization_settings):
    #################
    # Test get_k8s_pod
    #################

    container = task_with_pod_template.get_container(default_serialization_settings)
    assert container is None

    k8s_pod = task_with_pod_template.get_k8s_pod(default_serialization_settings)

    # labels/annotations should be passed
    metadata = k8s_pod.metadata
    assert metadata.labels == {"lKeyA": "lValA", "lKeyB": "lValB"}
    assert metadata.annotations == {"aKeyA": "aValA", "aKeyB": "aValB"}

    pod_spec = k8s_pod.pod_spec
    primary_container = pod_spec["containers"][1]

    # To test overwritten attributes

    # image
    assert primary_container["image"] == "repo/primaryImage:0.0.0"
    # command
    assert primary_container["command"] == []
    # args
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
        "tests.flytekit.unit.core.test_python_auto_container",
        "task-name",
        "task_with_pod_template",
    ]
    # resource
    assert primary_container["resources"]["requests"] == {"cpu": "3", "gpu": "1"}
    assert primary_container["resources"]["limits"] == {"cpu": "6", "gpu": "2"}

    # To test union attributes
    assert primary_container["env"] == [
        {"name": "FOO", "value": "bar"},
        {"name": "eKeyA", "value": "eValA"},
        {"name": "eKeyB", "value": "vKeyB"},
        {"name": "eKeyC", "value": "eValC"},
        {"name": "eKeyD", "value": "eValD"},
    ]

    # To test not overwritten attributes
    assert pod_spec["volumes"][0] == {"name": "volume"}

    #################
    # Test pod_template_name
    #################
    assert task_with_pod_template.metadata.pod_template_name == "podTemplateA"
    assert task_with_pod_template.metadata.retries == 3

    config = task_with_minimum_pod_template.get_config(default_serialization_settings)

    #################
    # Test config
    #################
    assert config == {"primary_container_name": "primary"}

    #################
    # Test Serialization
    #################
    ts = get_serializable_task(OrderedDict(), default_serialization_settings, task_with_pod_template)
    assert ts.template.container is None
    # k8s_pod content is already verified above, so only check the existence here
    assert ts.template.k8s_pod is not None

    assert ts.template.metadata.pod_template_name == "podTemplateA"
    assert ts.template.metadata.retries.retries == 3
    assert ts.template.config is not None


task_with_minimum_pod_template = DummyAutoContainerTask(
    name="x",
    task_config=None,
    task_type="t",
    container_image="repo/image:0.0.0",
    pod_template=PodTemplate(
        primary_container_name="primary",
        labels={"lKeyA": "lValA"},
        annotations={"aKeyA": "aValA"},
    ),
    pod_template_name="A",
)


def test_minimum_pod_template(default_serialization_settings):
    #################
    # Test get_k8s_pod
    #################

    container = task_with_minimum_pod_template.get_container(default_serialization_settings)
    assert container is None

    k8s_pod = task_with_minimum_pod_template.get_k8s_pod(default_serialization_settings)

    metadata = k8s_pod.metadata
    assert metadata.labels == {"lKeyA": "lValA"}
    assert metadata.annotations == {"aKeyA": "aValA"}

    pod_spec = k8s_pod.pod_spec
    primary_container = pod_spec["containers"][0]

    assert primary_container["image"] == "repo/image:0.0.0"
    assert primary_container["command"] == []
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
        "tests.flytekit.unit.core.test_python_auto_container",
        "task-name",
        "task_with_minimum_pod_template",
    ]

    config = task_with_minimum_pod_template.get_config(default_serialization_settings)
    assert config == {"primary_container_name": "primary"}

    #################
    # Test pod_teamplte_name
    #################
    assert task_with_minimum_pod_template.metadata.pod_template_name == "A"

    #################
    # Test Serialization
    #################
    ts = get_serializable_task(OrderedDict(), default_serialization_settings, task_with_minimum_pod_template)
    assert ts.template.container is None
    # k8s_pod content is already verified above, so only check the existence here
    assert ts.template.k8s_pod is not None
    assert ts.template.metadata.pod_template_name == "A"
    assert ts.template.config is not None


image_spec_1 = ImageSpec(
    name="image-1",
    packages=["numpy"],
    registry="localhost:30000",
    builder="test",
)

image_spec_2 = ImageSpec(
    name="image-2",
    packages=["pandas"],
    registry="localhost:30000",
    builder="test",
)


ps = V1PodSpec(
    containers=[
        V1Container(
            name="primary",
            image=image_spec_1,
        ),
        V1Container(
            name="secondary",
            image=image_spec_2,
            # use 1 cpu and 1Gi mem
            resources=V1ResourceRequirements(
                requests={"cpu": "1", "memory": "100Mi"},
                limits={"cpu": "1", "memory": "100Mi"},
            ),
        ),
    ]
)

pt = PodTemplate(pod_spec=ps, primary_container_name="primary")


image_spec_task = DummyAutoContainerTask(
    name="x",
    pod_template=pt,
    task_config=None,
    task_type="t",
)


def test_pod_template_with_image_spec(default_serialization_settings, mock_image_spec_builder):
    ImageBuildEngine.register("test", mock_image_spec_builder)

    pod = image_spec_task.get_k8s_pod(default_serialization_settings)
    assert pod.pod_spec["containers"][0]["image"] == image_spec_1.image_name()
    assert pod.pod_spec["containers"][1]["image"] == image_spec_2.image_name()
