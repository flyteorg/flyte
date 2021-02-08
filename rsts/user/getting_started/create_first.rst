.. _getting-started-create-first:

########################################
Writing Your First Workflow
########################################

By the end of this getting started guide you'll be familiar with how easy it is to author a Flyte workflow, run and then deploy it to a sandbox cluster (which can be run locally or on any kubernetes cluster).

The easiest way to author a Flyte Workflow is using the provided python SDK called "FlyteKit".

You can save some effort by cloning the ``flytekit-python-template`` repo, and re-initializing it as a new git repository ::

  git clone git@github.com:lyft/flytekit-python-template.git myflyteproject
  cd myflyteproject
  rm -rf .git
  git init
  cd python

Writing a Task
*****************

The most basic Flyte primitive is a "task". Flyte Tasks are units of work that can be composed in a workflow. The simplest way to write a Flyte task is using the Flyte Python SDK - flytekit.

Start by creating a new file ::


   touch myapp/workflows/first.py

And add the required imports we'll need for this example:


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

* Use the :py:func:`flytekit.task` decorator to convert your typed python function to a Flyte task.
* A :py:class:`flytekit.types.file.FlyteFile` is a Flytekit type that represents binary data.  It is used to offload data to a storage location like S3.  Here we use it to store an image.


You can call this task ``edge_detection_canny(image_location="https://images.unsplash.com/photo-1512289984044-071903207f5e?ixid=MXwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHw%3D&ixlib=rb-1.2.1&auto=format&fit=crop&w=2250&q=80")`` and iterate locally before adding it to part of a larger overall workflow.



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

.. note::

   Every invocation of a Flyte workflow requires specifying keyword args.

Interacting with Flyte
************************

TODO: fill this section out.
1. Setup a sandbox deployment
2. Create a project
3. Register your workflows
4. Run your workflows


Expanded examples
*****************

If you're interested in learning more and want to try more complex examples, `Flytesnacks Cookbook <https://flytecookbook.readthedocs.io/en/latest/>`__
