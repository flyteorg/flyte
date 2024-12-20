from collections import OrderedDict

import pytest
from kubernetes.client.models import (
    V1Affinity,
    V1Container,
    V1NodeAffinity,
    V1NodeSelectorRequirement,
    V1NodeSelectorTerm,
    V1PodSpec,
    V1PreferredSchedulingTerm,
    V1ResourceRequirements,
    V1Toleration,
)

from flytekit import kwtypes
from flytekit.configuration import Image, ImageConfig, SerializationSettings
from flytekit.core.container_task import ContainerTask
from flytekit.core.pod_template import PodTemplate
from flytekit.image_spec.image_spec import ImageBuildEngine, ImageSpec
from flytekit.tools.translator import get_serializable_task


def test_pod_template():
    ps = V1PodSpec(
        containers=[], tolerations=[V1Toleration(effect="NoSchedule", key="nvidia.com/gpu", operator="Exists")]
    )
    ps.runtime_class_name = "nvidia"
    nsr = V1NodeSelectorRequirement(key="nvidia.com/gpu.memory", operator="Gt", values=["10000"])
    pref_sched = V1PreferredSchedulingTerm(preference=V1NodeSelectorTerm(match_expressions=[nsr]), weight=1)
    ps.affinity = V1Affinity(
        node_affinity=V1NodeAffinity(preferred_during_scheduling_ignored_during_execution=[pref_sched])
    )
    pt = PodTemplate(pod_spec=ps, labels={"somelabel": "foobar"})

    image = "ghcr.io/flyteorg/rawcontainers-shell:v2"
    cmd = [
        "./calculate-ellipse-area.sh",
        "{{.inputs.a}}",
        "{{.inputs.b}}",
        "/var/outputs",
    ]
    ct = ContainerTask(
        name="ellipse-area-metadata-shell",
        input_data_dir="/var/inputs",
        output_data_dir="/var/outputs",
        inputs=kwtypes(a=float, b=float),
        outputs=kwtypes(area=float, metadata=str),
        image=image,
        command=cmd,
        pod_template=pt,
        pod_template_name="my-base-template",
    )

    assert ct.metadata.pod_template_name == "my-base-template"

    default_image = Image(name="default", fqn="docker.io/xyz", tag="some-git-hash")
    default_image_config = ImageConfig(default_image=default_image)
    default_serialization_settings = SerializationSettings(
        project="p", domain="d", version="v", image_config=default_image_config
    )

    container = ct.get_container(default_serialization_settings)
    assert container is None

    k8s_pod = ct.get_k8s_pod(default_serialization_settings)
    assert k8s_pod.metadata.labels == {"somelabel": "foobar"}

    primary_container = k8s_pod.pod_spec["containers"][0]

    assert primary_container["image"] == image
    assert primary_container["command"] == cmd

    #################
    # Test Serialization
    #################
    ts = get_serializable_task(OrderedDict(), default_serialization_settings, ct)
    assert ts.template.metadata.pod_template_name == "my-base-template"
    assert ts.template.container is None
    assert ts.template.k8s_pod is not None
    serialized_pod_spec = ts.template.k8s_pod.pod_spec
    assert serialized_pod_spec["affinity"]["nodeAffinity"] is not None
    assert serialized_pod_spec["tolerations"] == [
        {"effect": "NoSchedule", "key": "nvidia.com/gpu", "operator": "Exists"}
    ]
    assert serialized_pod_spec["runtimeClassName"] == "nvidia"


def test_local_execution():
    ct = ContainerTask(
        name="name",
        input_data_dir="/var/inputs",
        output_data_dir="/var/outputs",
        image="inexistent-image:v42",
        command=["some", "command"],
    )

    with pytest.raises(RuntimeError):
        ct()


def test_raw_container_with_image_spec(mock_image_spec_builder):
    ImageBuildEngine.register("test-raw-container", mock_image_spec_builder)
    image_spec = ImageSpec(registry="flyte", base_image="r-base", builder="test-raw-container")

    calculate_ellipse_area_r = ContainerTask(
        name="ellipse-area-metadata-r",
        input_data_dir="/var/inputs",
        output_data_dir="/var/outputs",
        inputs=kwtypes(a=float, b=float),
        outputs=kwtypes(area=float, metadata=str),
        image=image_spec,
        command=[
            "Rscript",
            "--vanilla",
            "/root/calculate-ellipse-area.R",
            "{{.inputs.a}}",
            "{{.inputs.b}}",
            "/var/outputs",
        ],
    )

    default_serialization_settings = SerializationSettings(
        project="p", domain="d", version="v", image_config=ImageConfig.auto()
    )
    container = calculate_ellipse_area_r.get_container(default_serialization_settings)
    assert container.image == image_spec.image_name()


def test_container_task_image_spec(mock_image_spec_builder):
    default_image = Image(name="default", fqn="docker.io/xyz", tag="some-git-hash")
    default_image_config = ImageConfig(default_image=default_image)

    default_serialization_settings = SerializationSettings(
        project="p", domain="d", version="v", image_config=default_image_config, env={"FOO": "bar"}
    )

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

    ct = ContainerTask(
        name="x",
        image="ddd",
        command=["ccc"],
        pod_template=pt,
    )
    ImageBuildEngine.register("test", mock_image_spec_builder)
    pod = ct.get_k8s_pod(default_serialization_settings)
    assert pod.pod_spec["containers"][0]["image"] == image_spec_1.image_name()
    assert pod.pod_spec["containers"][1]["image"] == image_spec_2.image_name()
