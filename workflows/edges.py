from __future__ import absolute_import
from __future__ import division
from __future__ import print_function

import urllib.request as _request

import cv2
from flytekit.common import utils
from flytekit.sdk.tasks import python_task, outputs, inputs
from flytekit.sdk.types import Types
from flytekit.sdk.workflow import workflow_class, Output, Input

@inputs(image_location=Types.String)
@outputs(parsed_image=Types.Blob)
@python_task
def edge_detection_canny(wf_params, image_location, parsed_image):
    with utils.AutoDeletingTempDir('test') as tmpdir:
        plane_fname = '{}/plane.jpg'.format(tmpdir.name)
        with _request.urlopen(image_location) as d, open(plane_fname, 'wb') as opfile:
            data = d.read()
            opfile.write(data)

        img = cv2.imread(plane_fname, 0)
        edges = cv2.Canny(img, 50, 200)  # hysteresis thresholds

        output_file = '{}/output.jpg'.format(tmpdir.name)
        cv2.imwrite(output_file, edges)

        parsed_image.set(output_file)

    
@workflow_class
class EdgeDetectorWf(object):
    image_input = Input(Types.String, required=True, help="Image to run for")
    run_edge_detection = edge_detection_canny(image_location=image_input)
    edges = Output(run_edge_detection.outputs.parsed_image, sdk_type=Types.Blob)

