from flytekit.sdk.tasks import dynamic_sidecar_task, inputs, outputs, python_task, sidecar_task
from flytekit.sdk.types import Types
from flytekit.sdk.workflow import workflow_class, Input
from k8s.io.api.core.v1 import generated_pb2


def generate_simple_pod_spec_for_task():
    pod_spec = generated_pb2.PodSpec()
    primary_container = generated_pb2.Container(name="primary")
    pod_spec.containers.extend([primary_container])
    return pod_spec


@inputs(index=Types.Integer)
@outputs(output=Types.String)
@python_task
def individual_python_task(wf_params, index, output):
    output.set("I'm task #{}".format(index))


@inputs(tasks_count=Types.Integer)
@outputs(assembled_output=[Types.String])
@dynamic_sidecar_task(
    pod_spec=generate_simple_pod_spec_for_task(),
    primary_container_name="primary",
    cpu_limit='1',
    cpu_request='1',
    memory_limit='100Mi',
    memory_request='100Mi',
)
def dynamic_sidecar_task(wf_params, tasks_count, assembled_output):
    result = []
    for i in range(tasks_count):
        individual_task = individual_python_task(index=i)
        yield individual_task
        result.append(individual_task.outputs.output)
    assembled_output.set(result)


@workflow_class
class DynamicSidecarWorkflow(object):
    tasks_count = Input(Types.Integer, required=True, help="The number of dynamic python tasks to run")
    s = dynamic_sidecar_task(tasks_count=tasks_count)
