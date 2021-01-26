from flytekit.sdk.tasks import inputs, outputs, python_task
from flytekit.sdk.types import Types
from flytekit.sdk.workflow import workflow_class, Input, Output
import json


@inputs(custom=Types.Generic)
@outputs(counts=Types.Generic, replicated=Types.List(Types.Generic))
@python_task
def generic_type_task(wf_params, custom, counts, replicated):
    """
    Go through each of the values of the input and if it's a str, count the length
    Also, create a replicated list of the Generic
    """
    wf_params.logging.info("Running custom object task")
    results = {}
    for k, v in custom.items():
        if type(v) == str:
            results[k] = len(v)
        else:
            results[k] = v

    counts.set(results)
    replicated.set([custom, custom])

@inputs(replicated=Types.List(Types.Generic))
@outputs(str_repr=Types.String)
@python_task
def generic_to_json(wf_params, replicated, str_repr):
    """
    convert the list of generic to a json string
    """
    wf_params.logging.info("Running conversion task")
    str_repr.set(json.dumps(replicated))


@workflow_class
class GenericDemoWorkflow(object):
    a = Input(Types.Generic, default={}, help="Input for inner workflow")
    generic_type_example = generic_type_task(custom=a)
    generic_json = generic_to_json(replicated=generic_type_example.outputs.replicated)
    counts = Output(generic_type_example.outputs.counts, sdk_type=Types.Generic)
