from datetime import timedelta
from os import path as _path

from flyteidl.core import compiler_pb2 as _compiler_pb2
from flyteidl.core import workflow_pb2 as _workflow_pb2

from flytekit.models import literals as _literals
from flytekit.models import task as _task_model
from flytekit.models.core import compiler as _compiler_model
from flytekit.models.core import workflow as _workflow_model


def get_sample_node_metadata(node_id):
    """
    :param Text node_id:
    :rtype: flytekit.models.core.workflow.NodeMetadata
    """

    return _workflow_model.NodeMetadata(name=node_id, timeout=timedelta(seconds=10), retries=_literals.RetryStrategy(0))


def get_sample_container():
    """
    :rtype: flytekit.models.task.Container
    """
    cpu_resource = _task_model.Resources.ResourceEntry(_task_model.Resources.ResourceName.CPU, "1")
    resources = _task_model.Resources(requests=[cpu_resource], limits=[cpu_resource])

    return _task_model.Container(
        "my_image",
        ["this", "is", "a", "cmd"],
        ["this", "is", "an", "arg"],
        resources,
        {},
        {},
    )


def get_sample_task_metadata():
    """
    :rtype: flytekit.models.task.TaskMetadata
    """
    return _task_model.TaskMetadata(
        True,
        _task_model.RuntimeMetadata(_task_model.RuntimeMetadata.RuntimeType.FLYTE_SDK, "1.0.0", "python"),
        timedelta(days=1),
        _literals.RetryStrategy(3),
        True,
        "0.1.1b0",
        "This is deprecated!",
        True,
    )


def get_workflow_template():
    """
    This function retrieves a TasKTemplate object from the pb file in the resources directory.
    It was created by reading from Flyte Admin, the following workflow, after registration.

    from flytekit.common.types.primitives import Integer
    from flytekit.sdk.tasks import (
        python_task,
        inputs,
        outputs,
    )
    from flytekit.sdk.types import Types
    from flytekit.sdk.workflow import workflow_class, Input, Output


    @inputs(a=Types.Integer)
    @outputs(b=Types.Integer, c=Types.Integer)
    @python_task()
    def demo_task_for_promote(wf_params, a, b, c):
        b.set(a + 1)
        c.set(a + 2)


    @workflow_class()
    class OneTaskWFForPromote(object):
        wf_input = Input(Types.Integer, required=True)
        my_task_node = demo_task_for_promote(a=wf_input)
        wf_output_b = Output(my_task_node.outputs.b, sdk_type=Integer)
        wf_output_c = Output(my_task_node.outputs.c, sdk_type=Integer)


    :rtype: flytekit.models.core.workflow.WorkflowTemplate
    """
    workflow_template_pb = _workflow_pb2.WorkflowTemplate()
    # So that tests that use this work when run from any directory
    basepath = _path.dirname(__file__)
    filepath = _path.abspath(_path.join(basepath, "resources/protos", "OneTaskWFForPromote.pb"))
    with open(filepath, "rb") as fh:
        workflow_template_pb.ParseFromString(fh.read())

    wt = _workflow_model.WorkflowTemplate.from_flyte_idl(workflow_template_pb)
    return wt


def get_compiled_workflow_closure():
    """
    :rtype: flytekit.models.core.compiler.CompiledWorkflowClosure
    """
    cwc_pb = _compiler_pb2.CompiledWorkflowClosure()
    # So that tests that use this work when run from any directory
    basepath = _path.dirname(__file__)
    filepath = _path.abspath(_path.join(basepath, "resources/protos", "CompiledWorkflowClosure.pb"))
    with open(filepath, "rb") as fh:
        cwc_pb.ParseFromString(fh.read())

    return _compiler_model.CompiledWorkflowClosure.from_flyte_idl(cwc_pb)
