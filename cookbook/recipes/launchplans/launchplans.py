from flytekit.common import notifications as _notifications
from flytekit.models.common import Labels, Annotations
from flytekit.models.core import execution as _execution

from recipes.workflows import workflows

# This is a launchplan that is created with default values of inputs and added some annotations, labels and
# notification preferences. Default values make it possible to still override when an execution is requested
scale_rotate_default_launchplan = workflows.ScaleAndRotateWorkflow.create_launch_plan(
    labels=Labels({
        'flyte.org/managed': 'true',
    }),
    annotations=Annotations({
        'flyte.org/secret-inject':
            'required',
    }),
    notifications=[
        _notifications.Slack(
            [
                _execution.WorkflowExecutionPhase.SUCCEEDED,
                _execution.WorkflowExecutionPhase.FAILED,
                _execution.WorkflowExecutionPhase.TIMED_OUT,
                _execution.WorkflowExecutionPhase.ABORTED,
            ],
            ['test@flyte-org.slack.com'],
        ),
    ],
    default_inputs={}
)

# This launch plan customizes the input angle to 90 and the scale factor to 4x as `fixed_inputs`. Inputs cannot be
# changed now
scale4x_rotate90degrees_launchplan = workflows.ScaleAndRotateWorkflow.create_launch_plan(
    fixed_inputs={"angle": 90.0, "scale": 4},
)
