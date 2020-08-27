.. _getting-started-create-first:

########################################
Writing Your First Workflow
########################################

The easiest way to author a Flyte Workflow is using the provided python SDK called "FlyteKit".

You can save some effort by cloning the ``flytesnacks`` repo, and re-initializing it as a new git repository ::

  git clone git@github.com:lyft/flytesnacks.git myflyteproject
  cd myflyteproject
  rm -rf .git
  git init
  cd python

now open the "Makefile" and change the first line to ``IMAGE_NAME=myflyteproject``

Let's also remove the existing python task so we can write one from scratch. ::

  rm single_step/edges.py

Creating a Project
******************

In Flyte, workflows are organized into namespaces called "Projects". When you register a workflow, it must be registered under a project.

Lets create a new project called ``myflyteproject``. Use the project creation endpoint to create the new project ::

  curl -X POST localhost:30081/api/v1/projects -d '{"project": {"id": "myflyteproject", "name": "myflyteproject"} }'


Writing a Task
*****************

The most basic Flyte primitive is a "task". Flyte Tasks are units of work that can be composed in a workflow. The simplest way to write a Flyte task is using the Flyte Python SDK - flytekit.

Start by creating a new file ::

   mkdir -p workflows
   touch workflows/first.py

This directory has been marked in the `configuration file <https://github.com/lyft/flytesnacks/blob/764b82aca5701137ebc0eda4e818466e5acc9219/sandbox.config#L2>`_ as the location to look for workflows and tasks.  Begin by importing some of the libraries that we'll need for this example.

.. code-block:: python

  from __future__ import absolute_import
  from __future__ import division
  from __future__ import print_function
  
  import urllib.request as _request
  
  import cv2
  from flytekit.common import utils
  from flytekit.sdk.tasks import python_task, outputs, inputs
  from flytekit.sdk.types import Types
  from flytekit.sdk.workflow import workflow_class, Output, Input
  
From there, we can begin to write our first task.  It should look something like this. 

.. code-block:: python

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


Some of the new concepts demonstrated here are:

* ``wf_params``: The first argument to a python task is a Flyte SDK defined object that offers handlers like logging.
* Inputs and outputs are first defined in the decorator, and then passed into the argument of the function.  Note that the names in the function signature need to match those in the decorator arguments.
* A ``Blob`` is a Flyte Kit type that represents binary data.  It is used to offload data to a storage location like S3.  Here we use it to store an image.

Writing a Workflow
*********************
Next you need to call that task from a workflow.  In the same file, add these lines.

.. code-block:: python

  @workflow_class
  class EdgeDetectorWf(object):
      image_input = Input(Types.String, required=True, help="Image to run for")
      run_edge_detection = edge_detection_canny(image_location=image_input)
      edges = Output(run_edge_detection.outputs.parsed_image, sdk_type=Types.Blob)
  
This code block creates a workflow, with one task. The workflow itself has an input (the link to an image) that gets passed into the task, and an output, which is the processed image.


Interacting with Flyte
************************

Flyte fulfills tasks using docker images. You'll need to build a docker image from this code before it can run in Flyte. The repo has a make target to build the docker image for you ::

  make docker_build

If you have the flyte sandbox installed on your local machine, the image will be accessible to to your Flyte system. If you're running a remote Flyte instance, you'll need to upload this image to a remote registry such as Dockerhub, Amazon ECR, or Google Container Registry, so that it can be used by the Flyte system. 

To upload to a remote registry (or even local registry), use ::

  DOCKER_REGISTRY_USERNAME={username} DOCKER_REGISTRY_PASSWORD={pass} REGISTRY=docker.io make docker_build

Replace the values above with your registry username, password, and registry endpoint.

You may need to change the ``IMAGE_NAME`` in the Makefile to reflect your namespace in the docker registry. (ex ``{{my docker username}}/myflyteproject``)

With the image built, we just need to register the tasks and workflows. The process is the same as what we had done previously. ::

  docker run --network host -e FLYTE_PLATFORM_URL='127.0.0.1:30081' {{ your docker image }} pyflyte -p myflyteproject -d development -c sandbox.config register workflows

After this, you should be able to visit the Flyte UI, and run the workflow as you did with ``flytesnacks`` previously.
