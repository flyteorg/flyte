import tempfile
from collections import OrderedDict

import mock

import flytekit.configuration
from flytekit import ContainerTask, Resources, kwtypes
from flytekit.configuration import Image, ImageConfig
from flytekit.core.python_customized_container_task import PythonCustomizedContainerTask, TaskTemplateResolver
from flytekit.core.utils import write_proto_to_file
from flytekit.tools.translator import get_serializable

default_img = Image(name="default", fqn="test", tag="tag")
serialization_settings = flytekit.configuration.SerializationSettings(
    project="project",
    domain="domain",
    version="version",
    env=None,
    image_config=ImageConfig(default_image=default_img, images=[default_img]),
)


class Placeholder(object):
    ...


def test_resolver_load_task():
    # any task is fine, just copied one
    square = ContainerTask(
        name="square",
        input_data_dir="/var/inputs",
        output_data_dir="/var/outputs",
        inputs=kwtypes(val=int),
        outputs=kwtypes(out=int),
        image="alpine",
        command=["sh", "-c", "echo $(( {{.Inputs.val}} * {{.Inputs.val}} )) | tee /var/outputs/out"],
    )

    resolver = TaskTemplateResolver()
    ts = get_serializable(OrderedDict(), serialization_settings, square)
    file = tempfile.NamedTemporaryFile().name
    # load_task should create an instance of the path to the object given, doesn't need to be a real executor
    write_proto_to_file(ts.template.to_flyte_idl(), file)
    shim_task = resolver.load_task([file, f"{Placeholder.__module__}.Placeholder"])
    assert isinstance(shim_task.executor, Placeholder)
    assert shim_task.task_template.id.name == "square"
    assert shim_task.task_template.interface.inputs["val"] is not None
    assert shim_task.task_template.interface.outputs["out"] is not None


@mock.patch("flytekit.core.python_customized_container_task.PythonCustomizedContainerTask.get_config")
@mock.patch("flytekit.core.python_customized_container_task.PythonCustomizedContainerTask.get_custom")
def test_serialize_to_model(mock_custom, mock_config):
    mock_custom.return_value = {"a": "custom"}
    mock_config.return_value = {"a": "config"}
    ct = PythonCustomizedContainerTask(
        name="mytest",
        task_config=None,
        container_image="someimage",
        executor_type=Placeholder,
        requests=Resources(ephemeral_storage="200Mi"),
        limits=Resources(ephemeral_storage="300Mi"),
    )
    tt = ct.serialize_to_model(serialization_settings)
    assert tt.container.image == "someimage"
    assert len(tt.config) == 1
    assert tt.id.name == "mytest"
    assert len(tt.custom) == 1
    assert tt.container.resources.requests[0].value == "200Mi"
    assert tt.container.resources.limits[0].value == "300Mi"
