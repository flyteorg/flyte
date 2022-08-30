.. custom-python-dependencies:

###############
Custom Python Dependencies
###############

This section guides you on how to run a task that relies on pip (or apt) dependencies that are not in the base image. 
In the getting started guide, we used a default image in the first step. Now if you are project has some custom Python or even apt dependencies, then meet the --image flag inside of ``pyflyte``. 

Consistent dependency management is a challenge with python projects, so Flyte uses `Docker containers <https://www.docker.com/resources/what-container/>`__ to manage dependencies for your project.
``pyflyte run --remote`` uses a default image bundled with flytekit, which contains numpy, pandas, and flytekit and matches your current python (major, minor) version.
If you want to use a custom image, create a Dockerfile, build the Docker image, and push it to a registry that is accessible to your cluster.
(Note: You’ll have to build one such image and push it to a registry, say GitHub registry. Make sure you modify the requirements (in the case you’re looking at flytesnacks) or the dockerfile for apt deps.)

      .. prompt :: bash $

        docker build . --tag <registry/repo:version>
        docker push <registry/repo:version>

And, use the ``--image`` flag and provide the fully qualified image name of your image to the ``pyflyte run`` command.

      .. prompt :: bash $

        pyflyte run --image <registry/repo:version> --remote example.py wf --n 500 --mean 42 --sigma 2

Here is an example:
      
      .. prompt :: bash $
        
         docker build . -f core/Dockerfile -t ghcr.io/user/torchimage:0.0.1
         docker push ghcr.io/user/torchimage:0.0.1

And then if you’re using the ``pyflyte run`` command, you could use: 
      .. prompt :: bash $  

         pyflyte run --remote --image ghcr.io/user/torchimage:0.0.1 basics.py wf       

If you want to build an image with your Flyte project's code built-in, refer to the :doc:`Deploying Workflows Guide <cookbook:auto/deployment/deploying_workflows>`.