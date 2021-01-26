from flytekit.sdk.types import Types
from flytekit.sdk.workflow import workflow_class, Input, Output
from recipes.task.tasks import rotate


@workflow_class
class BackfillWorkflow(object):
    """
    So if FailingWorkflow Fails, we can resurrect and backfill the FailingWorkflow, using the BackfillWorkflow.
    The Backfill workflow just has one step
    """
    in_image = Input(Types.Blob, required=True)
    angle = Input(Types.Float, default=180.0)

    rotate_task = rotate(image=in_image, angle=angle, fail=False)

    out_image = Output(rotate_task.outputs.out_image, sdk_type=Types.Blob)
