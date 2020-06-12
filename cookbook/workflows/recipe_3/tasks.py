from __future__ import absolute_import
from __future__ import print_function

from flytekit.sdk.tasks import inputs, outputs, python_task
from flytekit.sdk.types import Types
from flytekit.sdk.workflow import workflow_class, Input, Output


@inputs(custom=Types.Generic)
@outputs(counts=Types.Generic)
@python_task
def generic_type_task(wf_params, custom, counts):
    # Go through each of the values of the input and if it's a str, count the length
    wf_params.logging.info("Running custom object task")
    results = {}
    for k, v in custom.items():
        if type(v) == str:
            results[k] = len(v)
        else:
            results[k] = v

    counts.set(results)


@workflow_class
class GenericDemoWorkflow(object):
    a = Input(Types.Generic, default={}, help="Input for inner workflow")
    generic_type_example = generic_type_task(custom=a)
    counts = Output(generic_type_example.outputs.counts, sdk_type=Types.Generic)
