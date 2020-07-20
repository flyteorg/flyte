from flytekit.sdk.tasks import inputs, outputs, python_task
from flytekit.sdk.types import Types
from flytekit.sdk.workflow import workflow_class, Input, Output, workflow
import cv2


@inputs(image=Types.Blob)
@outputs(out_image=Types.Blob)
@python_task(cache=True, cache_version="1")
def scale(wf_params, image, out_image):
    """
     Scale a given image by float.
     NOTE: Caching is enabled. So re-running for the same image should be NO-OP (or immediate return)
    """
    image.download()
    img = cv2.imread(image.local_path, 0)
    if img is None:
        raise Exception("Failed to read image")
    res = cv2.resize(img, None, fx=2, fy=2, interpolation=cv2.INTER_CUBIC)
    out_path = "scaled.jpg"
    cv2.imwrite(out_path, res)
    out_image.set(out_path)


@inputs(image=Types.Blob, angle=Types.Float, fail=Types.Boolean)
@outputs(out_image=Types.Blob)
@python_task(cache=True, cache_version="1")
def rotate(wf_params, image, angle, fail, out_image):
    """
    Rotate a given image by x-degress.
    NOTE: Caching is enabled. So re-running for the same image should be NO-OP (or immediate return)
    """
    if fail:
        raise Exception("User signaled failure")
    image.download()
    img = cv2.imread(image.local_path, 0)
    if img is None:
        raise Exception("Failed to read image")
    (h, w) = img.shape[:2]
    center = (w / 2, h / 2)
    mat = cv2.getRotationMatrix2D(center, angle, 1)
    res = cv2.warpAffine(img, mat, (w, h))
    out_path = "rotated.jpg"
    cv2.imwrite(out_path, res)
    out_image.set(out_path)


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

    scale_task = scale(image=in_image)
    rotate_task = rotate(image=scale_task.outputs.out_image, angle=angle, fail=True)

    out_image = Output(rotate_task.outputs.out_image, sdk_type=Types.Blob)


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
