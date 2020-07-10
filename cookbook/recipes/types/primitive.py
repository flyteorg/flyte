from __future__ import absolute_import

from flytekit.sdk.tasks import inputs, outputs, python_task
from flytekit.sdk.types import Types
from flytekit.sdk.workflow import workflow_class, Input, Output


@inputs(x=Types.Integer, y=Types.Float)
@outputs(z=Types.Float)
@python_task
def multiply(wf_params, x, y, z):
    z.set(x * y)


@inputs(z=Types.Float)
@outputs(s=Types.String)
@python_task
def convert_to_str(wf_params, z, s):
    s.set("{}".format(z))


@inputs(st=Types.String, b=Types.Boolean)
@outputs(s=Types.String)
@python_task()
def add_bool_to_str(wf_params, st, b, s):
    s.set("{}: {}".format(st, b))


@workflow_class
class PrimitiveDemoWorkflow(object):
    x = Input(Types.Integer, help="Integer")
    y = Input(Types.Float, help="Float")
    s = Input(Types.String, help="String")
    b = Input(Types.Boolean, help="Boolean")

    m = multiply(x=x, y=y)
    s1 = convert_to_str(z=m.outputs.z)
    s2 = add_bool_to_str(st=s, b=b)

    mult_str = Output(s1.outputs.s, sdk_type=Types.String)
    bool_str = Output(s2.outputs.s, sdk_type=Types.String)
