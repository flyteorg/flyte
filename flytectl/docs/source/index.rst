.. flytectl doc

####################################################
Flytectl: The Official Flyte Command-line Interface
####################################################

Overview
=========
This video will take you on a tour of Flytectl - how to install and configure it, as well as how to use the Verbs and Nouns sections on the left hand side menu. Detailed information can be found in the sections below the video.

.. youtube:: cV8ezYnBANE


Installation
============

Flytectl is a Golang binary that can be installed on any platform supported by Golang.

.. tabbed:: OSX

  .. prompt:: bash $

      brew install flyteorg/homebrew-tap/flytectl

  *Upgrade* existing installation using the following command:

  .. prompt:: bash $

      flytectl upgrade

.. tabbed:: Other Operating systems

  .. prompt:: bash $

      curl -sL https://ctl.flyte.org/install | bash

  *Upgrade* existing installation using the following command:

  .. prompt:: bash $

      flytectl upgrade

**Test** if Flytectl is installed correctly (your Flytectl version should be > 0.2.0) using the following command:

.. prompt:: bash $

  flytectl version

Configuration
=============

Flytectl allows you to communicate with FlyteAdmin using a YAML file or by passing every configuration value
on the command-line. The following configuration can be used for the setup:

Basic Configuration
--------------------

The full list of available configurable options can be found by running ``flytectl --help``, or `here <https://docs.flyte.org/projects/flytectl/en/stable/gen/flytectl.html#synopsis>`__.

.. NOTE::

    Currently, the Project ``-p``, Domain ``-d``, and Output ``-o`` flags cannot be used in the config file.

.. tabbed:: Local Flyte Sandbox

    Automatically configured for you by ``flytectl sandbox`` command.

    .. code-block:: yaml

        admin:
          # For GRPC endpoints you might want to use dns:///flyte.myexample.com
          endpoint: dns:///localhost:30081
          insecure: true # Set to false to enable TLS/SSL connection (not recommended except on local sandbox deployment).
          authType: Pkce # authType: Pkce # if using authentication or just drop this.
        storage:
          connection:
            access-key: minio
            auth-type: accesskey
            disable-ssl: true
            endpoint: http://localhost:30084
            region: my-region-here
            secret-key: miniostorage
          container: my-s3-bucket
          type: minio

.. tabbed:: AWS Configuration

    .. code-block:: yaml

        admin:
          # For GRPC endpoints you might want to use dns:///flyte.myexample.com
          endpoint: dns:///<replace-me>
          authType: Pkce # authType: Pkce # if using authentication or just drop this.
          insecure: true # insecure: True # Set to true if the endpoint isn't accessible through TLS/SSL connection (not recommended except on local sandbox deployment)
        storage:
          type: stow
          stow:
            kind: s3
            config:
                auth_type: iam
                region: <REGION> # Example: us-east-2
          container: <replace> # Example my-bucket. Flyte k8s cluster / service account for execution should have read access to this bucket

.. tabbed:: GCS Configuration

    .. code-block:: yaml

        admin:
          # For GRPC endpoints you might want to use dns:///flyte.myexample.com
          endpoint: dns:///<replace-me>
          authType: Pkce # authType: Pkce # if using authentication or just drop this.
          insecure: false # insecure: True # Set to true if the endpoint isn't accessible through TLS/SSL connection (not recommended except on local sandbox deployment)
        storage:
          type: stow
          stow:
            kind: google
            config:
                json: ""
                project_id: <replace-me> # TODO: replace <project-id> with the GCP project ID
                scopes: https://www.googleapis.com/auth/devstorage.read_write
          container: <replace> # Example my-bucket. Flyte k8s cluster / service account for execution should have access to this bucket

.. tabbed:: Others

    For other supported storage backends like Oracle, Azure, etc., refer to the configuration structure `here <https://pkg.go.dev/github.com/flyteorg/flytestdlib/storage#Config>`__.

    Place the config file in ``$HOME/.flyte`` directory with the name config.yaml.
    This file is typically searched in:

    * ``$HOME/.flyte``
    * currDir from where you run flytectl
    * ``/etc/flyte/config``
    
    You can also pass the file name in the command line using ``--config <config-file-path>``.


.. toctree::
   :maxdepth: 1
   :hidden:

   |plane| Getting Started <https://docs.flyte.org/en/latest/getting_started.html>
   |book-reader| User Guide <https://docs.flyte.org/projects/cookbook/en/latest/user_guide.html>
   |chalkboard| Tutorials <https://docs.flyte.org/projects/cookbook/en/latest/tutorials.html>
   |project-diagram| Concepts <https://docs.flyte.org/en/latest/concepts/basics.html>
   |rocket| Deployment <https://docs.flyte.org/en/latest/deployment/index.html>
   |book| API Reference <https://docs.flyte.org/en/latest/reference/index.html>
   |hands-helping| Community <https://docs.flyte.org/en/latest/community/index.html>


.. NOTE: the caption text is important for the sphinx theme to correctly render the nav header
.. https://github.com/flyteorg/furo
.. toctree::
   :maxdepth: -1
   :caption: Flytectl
   :hidden:

   Flytectl <self>
   CLI Entrypoint <gen/flytectl>
   verbs
   nouns
   contribute
