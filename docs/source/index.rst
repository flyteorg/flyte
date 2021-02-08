.. flytectl doc

##########################################
Welcome to ``Flytectl``'s documentation!
##########################################


Installation
=============
Flytectl is a Golang binary and can be installed on any platform supported by
golang. To install simply copy paste the following into the command-line

.. prompt:: bash

   curl -s https://raw.githubusercontent.com/lyft/flytectl/master/install.sh | bash


Configuration
==============
Flytectl allows configuring using a YAML file or pass every configuration value
on command-line. The follow configuration is useful to setup.

Basic Configuration
--------------------

.. code-block:: yaml

  admin:
    # For GRPC endpoints you might want to use dns:///flyte.myexample.com
    endpoint: dns:///flyte.lyft.net
    # Change insecure flag to ensure that you use the right setting for your environment
    insecure: true
  # Logger settings to control logger output. Useful to debug
  #logger:
    #show-source: true
    #level: 1



.. toctree::
   :maxdepth: 1
   :caption: Flyte Core docs

   Flyte Documentation <https://lyft.github.io/flyte/>

.. toctree::
   :maxdepth: 2
   :caption: Flytectl docs - Entities
   
   tasks
   workflow
   launchplan


.. toctree::
   :maxdepth: 2
   :caption: Flytectl verbs
   
   get
   update
   delete
   register

.. toctree::
   :maxdepth: 2
   :caption: Contribute

   contribute


Indices and tables
==================

* :ref:`genindex`
* :ref:`modindex`
* :ref:`search`
