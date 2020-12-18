"""
08: Work with files
-------------------

Files are one of the most fundamental things that users of Python work with, and they are fully supported by Flyte.
In the IDL, they are known as `Blob <https://github.com/lyft/flyteidl/blob/cee566b2e6e109120f1bb34c980b1cfaf006a473/protos/flyteidl/core/literals.proto#L33>`__ literals
and are backed by the `blob type <https://github.com/lyft/flyteidl/blob/cee566b2e6e109120f1bb34c980b1cfaf006a473/protos/flyteidl/core/types.proto#L47>`__

Note that the type message includes a ``format`` field which is text.
"""

import os
import urllib.request

import cv2
import flytekit
from flytekit import task, workflow
from flytekit.types import FlyteFile

# %%
# Let's assume our mission here is pretty simple. We want to take each of these links, download the picture, rotate it
# and return the file.
default_images = [
    "https://upload.wikimedia.org/wikipedia/commons/a/a8/Fractal_pyramid.jpg",
    "https://upload.wikimedia.org/wikipedia/commons/thumb/2/21/Mandel_zoom_00_mandelbrot_set.jpg/"
    "640px-Mandel_zoom_00_mandelbrot_set.jpg",
    "https://upload.wikimedia.org/wikipedia/commons/thumb/a/ad/Julian_fractal.jpg/256px-Julian_fractal.jpg",
]


# %%
# Note the signature of the return type of this task is a ``FlyteFile``. Files do not have a native object in Python
# so we had to write one ourselves. There does exist the ``os.PathLike`` protocol, but nothing implements it.
#
# When this task finishes, the flytekit engine will detect the ``FlyteFile`` instance being returned, find a location
# in Flyte's object store (usually S3), upload the file to that location, and create a Blob literal pointing to it.
#
# Note that the ``FlyteFile`` literal is scoped with a string, which gets inserted into the format of the Blob type.
# The ``[]`` are entirely optional, and if you don't specify it, the format will just be an ``""``.
@task
def rotate(image_location: str) -> FlyteFile:
    """
    Download the given image, rotate it by 180 degrees
    """
    working_dir = flytekit.current_context().working_directory
    local_image = os.path.join(working_dir, "incoming.jpg")
    urllib.request.urlretrieve(image_location, local_image)
    img = cv2.imread(local_image, 0)
    if img is None:
        raise Exception("Failed to read image")
    (h, w) = img.shape[:2]
    center = (w / 2, h / 2)
    mat = cv2.getRotationMatrix2D(center, 180, 1)
    res = cv2.warpAffine(img, mat, (w, h))
    out_path = os.path.join(working_dir, "rotated.jpg")
    cv2.imwrite(out_path, res)
    return FlyteFile["jpg"](path=out_path)


@workflow
def rotate_one_workflow(in_image: str) -> FlyteFile:
    return rotate(image_location=in_image)


# %%
# Execute it
if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(
        f"Running rotate_one_workflow(in_image=default_images[0]) {rotate_one_workflow(in_image=default_images[0])}"
    )
