.. flytectl doc

##########################################
``Flytectl`` Reference
##########################################

Overview
=========

.. youtube:: qUKf07uUiTU


Install
=======
Flytectl is a Golang binary and can be installed on any platform supported by
golang


.. tabs::

  .. tab:: OSX

    .. prompt:: bash $

       brew install flyteorg/homebrew-tap/flytectl

    *Upgrade* existing installation using the following command:

    .. prompt:: bash $

       brew update && brew upgrade flytectl

  .. tab:: Other Operating systems

    .. prompt:: bash $

        curl -s https://raw.githubusercontent.com/lyft/flytectl/master/install.sh | bash

**Test** if Flytectl is installed correctly (your Flytectl version should be > 0.2.0) using the following command:

.. prompt:: bash $

  flytectl version

Configure
=========
Flytectl allows configuring using a YAML file or pass every configuration value
on command-line. The following configuration is useful to setup.

Basic Configuration
--------------------

.. tabs:: Flytectl configuration

   .. tab:: Local Flyte Sandbox

       Automatically configured for you by ``flytectl sandbox`` command.

       .. code-block:: yaml

           admin:
             # For GRPC endpoints you might want to use dns:///flyte.myexample.com
             endpoint: dns:///localhost:30081
             insecure: false # insecure: True # Set to true if the endpoint isn't accessible through TLS/SSL connection (not recommended except on local sandbox deployment)
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

   .. tab:: AWS Configuration

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

   .. tab:: GCS Configuration

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

   .. tab:: Others

           For other supported storage backends like Oracle, Azure, etc., refer to the configuration structure `here <https://pkg.go.dev/github.com/flyteorg/flytestdlib/storage#Config>`__.


    Place this in $HOME/.flyte directory with name config.yaml.
    This file is searched in

    * $HOME/.flyte
    * currDir from where you run flytectl
    * /etc/flyte/config
        You can pass it commandline using --config <config-file-path> aswell


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

.. toctree::
   :maxdepth: -1
   :caption: Flytectl
   :hidden:

   Install and Configure <self>
   verbs
   nouns
   Contribute <https://github.com/flyteorg/flytectl/blob/master/docs/CONTRIBUTING.md>
