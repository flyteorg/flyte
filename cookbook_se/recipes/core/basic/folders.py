"""
Work with folders
---------------------

Please also see the entry on files. After files, folders are the other fundamental operating system primitive users might find themselves working with. The Flyte IDL's support of folders take the form of `multi-part blobs <https://github.com/lyft/flyteidl/blob/cee566b2e6e109120f1bb34c980b1cfaf006a473/protos/flyteidl/core/types.proto#L50>`__.
"""
import pathlib
import os
import urllib.request

import cv2
import flytekit
from flytekit import task, workflow
from flytekit.types.directory import FlyteDirectory


# %%
# Playing on the same example used in the File chapter, this first task downloads a bunch of files into a directory,
# and then returns a Flyte object referencing them.
default_images = [
    "https://upload.wikimedia.org/wikipedia/commons/a/a8/Fractal_pyramid.jpg",
    "https://upload.wikimedia.org/wikipedia/commons/thumb/a/ad/Julian_fractal.jpg/256px-Julian_fractal.jpg",
]


# %%
# This task downloads the two files above using non-Flyte libraries, and returns the path to the folder, in a FlyteDirectory object.
@task
def download_files() -> FlyteDirectory:
    working_dir = flytekit.current_context().working_directory
    pp = pathlib.Path(os.path.join(working_dir, "images"))
    pp.mkdir(exist_ok=True)
    for idx, remote_location in enumerate(default_images):
        local_image = os.path.join(working_dir, "images", f"image_{idx}.jpg")
        urllib.request.urlretrieve(remote_location, local_image)

    return FlyteDirectory(path=os.path.join(working_dir, "images"))


# %%
# This helper method is a purely Python function; no Flyte components here.
def rotate(local_image: str):
    """
    In place rotation of the image
    """
    img = cv2.imread(local_image, 0)
    if img is None:
        raise Exception("Failed to read image")
    (h, w) = img.shape[:2]
    center = (w / 2, h / 2)
    mat = cv2.getRotationMatrix2D(center, 180, 1)
    res = cv2.warpAffine(img, mat, (w, h))
    # out_path = os.path.join(working_dir, "rotated.jpg")
    cv2.imwrite(local_image, res)


# %%
# This task accepts the previously downloaded folder, and calls the rotate function above on each. Since the rotate function does the image manipulation in place, we just create a new FlyteDirectory object pointed to the same place.
@task
def rotate_all(img_dir: FlyteDirectory) -> FlyteDirectory:
    """
    Download the given image, rotate it by 180 degrees
    """
    for img in [os.path.join(img_dir, x) for x in os.listdir(img_dir)]:
        rotate(img)
    return FlyteDirectory(path=img_dir.path)


@workflow
def download_and_rotate() -> FlyteDirectory:
    directory = download_files()
    return rotate_all(img_dir=directory)


if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(f"Running main {download_and_rotate()}")
