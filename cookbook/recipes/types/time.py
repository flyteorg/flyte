from __future__ import absolute_import
from __future__ import print_function

from flytekit.sdk.tasks import inputs, outputs, python_task
from flytekit.sdk.types import Types
from flytekit.sdk.workflow import workflow_class, Input, Output


@inputs(dt=Types.Datetime, duration=Types.Timedelta)
@outputs(new_time=Types.Datetime)
@python_task
def time_task(wf_params, dt, duration, new_time):
    """
    Simple Task that adds the duration to the passed date time.
    Important to note, the received dt is a datetime object
    """
    wf_params.logging.info("Running time_task")
    new_time.set(dt + duration) 


@workflow_class
class TimeDemoWorkflow(object):
    dt = Input(Types.Datetime, help="Input time")
    duration = Input(Types.Timedelta, help="Input timedelta")
    time_example = time_task(dt=dt, duration=duration)
    new_time = Output(time_example.outputs.new_time, sdk_type=Types.Datetime)
