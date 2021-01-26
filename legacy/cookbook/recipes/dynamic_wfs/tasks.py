from __future__ import absolute_import
from __future__ import print_function

from flytekit.sdk.tasks import inputs, outputs, python_task
from flytekit.sdk.types import Types


@inputs(num=Types.Integer)
@outputs(out=Types.Integer)
@python_task
def inner_task(wf_params, num, out):
    wf_params.logging.info("Running inner task... setting output to input")
    out.set(num)


@inputs(num=Types.Integer)
@outputs(out=Types.Integer)
@python_task
def inverse_inner_task(wf_params, num, out):
    wf_params.logging.info("Running inner task... setting output to inverse of input")
    out.set(-num)


@inputs(in1=Types.Integer)
@outputs(out1=Types.Integer)
@python_task
def sq_sub_task(wf_params, in1, out1):
    wf_params.logging.info("Running inner task... setting output to square of input")
    out1.set(in1 * in1)
