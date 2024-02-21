import pytest
from mock import patch as _system_patch

import flytekit.configuration
from flytekit.configuration import Image, ImageConfig
from flytekit.core.task import task
from flytekit.core.testing import patch as flyte_patch
from flytekit.core.workflow import ImperativeWorkflow, workflow

default_img = Image(name="default", fqn="test", tag="tag")
serialization_settings = flytekit.configuration.SerializationSettings(
    project="project",
    domain="domain",
    version="version",
    env=None,
    image_config=ImageConfig(default_image=default_img, images=[default_img]),
)


@task
def t1(a: str) -> str:
    return a + " world"


wb = ImperativeWorkflow(name="my.workflow")
wb.add_workflow_input("in1", str)
node = wb.add_entity(t1, a=wb.inputs["in1"])
wb.add_workflow_output("from_n0t1", node.outputs["o0"])


def test_base_case():
    assert wb(in1="hello") == "hello world"


# Please see https://github.com/flyteorg/flyte/issues/854 for more information.
# This mock_patch_wf object is a duplicate of the wb object above. Because of the issue 854, we can't
# use the same object.
# TODO: Remove this duplicate object pending resolution of #854
mock_patch_wf = ImperativeWorkflow(name="my.workflow")
mock_patch_wf.add_workflow_input("in1", str)
node = mock_patch_wf.add_entity(t1, a=mock_patch_wf.inputs["in1"])
mock_patch_wf.add_workflow_output("from_n0t1", node.outputs["o0"])


@_system_patch("flytekit.core.workflow.ImperativeWorkflow.execute")
def test_return_none_errors(mock_execute):
    mock_execute.return_value = None
    with pytest.raises(Exception):
        mock_patch_wf(in1="hello")


@flyte_patch(t1)
def test_none_conversion(mock_t1):
    mock_t1.return_value = None
    # This will try to convert None to a string
    with pytest.raises(TypeError):
        wb(in1="hello")


@flyte_patch(wb)
def test_imperative_patching(mock_wb):
    mock_wb.return_value = "hi"

    @workflow
    def my_functional_wf(a: str) -> str:
        x = wb(in1=a)
        return x

    assert my_functional_wf(a="hello") == "hi"
