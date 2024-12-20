from collections import OrderedDict

import pytest
from kubernetes.client.models import V1Container, V1PodSpec

from flytekit import task
from flytekit.configuration import Image, ImageConfig, SerializationSettings
from flytekit.core.pod_template import PodTemplate
from flytekit.core.python_auto_container import get_registerable_container_image
from flytekit.core.python_function_task import PythonFunctionTask
from flytekit.core.tracker import isnested, istestfunction
from flytekit.image_spec.image_spec import ImageBuildEngine, ImageSpec
from flytekit.tools.translator import get_serializable_task
from tests.flytekit.unit.core import tasks


def foo():
    pass


def test_isnested():
    def inner_foo():
        pass

    assert isnested(foo) is False
    assert isnested(inner_foo) is True

    # Uses tasks.tasks method
    with pytest.raises(ValueError):
        tasks.tasks()


def test_istestfunction():
    assert istestfunction(foo) is True
    assert istestfunction(isnested) is False
    assert istestfunction(tasks.tasks) is False


def test_container_image_conversion(mock_image_spec_builder):
    default_img = Image(name="default", fqn="xyz.com/abc", tag="tag1")
    other_img = Image(name="other", fqn="xyz.com/other", tag="tag-other")
    cfg = ImageConfig(default_image=default_img, images=[default_img, other_img])
    assert get_registerable_container_image(None, cfg) == "xyz.com/abc:tag1"
    assert get_registerable_container_image("", cfg) == "xyz.com/abc:tag1"
    assert get_registerable_container_image("abc", cfg) == "abc"
    assert get_registerable_container_image("abc:latest", cfg) == "abc:latest"
    assert get_registerable_container_image("abc:{{.image.default.version}}", cfg) == "abc:tag1"
    assert (
        get_registerable_container_image("{{.image.default.fqn}}:{{.image.default.version}}", cfg) == "xyz.com/abc:tag1"
    )
    assert (
        get_registerable_container_image("{{.image.other.fqn}}:{{.image.other.version}}", cfg)
        == "xyz.com/other:tag-other"
    )
    assert (
        get_registerable_container_image("{{.image.other.fqn}}:{{.image.default.version}}", cfg) == "xyz.com/other:tag1"
    )
    assert get_registerable_container_image("{{.image.other.fqn}}", cfg) == "xyz.com/other"
    # Works with images instead of just image
    assert get_registerable_container_image("{{.images.other.fqn}}", cfg) == "xyz.com/other"

    with pytest.raises(AssertionError):
        get_registerable_container_image("{{.image.blah.fqn}}:{{.image.other.version}}", cfg)

    with pytest.raises(AssertionError):
        get_registerable_container_image("{{.image.fqn}}:{{.image.other.version}}", cfg)

    with pytest.raises(AssertionError):
        get_registerable_container_image("{{.image.blah}}", cfg)

    assert get_registerable_container_image("{{.image.default}}", cfg) == "xyz.com/abc:tag1"

    ImageBuildEngine.register("test", mock_image_spec_builder)
    image_spec = ImageSpec(builder="test", python_version="3.7", registry="")
    assert get_registerable_container_image(image_spec, cfg) == image_spec.image_name()


def test_get_registerable_container_image_no_images():
    cfg = ImageConfig()

    with pytest.raises(ValueError):
        get_registerable_container_image("", cfg)


def test_py_func_task_get_container():
    def foo(i: int):
        pass

    default_img = Image(name="default", fqn="xyz.com/abc", tag="tag1")
    other_img = Image(name="other", fqn="xyz.com/other", tag="tag-other")
    cfg = ImageConfig(default_image=default_img, images=[default_img, other_img])

    settings = SerializationSettings(project="p", domain="d", version="v", image_config=cfg, env={"FOO": "bar"})

    pytask = PythonFunctionTask(None, foo, None, environment={"BAZ": "baz"})
    c = pytask.get_container(settings)
    assert c.image == "xyz.com/abc:tag1"
    assert c.env == {"FOO": "bar", "BAZ": "baz"}


def test_metadata():
    # test cache, cache_serialize, and cache_version are correctly set
    @task(cache=True, cache_serialize=True, cache_version="1.0")
    def foo(i: str):
        print(f"{i}")

    foo_metadata = foo.metadata
    assert foo_metadata.cache is True
    assert foo_metadata.cache_serialize is True
    assert foo_metadata.cache_version == "1.0"

    # test cache, cache_serialize, and cache_version at no unnecessarily set
    @task()
    def bar(i: str):
        print(f"{i}")

    bar_metadata = bar.metadata
    assert bar_metadata.cache is False
    assert bar_metadata.cache_serialize is False
    assert bar_metadata.cache_version == ""

    # test missing cache_version
    with pytest.raises(ValueError):

        @task(cache=True)
        def foo_missing_cache_version(i: str):
            print(f"{i}")

    # test missing cache
    with pytest.raises(ValueError):

        @task(cache_serialize=True)
        def foo_missing_cache(i: str):
            print(f"{i}")


def test_pod_template():
    @task(
        container_image="repo/image:0.0.0",
        pod_template=PodTemplate(
            primary_container_name="primary",
            labels={"lKeyA": "lValA"},
            annotations={"aKeyA": "aValA"},
            pod_spec=V1PodSpec(
                containers=[
                    V1Container(
                        name="primary",
                    ),
                ]
            ),
        ),
        pod_template_name="A",
    )
    def func_with_pod_template(i: str):
        print(i + "a")

    default_image = Image(name="default", fqn="docker.io/xyz", tag="some-git-hash")
    default_image_config = ImageConfig(default_image=default_image)
    default_serialization_settings = SerializationSettings(
        project="p", domain="d", version="v", image_config=default_image_config
    )

    #################
    # Test get_k8s_pod
    #################

    container = func_with_pod_template.get_container(default_serialization_settings)
    assert container is None

    k8s_pod = func_with_pod_template.get_k8s_pod(default_serialization_settings)

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
        "tests.flytekit.unit.core.test_python_function_task",
        "task-name",
        "func_with_pod_template",
    ]

    #################
    # Test pod_template_name
    #################
    assert func_with_pod_template.metadata.pod_template_name == "A"

    #################
    # Test Serialization
    #################
    ts = get_serializable_task(OrderedDict(), default_serialization_settings, func_with_pod_template)
    assert ts.template.container is None
    # k8s_pod content is already verified above, so only check the existence here
    assert ts.template.k8s_pod is not None
    assert ts.template.metadata.pod_template_name == "A"


def test_node_dependency_hints_are_not_allowed():
    @task
    def t1(i: str):
        pass

    with pytest.raises(ValueError, match="node_dependency_hints should only be used on dynamic tasks"):

        @task(node_dependency_hints=[t1])
        def t2(i: str):
            pass
