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


Let's take a look at how easy it is to use Flyte by writing a task from scratch.

Creating a Project
******************

In Flyte, workflows are organized into namespaces called "Projects". When you register a workflow, it must be registered under a project.

For example, to create a new project called ``myflyteproject``: use the project creation endpoint to create the new project ::

  curl -X POST localhost:30081/api/v1/projects -d '{"project": {"id": "myflyteproject", "name": "myflyteproject"} }'


Writing a Task
*****************

The most basic Flyte primitive is a "task". Flyte Tasks are units of work that can be composed in a workflow. The simplest way to write a Flyte task is using the Flyte Python SDK - flytekit.

Start by creating a new file ::


   touch cookbook/recipes/core/first.py


.. code-block:: python

  from urllib import request
  
  import cv2
  import flytekit
  from flytekit import task, workflow
  from flytekit.types.file import FlyteFile
  
From there, we can begin to write our first task.  It should look something like this:

.. code-block:: python

  @task
  def edge_detection_canny(image_location:str) -> FlyteFile:
      working_dir = flytekit.current_context().working_directory:
      plane_fname = '{}/plane.jpg'.format(working_dir.name)
      with request.urlopen(image_location) as d, open(plane_fname, 'wb') as opfile:
          data = d.read()
          opfile.write(data)

      img = cv2.imread(plane_fname, 0)
      edges = cv2.Canny(img, 50, 200)  # hysteresis thresholds

      output_file = '{}/output.jpg'.format(working_dir.name)
      cv2.imwrite(output_file, edges)

      return FlyteFile["jpg"](path=output_file)


Some of the new concepts demonstrated here are:

* Use the ``@task`` decorator to convert your typed python function to a Flyte task.
* A ``FlyteFile`` is a Flyte Kit type that represents binary data.  It is used to offload data to a storage location like S3.  Here we use it to store an image.

Writing a Workflow
*********************
Next you need to call that task from a workflow.  In the same file, add these lines.

.. code-block:: python

   @workflow
   class EdgeDetectorWf(image_input: str) -> FlyteFile:
       edges = edge_detection_canny(image_location=image_input)
       return edges

This code block creates a workflow, with one task. The workflow itself has an input (the link to an image) that gets passed into the task, and an output, which is the processed image.

You can call this workflow ``EdgeDetectorWf(image_input=...)`` and iterate locally before moving on to register it with Flyte.

Interacting with Flyte
************************

For detailed and interactive example workflow 'recipes' check out the `Flytesnacks Cookbook <https://flytecookbook.readthedocs.io/en/latest//>`_
