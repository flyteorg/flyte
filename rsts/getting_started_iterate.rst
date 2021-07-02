.. _gettingstarted_iterate:

Getting Started: Iterate
-------------------------

.. raw:: html
  
    <p style="color: #808080; font-weight: 500; font-size: 20px; padding-top: 10px;">A step-by-step guide to building, deploying, and iterating on Flyte tasks and workflows</p>

.. panels::
    :body: text-justify
    :container: container-xs
    :column: col-lg-4 col-md-4 col-sm-6 col-xs-12 p-2

    ---
    .. link-button:: gettingstarted_implement
        :type: ref
        :text: ✔ Implement
        :classes: btn-outline-success btn-block stretched-link
    ---
    .. link-button:: gettingstarted_scale
            :type: ref
            :text: ✔ Scale
            :classes: btn-outline-success btn-block stretched-link
    ---
    .. link-button:: gettingstarted_iterate
            :type: ref
            :text: ➔ Iterate
            :classes: btn-outline-primary btn-block stretched-link


3. Iterate on your ideas
==========================


Modify code - Test
^^^^^^^^^^^^^^^^^^^
#. Open ``example.py`` in your favorite editor.

   .. code-block::

       myapp/workflows/example.py

   .. dropdown:: myapp/workflows/example.py

      .. rli:: https://raw.githubusercontent.com/flyteorg/flytekit-python-template/simplify-template/myapp/workflows/example.py
         :language: python

#. Add ``name: str`` as an argument to both ``my_wf`` and ``say_hello`` functions. Then update the body of ``say_hello`` to consume that argument.

   .. code-block:: python

     @task
     def say_hello(name: str) -> str:
         return f"hello world, {name}"

   .. code-block:: python

     @workflow
     def my_wf(name: str) -> str:
         res = say_hello(name=name)
         return res

#. Update the simple test at the bottom of the file to pass in a name, e.g.

   .. code-block:: python

     print(f"Running my_wf(name='adam') {my_wf(name='adam')}")

#. When you run this file locally, it should output ``hello world, adam``.

   .. prompt:: bash (venv)$

     python myapp/workflows/example.py


   .. dropdown:: Expected output

       .. prompt:: text

            Running my_wf(name='adam') hello world, adam

#. Flytekit also provides ways to programmatically tests your tasks and workflows and provides easy mocking hooks to mock out certain functions.


[Bonus] Build & Deploy Your Application "Fast"er!
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
#. To deploy this workflow to the Flyte cluster (sandbox), you can repeat the steps previously covered in :ref:`getting-started-build-deploy`. Flyte provides a **faster** way to iterate on your workflows. Since you have not updated any of the dependencies in your requirements file, it is possible to push just the code to Flyte backend without re-building the entire Docker container. To do so, run the following commands.

   .. prompt:: bash (venv)$

       pyflyte --pkgs myapp.workflows package --image myapp:v1 --fast --force

   .. note::

      ``--fast`` flag will take the code from your local machine and provide it for execution without having to build the container and push it. The ``--force`` flag allows overriding your previously created package.

   .. caution::

      The ``fast`` registration method can only be used if you do not modify any requirements (that is, you re-use an existing environment). But, if you add a dependency to your requirements file or env you have to follow the :ref:`getting-started-build-deploy` method.

#. The code can now be deployed using Flytectl, similar to what we've done previously. ``flytectl`` automatically understands that the package is for ``fast`` registration.
   For this to work, a new ``storage`` block has to be added to the Flytectl configuration with appropriate permissions at runtime. The storage block configures Flytectl to write to a specific ``S3 / GCS bucket``. If you're using the sandbox, this is automatically configured by Flytectl, so you can skip this for now. But do take a note for the future.

   .. prompt:: bash $

       flytectl register files -p flytesnacks -d development -a flyte-package.tgz  -v v1-fast1

   .. tabs:: Flytectl configuration with ``storage`` block for Fast registration

       .. tab:: Local Flyte Sandbox

           Automatically configured for you by ``flytectl sandbox`` command.

           .. code-block:: yaml

               admin:
                 # For GRPC endpoints you might want to use dns:///flyte.myexample.com
                 endpoint: dns:///localhost:30081
                 insecure: true
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

       .. tab:: S3 Configuration

           .. code-block:: yaml

               admin:
                 # For GRPC endpoints you might want to use dns:///flyte.myexample.com
                 endpoint: dns:///<replace-me>
                 authType: Pkce # if using authentication or just drop this. If insecure set insecure: True
               storage:
                 kind: s3
                 config:
                   auth_type: iam
                   region: <replace> # Example: us-east-2
                 container: <replace> # Example my-bucket. Flyte k8s cluster / service account for execution should have read access to this bucket

       .. tab:: GCS Configuration

           .. code-block:: yaml

               admin:
                 # For GRPC endpoints you might want to use dns:///flyte.myexample.com
                 endpoint: dns:///<replace-me>
                 authType: Pkce # if using authentication or just drop this. If insecure set insecure: True
               storage:
                 kind: google
                 config:
                   json: ""
                   project_id: <replace-me> # TODO: replace <project-id> with the GCP project ID
                   scopes: https://www.googleapis.com/auth/devstorage.read_write
                 container: <replace> # Example my-bucket. Flyte k8s cluster / service account for execution should have access to this bucket

       .. tab:: Others

               For other supported storage backends like Oracle, Azure, etc., refer to the configuration structure `here <https://pkg.go.dev/github.com/flyteorg/flytestdlib/storage#Config>`__.


#. Finally, visit `the sandbox console <http://localhost:30081/console/projects/flytesnacks/domains/development/workflows/myapp.workflows.example.my_wf>`__, click launch, and give your name as the input.

   Alternatively, use flytectl. To pass arguments to the workflow update the execution spec file (previously generated) to send arguments to the workflow. ::

        ....
        inputs:
          name: "adam"
        ....

   Create an execution using the exec spec file.

   .. prompt:: bash $

      flytectl create execution -p flytesnacks -d development --execFile exec_spec.yaml



.. admonition:: Recap

  🎉 Congratulations! You have now

  1. Run a Flyte workflow locally,
  2. Started a Flyte sanbox cluster,
  3. Run a Flyte workflow on a cluster,
  4. Iterated on a Flyte workflow.


Next Steps: User Guide
========================

To experience the full capabilities of Flyte, take a look at the `User Guide <https://docs.flyte.org/projects/cookbook/en/latest/user_guide.html>`__.

.. toctree::
   :maxdepth: -1
   :caption: Getting Started
   :hidden:

   User Guide <https://docs.flyte.org/projects/cookbook/en/latest/user_guide.html>