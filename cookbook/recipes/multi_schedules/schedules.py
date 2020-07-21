from recipes.workflows import workflows
from flytekit.common import schedules as _schedules, notifications as _notifications
from flytekit.models.core import execution as _execution
from flytekit.models.common import Labels, Annotations
from datetime import timedelta

# A Workflow can have multiple schedules. One per launch plan

# Example 1 show cron schedule
scale_rotate_cronscheduled_launchplan = workflows.ScaleAndRotateWorkflow.create_launch_plan(
    schedule=_schedules.CronSchedule("0/30 * * * ? *"),
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
)

# Example 2 shows Fixed Rate schedule as an example
scale_rotate_fixedRateScheduled_launchplan = workflows.ScaleAndRotateWorkflow.create_launch_plan(
    schedule=_schedules.FixedRate(duration=timedelta(hours=1)), )

# Example 3: Execution time is implicitly passed as an input
