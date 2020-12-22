"""

06: Getting notifications on workflow termination
-------------------------------------------------

For background on launch plans, refer to :any:`launch_plans`.

For up-to-date documentation on notifications see the `official docs <https://lyft.github.io/flyte/user/features/notifications.html>`_

"""

# %%
# Let's consider the following example workflow:
from flytekit import LaunchPlan, task, workflow
from flytekit.annotated.notification import Email
from flytekit.models.core.execution import WorkflowExecutionPhase


@task
def double_int_and_print(a: int) -> str:
    return str(a * 2)


@workflow
def int_doubler_wf(a: int) -> str:
    doubled = double_int_and_print(a=a)
    return doubled

# This launch plan triggers email notifications when the workflow execution it triggered reaches the phase `SUCCEEDED`.
int_doubler_wf_lp = LaunchPlan.create(
    "int_doubler_wf",
    int_doubler_wf,
    default_inputs={"a": 4},
    notifications=[
        Email(
            phases=[WorkflowExecutionPhase.SUCCEEDED],
            recipients_email=["admin@example.com"],
        )
    ],
)

# %%
# Notifications shine when used for scheduled workflows to alert on failures:
from datetime import timedelta

from flytekit.annotated.notification import PagerDuty
from flytekit.annotated.schedule import FixedRate

int_doubler_wf_scheduled_lp = LaunchPlan.create(
    "int_doubler_wf_scheduled",
    int_doubler_wf,
    default_inputs={"a": 4},
    notifications=[
        PagerDuty(
            phases=[WorkflowExecutionPhase.FAILED, WorkflowExecutionPhase.TIMED_OUT],
            recipients_email=["abc@pagerduty.com"],
        )
    ],
    schedule=FixedRate(duration=timedelta(days=1)),
)


# %%
# If you desire you can combine notifications with different permutations of terminal phases and recipient targets:
from flytekit.annotated.notification import Slack

wacky_int_doubler_lp = LaunchPlan.create(
    "wacky_int_doubler",
    int_doubler_wf,
    default_inputs={"a": 4},
    notifications=[
        Email(
            phases=[WorkflowExecutionPhase.FAILED],
            recipients_email=["me@example.com", "you@example.com"],
        ),
        Email(
            phases=[WorkflowExecutionPhase.SUCCEEDED],
            recipients_email=["myboss@example.com"],
        ),
        Slack(
            phases=[
                WorkflowExecutionPhase.SUCCEEDED,
                WorkflowExecutionPhase.ABORTED,
                WorkflowExecutionPhase.TIMED_OUT,
            ],
            recipients_email=["myteam@slack.com"],
        ),
    ],
)
