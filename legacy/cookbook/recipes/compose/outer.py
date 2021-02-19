from __future__ import absolute_import
from __future__ import print_function

from flytekit.sdk.tasks import inputs, outputs, dynamic_task
from flytekit.sdk.types import Types
from flytekit.sdk.workflow import workflow_class, Input, Output

from compose.inner import IdentityWorkflow, secondary_sibling_identity_lp


@workflow_class()
class StaticSubWorkflowCaller(object):
    outer_a = Input(Types.Integer, default=5, help="Input for inner workflow")
    identity_wf_execution = IdentityWorkflow(a=outer_a)
    wf_output = Output(identity_wf_execution.outputs.task_output, sdk_type=Types.Integer)


# Note that this causes a duplicate copy of the default launch plan effectively, stored with the name id_lp
id_lp = IdentityWorkflow.create_launch_plan()

# Fetch a specific launch plan that's already been registered with Flyte Admin. Note that when you run registration on
# this file, unlike the first example, this will not be registered, no copy will be made.
# Also, using a fetch call like this is a bit of an anti-pattern, since it requires access to Flyte control plane
# from within a running task, something we try to avoid.
# from flytekit.common.launch_plan import SdkLaunchPlan
# fetched_identity_lp = SdkLaunchPlan.fetch('flyteeexamples', 'development',
#                               'cookbook.sample_workflows.formula_1.inner.IdentityWorkflow',
#                               '8d291bdf163674dcb6ea8a047a4de6cc7cf4853f')


@workflow_class()
class StaticLaunchPlanCaller(object):
    outer_a = Input(Types.Integer, default=5, help="Input for inner workflow")
    identity_lp_execution = id_lp(a=outer_a)
    wf_output = Output(identity_lp_execution.outputs.task_output, sdk_type=Types.Integer)


@inputs(num=Types.Integer)
@outputs(out=Types.Integer, imported_output=Types.Integer)
@dynamic_task
def lp_yield_task(wf_params, num, out, imported_output):
    wf_params.logging.info("Running inner task... yielding a launchplan")
    wf_params.logging.info("{} {}".format(id_lp._id, id_lp.workflow_id))

    # Test one launch plan defined in this file, but from a workflow imported from another file
    identity_lp_execution = id_lp(a=num)
    yield identity_lp_execution
    out.set(identity_lp_execution.outputs.task_output)

    # Test another defined in another file and just imported.
    imported_lp_execution = secondary_sibling_identity_lp()
    yield imported_lp_execution
    imported_output.set(imported_lp_execution.outputs.task_output)

    # Test a launch plan that's fetched
    # fetched_lp_execution = fetched_identity_lp(a=15)
    # yield fetched_lp_execution
    # fetched_output.set(fetched_lp_execution.outputs.task_output)

    # Note that a launch plan created here, like SomeWorkflow.create_launch_plan() would never work, because
    # it will never have been registered with the Flyte control plane (because the body of tasks like this one do
    # not run at registration time).


@workflow_class
class DynamicLaunchPlanCaller(object):
    outer_a = Input(Types.Integer, default=5, help="Input for inner workflow")
    lp_task = lp_yield_task(num=outer_a)
    wf_output = Output(lp_task.outputs.out, sdk_type=Types.Integer)


@inputs(num=Types.Integer)
@outputs(out=Types.Integer)
@dynamic_task
def sub_wf_yield_task(wf_params, num, out):
    wf_params.logging.info("Running inner task... yielding a sub-workflow")
    identity_wf_execution = IdentityWorkflow(a=num)
    yield identity_wf_execution
    out.set(identity_wf_execution.outputs.task_output)


@workflow_class
class DynamicSubWorkflowCaller(object):
    outer_a = Input(Types.Integer, default=5, help="Input for inner workflow")
    sub_wf_task = sub_wf_yield_task(num=outer_a)
    wf_output = Output(sub_wf_task.outputs.out, sdk_type=Types.Integer)
