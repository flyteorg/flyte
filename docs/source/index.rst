.. flytectl doc

##########################################
``Flytectl`` Reference
##########################################


Install
=======
Flytectl is a Golang binary and can be installed on any platform supported by
golang. To install simply copy paste the following into the command-line:

.. prompt:: bash

   curl -s https://raw.githubusercontent.com/lyft/flytectl/master/install.sh | bash


Configure
=========
Flytectl allows configuring using a YAML file or pass every configuration value
on command-line. The follow configuration is useful to setup.
Place this in $HOME/.flyte directory with name config.yaml.
This file is searched in
* $HOME/.flyte
* currDir from where you run flytectl
* /etc/flyte/config
* You can pass it commandline using --config <config-file-path>

Basic Configuration
--------------------

.. code-block:: yaml

  admin:
    # For GRPC endpoints you might want to use dns:///flyte.myexample.com
    endpoint: dns:///flyte.lyft.net
    # Change insecure flag to ensure that you use the right setting for your environment
    insecure: true
  logger:
    # Logger settings to control logger output. Useful to debug logger:
    show-source: true
    level: 1


.. toctree::
   :maxdepth: 1
   :hidden:

   Getting Started <https://docs.flyte.org/en/latest/getting_started.html>
   User Guide <https://docs.flyte.org/projects/cookbook/en/latest/user_guide.html>
   Tutorials <https://docs.flyte.org/projects/cookbook/en/latest/tutorials.html>
   Concepts <https://docs.flyte.org/en/latest/concepts/basics.html>
   API Reference <https://docs.flyte.org/en/latest/reference/index.html>
   Community <https://docs.flyte.org/en/latest/community/index.html>

.. toctree::
   :maxdepth: -1
   :caption: Flytectl
   :hidden:

   Install and Configure <self>
   verbs
   nouns
   Contribute <https://github.com/flyteorg/flytectl/blob/master/docs/CONTRIBUTING.md>
