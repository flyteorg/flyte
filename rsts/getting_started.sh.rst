#!/bin/bash
:'

.. _gettingstarted:

Getting started
---------------

.. rubric:: Estimated time to complete: 3 minutes.


Prerequisites
***************

Make sure you have `docker installed <https://docs.docker.com/get-docker/>`__ and `git <https://git-scm.com/>`__ installed, then install flytekit:

Steps
*****

'

# #. First install the python Flytekit SDK (maybe in a virtual environment) and clone the `flytekit-python-template <https://github.com/flyteorg/flytekit-python-template>`_ repo::

  pip install flytekit
  git clone git@github.com:flyteorg/flytekit-python-template.git myflyteapp
  cd myflyteapp


' #. The repo comes with a sample workflow, which can be found under ``myapp/workflows/example.py``. The structure below shows the most important files and how a typical flyteapp should be laid out.
.. note::

    .
    ├── Dockerfile
    ├── docker_build_and_tag.sh
    ├── myapp
    │   ├── __init__.py
    │   └── workflows
    │       ├── __init__.py
    │       └── example.py
    └── requirements.txt

You can use pip-compile to build your requirements file. the Dockerfile that comes with this is not GPU ready, but is a simple Dockerfile that should work for most apps.
'
The workflow can be run locally simply by running it as a python script - ``note the __main__ at the bottom of the file``::

    python myapp/workflows/example.py

# #. Let us install ``flytectl`` (link). ``flytectl`` is a commandline interface for flyte. ::

    brew install flyteorg/homebrew-tap/flytectl

#. #. you can upgrade flytectl using::

    brew upgrade flytectl



# ``homebrew`` supports only osx. You can install flytectl on most other platforms using::

    curl -s https://raw.githubusercontent.com/lyft/flytectl/master/install.sh | bash


# Setup flytectl config using ... doc to configuring flytectl

# #. Now lets start a local sandbox cluster to test your workflow deployment. link to doc explaining what is sandbox. You can skip this step if you already have a Flyte sandbox or a hosted Flyte deployed.::

    flytectl sandbox start --source myflyteapp

# #. Flyte uses docker containers to package your workflows and tasks and send it to the remote Flyte cluster. Thus if you notice there is a ``Dockerfile`` already in the cloned repo::

    flytectl sandbox exec -- docker build . --tag
    optional: docker push

# #. Now that the container is built, lets provide this information to the Flyte backend. To do that you have to package the workflow using the pyflyte cli, that is bundled with flytekit::

    pyflyte package ...

# #. Now lets upload this package to flyte backend::

    flytectl register files my_wf.pb

# #. You can create an execution using flytectl as follows::

    blah

# #. you can also create the execution using the UI::

    blah


