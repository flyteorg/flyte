"""
.. autoclass:: flytekit.core.schedule.CronSchedule
   :noindex:

"""

import datetime
import re as _re
from typing import Optional

import croniter as _croniter

from flytekit.models import schedule as _schedule_models


# Duplicates flytekit.common.schedules.Schedule to avoid using the ExtendedSdkType metaclass.
class CronSchedule(_schedule_models.Schedule):
    """
    Use this when you have a launch plan that you want to run on a cron expression.
    This uses standard `cron format <https://docs.flyte.org/en/latest/concepts/schedules.html#cron-expression-table>`__
    in case where you are using default native scheduler using the schedule attribute.

    .. code-block::

        CronSchedule(
            schedule="*/1 * * * *",  # Following schedule runs every min
        )

    See the :std:ref:`User Guide <cookbook:cron schedules>` for further examples.
    """

    _VALID_CRON_ALIASES = [
        "hourly",
        "hours",
        "@hourly",
        "daily",
        "days",
        "@daily",
        "weekly",
        "weeks",
        "@weekly",
        "monthly",
        "months",
        "@monthly",
        "annually",
        "@annually",
        "yearly",
        "years",
        "@yearly",
    ]

    # Not a perfect regex but good enough and simple to reason about
    _OFFSET_PATTERN = _re.compile("([-+]?)P([-+0-9YMWD]+)?(T([-+0-9HMS.,]+)?)?")

    def __init__(
        self,
        cron_expression: Optional[str] = None,
        schedule: Optional[str] = None,
        offset: Optional[str] = None,
        kickoff_time_input_arg: Optional[str] = None,
    ):
        """
        :param str cron_expression: This should be a cron expression in AWS style.Shouldn't be used in case of native scheduler.
        :param str schedule: This takes a cron alias (see ``_VALID_CRON_ALIASES``) or a croniter parseable schedule.
          Only one of this or ``cron_expression`` can be set, not both. This uses standard `cron format <https://docs.flyte.org/en/latest/concepts/schedules.html#cron-expression-table>`_
          and is supported by native scheduler
        :param str offset:
        :param str kickoff_time_input_arg: This is a convenient argument to use when your code needs to know what time
          a run was kicked off. Supply the name of the input argument of your workflow to this argument here. Note
          that until Flyte has an atomic clock, there could be a few seconds here and there. That is, if your run is
          supposed to kick off at 3pm UTC every Weds, it may actually be 15:00:02 or something.  Example ::

            @workflow
            def my_wf(kickoff_time: datetime): ...

            schedule = CronSchedule(
                schedule="*/1 * * * *"
                kickoff_time_input_arg="kickoff_time")

        """
        if cron_expression is None and schedule is None:
            raise AssertionError("Either `cron_expression` or `schedule` should be specified.")

        if cron_expression is not None and offset is not None:
            raise AssertionError("Only `schedule` is supported when specifying `offset`.")

        if cron_expression is not None:
            CronSchedule._validate_expression(cron_expression)

        if schedule is not None:
            CronSchedule._validate_schedule(schedule)

        if offset is not None:
            CronSchedule._validate_offset(offset)

        super(CronSchedule, self).__init__(
            kickoff_time_input_arg,
            cron_expression=cron_expression,
            cron_schedule=_schedule_models.Schedule.CronSchedule(schedule, offset) if schedule is not None else None,
        )

    @staticmethod
    def _validate_expression(cron_expression: str):
        """
        Ensures that the set value is a valid cron string.  We use the format used in Cloudwatch and the best
        explanation can be found here:
            https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html#CronExpressions
        :param str cron_expression: cron expression
        """
        # We use the croniter lib to validate our cron expression.  Since on the admin side we use Cloudwatch,
        # we have a couple checks in order to line up Cloudwatch with Croniter.
        tokens = cron_expression.split()
        if len(tokens) != 6:
            raise ValueError(
                "Cron expression is invalid.  A cron expression must have 6 fields.  Cron expressions are in the "
                "format of: `minute hour day-of-month month day-of-week year`.  "
                "Use `schedule` for 5 fields cron expression.  Received: `{}`".format(cron_expression)
            )

        if tokens[2] != "?" and tokens[4] != "?":
            raise ValueError(
                "Scheduled string is invalid.  A cron expression must have a '?' for either day-of-month or "
                "day-of-week.  Please specify '?' for one of those fields.  Cron expressions are in the format of: "
                "minute hour day-of-month month day-of-week year.\n\n"
                "For more information: "
                "https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html#CronExpressions"
            )

        try:
            # Cut to 5 fields and just assume year field is good because croniter treats the 6th field as seconds.
            # TODO: Parse this field ourselves and check
            _croniter.croniter(" ".join(cron_expression.replace("?", "*").split()[:5]))
        except Exception:
            raise ValueError(
                "Scheduled string is invalid.  The cron expression was found to be invalid."
                f" Provided cron expr: {cron_expression}"
            )

    @staticmethod
    def _validate_schedule(schedule: str):
        if schedule.lower() not in CronSchedule._VALID_CRON_ALIASES:
            try:
                _croniter.croniter(schedule)
            except Exception:
                raise ValueError(
                    "Schedule is invalid. It must be set to either a cron alias or valid cron expression."
                    f" Provided schedule: {schedule}"
                )

    @staticmethod
    def _validate_offset(offset: str):
        if CronSchedule._OFFSET_PATTERN.fullmatch(offset) is None:
            raise ValueError("Offset is invalid. It must be an ISO 8601 duration. Provided offset: {}".format(offset))


class FixedRate(_schedule_models.Schedule):
    """
    Use this class to schedule a fixed-rate interval for a launch plan.

    .. code-block:: python

        from datetime import timedelta

        FixedRate(duration=timedelta(minutes=10))

    See the :std:ref:`fixed rate intervals` chapter in the cookbook for additional usage examples.
    """

    def __init__(self, duration: datetime.timedelta, kickoff_time_input_arg: Optional[str] = None):
        """
        :param datetime.timedelta duration:
        :param str kickoff_time_input_arg:
        """
        super(FixedRate, self).__init__(kickoff_time_input_arg, rate=self._translate_duration(duration))

    @staticmethod
    def _translate_duration(duration: datetime.timedelta):
        """
        :param datetime.timedelta duration: timedelta between runs
        :rtype: flytekit.models.schedule.Schedule.FixedRate
        """
        _SECONDS_TO_MINUTES = 60
        _SECONDS_TO_HOURS = _SECONDS_TO_MINUTES * 60
        _SECONDS_TO_DAYS = _SECONDS_TO_HOURS * 24

        if duration.microseconds != 0 or duration.seconds % _SECONDS_TO_MINUTES != 0:
            raise AssertionError(
                f"Granularity of less than a minute is not supported for FixedRate schedules.  Received: {duration}"
            )
        elif int(duration.total_seconds()) % _SECONDS_TO_DAYS == 0:
            return _schedule_models.Schedule.FixedRate(
                int(duration.total_seconds() / _SECONDS_TO_DAYS),
                _schedule_models.Schedule.FixedRateUnit.DAY,
            )
        elif int(duration.total_seconds()) % _SECONDS_TO_HOURS == 0:
            return _schedule_models.Schedule.FixedRate(
                int(duration.total_seconds() / _SECONDS_TO_HOURS),
                _schedule_models.Schedule.FixedRateUnit.HOUR,
            )
        else:
            return _schedule_models.Schedule.FixedRate(
                int(duration.total_seconds() / _SECONDS_TO_MINUTES),
                _schedule_models.Schedule.FixedRateUnit.MINUTE,
            )
