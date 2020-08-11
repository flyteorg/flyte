import os
import urllib.request

import cv2
from flytekit.common import utils
from flytekit.sdk.tasks import dynamic_task, inputs, outputs, python_task
from flytekit.sdk.types import Types
from flytekit.sdk.workflow import Input, workflow_class, Output

default_images = [
    'https://upload.wikimedia.org/wikipedia/commons/a/a8/Fractal_pyramid.jpg',
    'https://upload.wikimedia.org/wikipedia/commons/thumb/2/21/Mandel_zoom_00_mandelbrot_set.jpg/640px-Mandel_zoom_00_mandelbrot_set.jpg',
    'https://upload.wikimedia.org/wikipedia/commons/thumb/a/ad/Julian_fractal.jpg/256px-Julian_fractal.jpg',
]


@inputs(image_location=Types.String)
@outputs(out_image=Types.Blob)
@python_task(cpu_request="200m", cpu_limit="200m", memory_request="500Mi", memory_limit="500Mi")
def rotate(wf_params, image_location, out_image):
    """
    Download the given image, rotate it by 180 degrees
    """
    with utils.AutoDeletingTempDir('rotation') as tmp:
        local_image = os.path.join(tmp.name, 'incoming.jpg')
        urllib.request.urlretrieve(image_location, local_image)
        img = cv2.imread(local_image, 0)
        if img is None:
            raise Exception("Failed to read image")
        (h, w) = img.shape[:2]
        center = (w / 2, h / 2)
        mat = cv2.getRotationMatrix2D(center, 180, 1)
        res = cv2.warpAffine(img, mat, (w, h))
        out_path = os.path.join(tmp.name, "rotated.jpg")
        cv2.imwrite(out_path, res)
        out_image.set(out_path)


@inputs(input_images=[Types.String])
@outputs(rotated_images=[Types.Blob])
@dynamic_task
def batch_rotator(wf_params, input_images, rotated_images):
    results = []
    for ii in input_images:
        rotate_task = rotate(image_location=ii)
        yield rotate_task
        results.append(rotate_task.outputs.out_image)

    rotated_images.set(results)


@workflow_class
class BatchRotateWorkflow(object):
    in_images = Input([Types.String], default=default_images)
    run_map = batch_rotator(input_images=in_images)
    wf_output = run_map.outputs.rotated_images
