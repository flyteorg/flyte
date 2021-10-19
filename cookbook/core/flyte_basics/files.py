"""
Working With Files
-------------------

Files are one of the most fundamental entities that users of Python work with, and they are fully supported by Flyte.
In the IDL, they are known as `Blob <https://github.com/lyft/flyteidl/blob/cee566b2e6e109120f1bb34c980b1cfaf006a473/protos/flyteidl/core/literals.proto#L33>`__ literals
which are backed by the `blob type <https://github.com/lyft/flyteidl/blob/cee566b2e6e109120f1bb34c980b1cfaf006a473/protos/flyteidl/core/types.proto#L47>`__.

Let's assume our mission here is pretty simple. We take in a couple of links, download the pictures, rotate them, and return the rotated images.
"""

# %%
# First, let's import the libraries.
import os

import cv2
import flytekit
from flytekit import task, workflow
from flytekit.types.file import JPEGImageFile
from flytekit import FlyteContext

# %%
# ``JPEGImageFile`` is a pre-formatted FlyteFile type. It is equivalent to ``FlyteFile[typing.TypeVar("jpeg")]``.
#
# .. note::
#   The ``FlyteFile`` literal can be scoped with a string, which gets inserted into the format of the Blob type ("jpeg" is the string in
#   ``FlyteFile[typing.TypeVar("jpeg")]``). The format is entirely optional, and if not specified, defaults to ``""``.


# %%
# Next, we write a task that accepts ``JPEGImageFile`` as an input and returns the rotated image as an output,
# which again is the ``JPEGImageFile``.
# Files do not have a native object in Python, so we had to write one ourselves.
# There does exist the ``os.PathLike`` protocol, but nothing implements it.
@task
def rotate(image_location: JPEGImageFile) -> JPEGImageFile:
    """
    Download the given image, rotate it by 180 degrees
    """
    working_dir = flytekit.current_context().working_directory
    image_location.download()
    print(image_location.path)
    img = cv2.imread(image_location.path, 0)
    if img is None:
        raise Exception("Failed to read image")
    (h, w) = img.shape[:2]
    center = (w / 2, h / 2)
    mat = cv2.getRotationMatrix2D(center, 180, 1)
    res = cv2.warpAffine(img, mat, (w, h))
    out_path = os.path.join(
        working_dir,
        f"rotated-{os.path.basename(image_location.path).rsplit('.')[0]}.jpg",
    )
    cv2.imwrite(out_path, res)
    return JPEGImageFile(path=out_path)


# %%
# When image URL is sent to the task, the Flytekit engine translates it into a ``FlyteFile`` object on the local drive
# (but doesn't download it).
# The act of calling ``download`` method should trigger the download. 
# ``_SpecificFormatClass``'s path enables OpenCV to read the file.
#
# When this task finishes, Flytekit engine returns the ``FlyteFile`` instance, finds a location
# in Flyte's object store (usually S3), uploads the file to that location and creates a Blob literal pointing to it.
#
# .. tip::
#
#   The ``rotate`` task works with ``FlyteFile``, too. However, ``JPEGImageFile`` helps attach the content information.

# %%
# We now define the workflow.
@workflow
def rotate_one_workflow(in_image: JPEGImageFile) -> JPEGImageFile:
    return rotate(image_location=in_image)


# %%
# Finally, let's execute it!
if __name__ == "__main__":
    default_images = [
        "https://media.sketchfab.com/models/e13940161fb64746a4f6753f76abe886/thumbnails/b7e1ba951ffb46a4ad584ba8ae400d17/e9de09ac9c7941f1924bd384e74a5e2e.jpeg",
        "https://upload.wikimedia.org/wikipedia/commons/d/d2/Julia_set_%28C_%3D_0.285%2C_0.01%29.jpg",
    ]
    print(f"Running {__file__} main...")
    for index, each_image in enumerate(default_images):
        print(
            f"Running rotate_one_workflow(in_image=default_images[{index}]) {rotate_one_workflow(in_image=each_image)}"
        )
