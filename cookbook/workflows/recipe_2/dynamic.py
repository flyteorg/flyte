from __future__ import absolute_import
from __future__ import print_function

from flytekit.sdk.tasks import inputs, outputs, dynamic_task
from flytekit.sdk.types import Types
from flytekit.sdk.workflow import workflow_class, Input, Output, workflow

from workflows.recipe_2.tasks import inverse_inner_task, inner_task, sq_sub_task


def manual_assign_name():
    pass


@inputs(task_input_num=Types.Integer)
@outputs(out=Types.Integer)
@dynamic_task
def dynamic_wf_task(wf_params, task_input_num, out):
    wf_params.logging.info("Running dynamic task... yielding a code generated sub workflow")

    input_a = Input(Types.Integer, help="Tell me something")
    node1 = sq_sub_task(in1=input_a)

    MyUnregisteredWorkflow = workflow(
        inputs={
            'a': input_a,
        },
        outputs={
            'ooo': Output(node1.outputs.out1, sdk_type=Types.Integer,
                          help='This is an integer output')
        },
        nodes={
            'node_one': node1,
        }
    )

    # This is an unfortunate setting that will hopefully not be necessary in the future.
    setattr(MyUnregisteredWorkflow, 'auto_assign_name', manual_assign_name)
    MyUnregisteredWorkflow._platform_valid_name = 'unregistered'

    unregistered_workflow_execution = MyUnregisteredWorkflow(a=task_input_num)
    out.set(unregistered_workflow_execution.outputs.ooo)


@workflow_class
class SimpleDynamicSubworkflow(object):
    input_a = Input(Types.Integer, default=5, help="Input for inner workflow")
    lp_task = dynamic_wf_task(task_input_num=input_a)
    wf_output = Output(lp_task.outputs.out, sdk_type=Types.Integer)


@inputs(task_input_num=Types.Integer)
@inputs(decider=Types.Boolean)
@outputs(out=Types.Integer)
@dynamic_task
def workflow_builder(wf_params, task_input_num, decider, out):
    wf_params.logging.info("Running inner task... yielding a code generated sub workflow")

    input_a = Input(Types.Integer, help="Tell me something")
    if decider:
        node1 = inverse_inner_task(num=input_a)
    else:
        node1 = inner_task(num=input_a)

    MyUnregisteredWorkflow = workflow(
        inputs={
            'a': input_a,
        },
        outputs={
            'ooo': Output(node1.outputs.out, sdk_type=Types.Integer, help='This is an integer output')
        },
        nodes={
            'node_one': node1,
        }
    )

    # This is an unfortunate setting that will hopefully not be necessary in the future.
    setattr(MyUnregisteredWorkflow, 'auto_assign_name', manual_assign_name)
    MyUnregisteredWorkflow._platform_valid_name = 'unregistered'

    unregistered_workflow_execution = MyUnregisteredWorkflow(a=task_input_num)

    yield unregistered_workflow_execution
    out.set(unregistered_workflow_execution.outputs.ooo)


@workflow_class
class InverterDynamicWorkflow(object):
    input_a = Input(Types.Integer, default=5, help="Input for inner workflow")
    inverter_input = Input(Types.Boolean, default=False, help="Should invert or not")
    lp_task = workflow_builder(task_input_num=input_a, decider=inverter_input)
    wf_output = Output(lp_task.outputs.out, sdk_type=Types.Integer)


@inputs(task_input_num=Types.Integer)
@outputs(out=Types.Integer)
@dynamic_task
def nested_dynamic_wf_task(wf_params, task_input_num, out):
    wf_params.logging.info("Running inner task... yielding a code generated sub workflow")

    # Inner workflow
    input_a = Input(Types.Integer, help="Tell me something")
    node1 = sq_sub_task(in1=input_a)

    MyUnregisteredWorkflowInner = workflow(
        inputs={
            'a': input_a,
        },
        outputs={
            'ooo': Output(node1.outputs.out1, sdk_type=Types.Integer,
                          help='This is an integer output')
        },
        nodes={
            'node_one': node1,
        }
    )

    setattr(MyUnregisteredWorkflowInner, 'auto_assign_name', manual_assign_name)
    MyUnregisteredWorkflowInner._platform_valid_name = 'unregistered'

    # Output workflow
    input_a = Input(Types.Integer, help="Tell me something")
    node1 = MyUnregisteredWorkflowInner(a=task_input_num)

    MyUnregisteredWorkflowOuter = workflow(
        inputs={
            'a': input_a,
        },
        outputs={
            'ooo': Output(node1.outputs.ooo, sdk_type=Types.Integer,
                          help='This is an integer output')
        },
        nodes={
            'node_one': node1,
        }
    )



    setattr(MyUnregisteredWorkflowOuter, 'auto_assign_name', manual_assign_name)
    MyUnregisteredWorkflowOuter._platform_valid_name = 'unregistered'

    unregistered_workflow_execution = MyUnregisteredWorkflowOuter(a=task_input_num)
    out.set(unregistered_workflow_execution.outputs.ooo)

