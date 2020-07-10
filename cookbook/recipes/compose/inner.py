from __future__ import absolute_import
from __future__ import print_function

from flytekit.sdk.tasks import inputs, outputs, python_task
from flytekit.sdk.types import Types
from flytekit.sdk.workflow import workflow_class, Input, Output


@inputs(num=Types.Integer)
@outputs(out=Types.Integer)
@python_task
def inner_task(wf_params, num, out):
    wf_params.logging.info("Running inner task... setting output to input")
    out.set(num)


@workflow_class
class IdentityWorkflow(object):
    a = Input(Types.Integer, default=5, help="Input for inner workflow")
    odd_nums_task = inner_task(num=a)
    task_output = Output(odd_nums_task.outputs.out, sdk_type=Types.Integer)


# Create a custom launch plan with a different default input
secondary_sibling_identity_lp = IdentityWorkflow.create_launch_plan(
    default_inputs={'a': Input(Types.Integer, default=42)}
)
