######################################
Quick Start Examples
######################################

.. _getting-started-examples:

Before you can run any workflows, you'll first need to register a workflow in Flyte. We've written an example Flyte workflow, and placed it into a docker image called ``flytesnacks``.

The example workflow takes in an image url as input, does edge detection, and produces an image showing the edges in the photo.  

Workflow Setup in Flyte
**************************

Registration
==========================
.. _getting-started-examples-registration:

If you're using the ``sandbox`` flyte installation, you can use the following command to register our example workflow with Flyte ::

  docker run --network host -e FLYTE_PLATFORM_URL='127.0.0.1:30081' lyft/flytesnacks:v0.1.0 pyflyte -p flytesnacks -d development -c sandbox.config register workflows

This command will register the workflow with your Flyte app under the ``development`` domain of the project ``flytesnacks``.

NOTE: if your Flyte endpoint is something other than ``localhost:30081``, change the ``FLYTE_PLATFORM_URL`` value accordingly. 

Running Workflows in Flyte
****************************

Creating an execution
==========================

Now that your workflow is registered, you can visit the Flyte console to run the workflow. 

From the flyte console homepage http://localhost:30081/console, click the "development" link under ``flytesnacks``. This will show you the list of workflows registered under the ``development`` domain of the ``flytesnacks`` project.

Click on the ``workflows.edges.EdgeDetectorWf`` workflow. This will take you to the Workflow Details page.

Click the "Launch Workflow" button. This will open the Launch Workflow Form. 

Leave the ``Workflow Version`` and ``Launch Plan`` default values, but insert any publicly accessible image url in the ``image_input`` section. For example, you can use https://images.ctfassets.net/q8mvene1wzq4/1OQ8OBLzXGv8pVjFTLf0QF/9a7b8cdb982161daebd5618fc7cb5041/Car_blue_L.png.

Click "Launch" at the bottom of the form (you may need to scroll down).

In a few moments, you'll see the execution getting fulfilled by the Flyte system. Observe as Flyte runs through the workflow steps.

When the workflow is complete, click the "run-edge-detection" link under "Node", this will show you some details about the execution. Click the "Outputs" tab. You should see that the workflow produced an output telling you where it stored the produced image.

To find this image, visit the minio UI at http://localhost:30081/minio (the sandbox username is ``minio`` and the sandbox password is ``miniostorage``).

Follow the path given in your workflow output. If you download the file, you'll see that the workflow produced the edge-detected version of the image url input. 
