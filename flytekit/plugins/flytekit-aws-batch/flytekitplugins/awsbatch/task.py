from dataclasses import dataclass
from typing import Any, Callable, Dict, List, Optional

from dataclasses_json import DataClassJsonMixin
from google.protobuf import json_format
from google.protobuf.struct_pb2 import Struct

from flytekit import PythonFunctionTask
from flytekit.configuration import SerializationSettings
from flytekit.extend import TaskPlugins


@dataclass
class AWSBatchConfig(DataClassJsonMixin):
    """
    Use this to configure SubmitJobInput for a AWS batch job. Task's marked with this will automatically execute
    natively onto AWS batch service.
    Refer to AWS SubmitJobInput for more detail: https://docs.aws.amazon.com/sdk-for-go/api/service/batch/#SubmitJobInput
    """

    parameters: Optional[Dict[str, str]] = None
    schedulingPriority: Optional[int] = None
    platformCapabilities: str = "EC2"
    propagateTags: Optional[bool] = None
    tags: Optional[Dict[str, str]] = None

    def to_dict(self):
        s = Struct()
        s.update(super().to_dict())
        return json_format.MessageToDict(s)


class AWSBatchFunctionTask(PythonFunctionTask):
    """
    Actual Plugin that transforms the local python code for execution within AWS batch job
    """

    _AWS_BATCH_TASK_TYPE = "aws-batch"

    def __init__(self, task_config: AWSBatchConfig, task_function: Callable, **kwargs):
        if task_config is None:
            task_config = AWSBatchConfig()
        super(AWSBatchFunctionTask, self).__init__(
            task_config=task_config, task_type=self._AWS_BATCH_TASK_TYPE, task_function=task_function, **kwargs
        )
        self._task_config = task_config

    def get_custom(self, settings: SerializationSettings) -> Dict[str, Any]:
        # task_config will be used to create SubmitJobInput in propeller except platformCapabilities.
        return self._task_config.to_dict()

    def get_config(self, settings: SerializationSettings) -> Dict[str, str]:
        # Parameters in taskTemplate config will be used to create aws job definition.
        # More detail about job definition: https://docs.aws.amazon.com/batch/latest/userguide/job_definition_parameters.html
        return {**super().get_config(settings), "platformCapabilities": self._task_config.platformCapabilities}

    def get_command(self, settings: SerializationSettings) -> List[str]:
        container_args = [
            "pyflyte-execute",
            "--inputs",
            "{{.input}}",
            "--output-prefix",
            # As of FlytePropeller v0.16.28, aws array batch plugin support to run single job.
            # This task will call aws batch plugin to execute the task on aws batch service.
            # For single job, FlytePropeller will always read the output from this directory (outputPrefix/0)
            # More detail, see https://github.com/flyteorg/flyteplugins/blob/0dd93c23ed2edeca65d58e89b0edb613f88120e0/go/tasks/plugins/array/catalog.go#L501.
            "{{.outputPrefix}}/0",
            "--raw-output-data-prefix",
            "{{.rawOutputDataPrefix}}",
            "--resolver",
            self.task_resolver.location,
            "--",
            *self.task_resolver.loader_args(settings, self),
        ]

        return container_args


# Inject the AWS batch plugin into flytekits dynamic plugin loading system
TaskPlugins.register_pythontask_plugin(AWSBatchConfig, AWSBatchFunctionTask)
