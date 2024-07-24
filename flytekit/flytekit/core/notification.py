"""
Notifications are primarily used when defining Launch Plans (also can be used when launching executions) and will trigger
the Flyte platform to send emails when a workflow run reaches certain stages (fails or succeeds, etc.).

.. note::

    Notifications require some setup and configuration on the Flyte platform side. Please contact your Flyte platform
    admins to get this feature enabled. See :std:ref:`cookbook:setting up workflow notifications`

Each notification type takes a list of :py:class:`flytekit.models.core.execution.WorkflowExecutionPhase` and a list of
emails. Even though there are different notification classes in this module, they all just send email. The differentiation
offers semantic meaning to the end-user but do not functionally behave differently. Successful integration with Slack
and Pagerduty is incumbent on those email API being set-up correctly.

.. autoclass:: flytekit.core.notification.Notification

"""
from typing import List

from flytekit.models import common as _common_model
from flytekit.models.core import execution as _execution_model


# Duplicates flytekit.common.notifications.Notification to avoid using the ExtendedSdkType metaclass.
class Notification(_common_model.Notification):
    VALID_PHASES = {
        _execution_model.WorkflowExecutionPhase.ABORTED,
        _execution_model.WorkflowExecutionPhase.FAILED,
        _execution_model.WorkflowExecutionPhase.SUCCEEDED,
        _execution_model.WorkflowExecutionPhase.TIMED_OUT,
    }

    def __init__(
        self,
        phases: List[int],
        email: _common_model.EmailNotification = None,
        pager_duty: _common_model.PagerDutyNotification = None,
        slack: _common_model.SlackNotification = None,
    ):
        """
        :param list[int] phases: A required list of phases for which to fire the event.  Events can only be fired for
            terminal phases.  Phases should be as defined in: flytekit.models.core.execution.WorkflowExecutionPhase
        """
        self._validate_phases(phases)
        super(Notification, self).__init__(phases, email=email, pager_duty=pager_duty, slack=slack)

    def _validate_phases(self, phases: List[int]):
        """
        :param list[int] phases:
        """
        if len(phases) == 0:
            raise AssertionError("You must specify at least one phase for a notification.")
        for phase in phases:
            if phase not in self.VALID_PHASES:
                raise AssertionError(f"Invalid phase: {phase}. only terminal states are permitted for notifications")


class PagerDuty(Notification):
    """
    This notification should be used when sending emails to the PagerDuty service.

    .. code-block:: python

        from flytekit.models.core.execution import WorkflowExecutionPhase

        PagerDuty(phases=[WorkflowExecutionPhase.SUCCEEDED], recipients_email=["my-team@email.com"])
    """

    def __init__(self, phases: List[int], recipients_email: List[str]):
        """
        :param list[int] phases: A required list of phases for which to fire the event.  Events can only be fired for
            terminal phases.  Phases should be as defined in: flytekit.models.core.execution.WorkflowExecutionPhase
        :param list[str] recipients_email: A required non-empty list of recipients for the notification.
        """
        super(PagerDuty, self).__init__(phases, pager_duty=_common_model.PagerDutyNotification(recipients_email))


class Email(Notification):
    """
    This notification should be used when sending regular emails to people.

    .. code-block:: python

        from flytekit.models.core.execution import WorkflowExecutionPhase

        Email(phases=[WorkflowExecutionPhase.SUCCEEDED], recipients_email=["my-team@email.com"])
    """

    def __init__(self, phases: List[int], recipients_email: List[str]):
        """
        :param list[int] phases: A required list of phases for which to fire the event.  Events can only be fired for
            terminal phases. Phases should be as defined in: :py:class:`flytekit.models.core.execution.WorkflowExecutionPhase`
        :param list[str] recipients_email: A required non-empty list of recipients for the notification.
        """
        super(Email, self).__init__(phases, email=_common_model.EmailNotification(recipients_email))


class Slack(Notification):
    """
    This notification should be used when sending emails to the Slack.

    .. code-block:: python

        from flytekit.models.core.execution import WorkflowExecutionPhase

        Slack(phases=[WorkflowExecutionPhase.SUCCEEDED], recipients_email=["my-team@email.com"])
    """

    def __init__(self, phases: List[int], recipients_email: List[str]):
        """
        :param list[int] phases: A required list of phases for which to fire the event.  Events can only be fired for
            terminal phases.  Phases should be as defined in: flytekit.models.core.execution.WorkflowExecutionPhase
        :param list[str] recipients_email: A required non-empty list of recipients for the notification.
        """
        super(Slack, self).__init__(phases, slack=_common_model.SlackNotification(recipients_email))
