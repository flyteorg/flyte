from flytekit.sdk.types import Types
from flytekit.sdk.workflow import workflow_class, Input, Output

from task import tasks


###
# Both workflows are identical, just one value for the rotate task is defaulted to "fail=True" to force the workflow to
# Fail

@workflow_class
class FailingWorkflow(object):
    """
    This workflow is  two step workflow,
    Step 1: scale an image
    Step 2: Rotate an image
    NOTE: This is not an efficient workflow as one image - scaling and rotation can be done with one OPEN CV call. But this example exists only for a demo

    Step 2: in this case will always fail as it is hard-coded to indicate fail=True
    """
    in_image = Input(Types.Blob, default=Types.Blob.create_at_known_location(
        "https://miro.medium.com/max/1400/1*qL8UYfaStcEo_YVPrA4cbA.png"))
    angle = Input(Types.Float, default=180.0)
    scale = Input(Types.Integer, default=2)

    scale_task = tasks.scale(image=in_image, scale_factor=scale)
    rotate_task = tasks.rotate(image=scale_task.outputs.out_image, angle=angle, fail=True)

    out_image = Output(rotate_task.outputs.out_image, sdk_type=Types.Blob)


@workflow_class
class ScaleAndRotateWorkflow(object):
    """
    This workflow is  two step workflow,
    Step 1: scale an image
    Step 2: Rotate an image
    NOTE: This is not an efficient workflow as one image - scaling and rotation can be done with one OPEN CV call. But this example exists only for a demo

    Step 2: in this case will always fail as it is hard-coded to indicate fail=True
    """
    in_image = Input(Types.Blob, default=Types.Blob.create_at_known_location(
        "https://miro.medium.com/max/1400/1*qL8UYfaStcEo_YVPrA4cbA.png"))
    angle = Input(Types.Float, default=180.0)
    scale = Input(Types.Integer, default=2)

    scale_task = tasks.scale(image=in_image, scale_factor=scale)
    # This is the same workflow as the FailingWorkflow, just that "fail=True"
    rotate_task = tasks.rotate(image=scale_task.outputs.out_image, angle=angle, fail=False)

    out_image = Output(rotate_task.outputs.out_image, sdk_type=Types.Blob)
