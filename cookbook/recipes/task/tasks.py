import cv2
from flytekit.sdk.tasks import inputs, outputs, python_task
from flytekit.sdk.types import Types

@inputs(image=Types.Blob, scale_factor=Types.Integer)
@outputs(out_image=Types.Blob)
@python_task(cache=True, cache_version="1")
def scale(wf_params, image, scale_factor, out_image):
    """
     Scale a given image by float.
     NOTE: Caching is enabled. So re-running for the same image should be NO-OP (or immediate return)
    """
    image.download()
    img = cv2.imread(image.local_path, 0)
    if img is None:
        raise Exception("Failed to read image")
    res = cv2.resize(img, None, fx=scale_factor, fy=scale_factor, interpolation=cv2.INTER_CUBIC)
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