"""
#################
Fast Registration
#################

.. NOTE:: 
   Experimental feature (beta)

To avoid having to wait for an image build to test out simple code changes to your Flyte workflows and reduce iteration cycles to mere seconds, read on below.

Caveats
=======
Fast registration only works when you're testing out code changes. If you need to update your container by installing a dependency or modifying a Dockerfile, you **must** use the conventional method of committing your changes and rebuilding a container image.

.. note::

 Flytekit releases before the introduction of native typing have a slightly modified workflow for fast-registering. Please check the below-highlighted sections if your Flytkit version is < ``0.16.0``.

Pre-requisites
==============
Upgrade your Flytekit dependency to ``>= 0.16.0`` and re-run pip-tools if necessary.

.. admonition:: Flytkit < 0.16.0

 You'll need to update your development flyte config and add a new required parameter to the SDK block specifying an intermittent directory for code distributions. Whichever role you use in the [auth] block must have read access to this directory. For example:

 .. code-block:: bash

    [sdk]
    fast_registration_dir=s3://my-s3-bucket/distributions/<myproject>

You'll need to build a base image with these changes incorporated before you can use fast registration.

Steps to Follow
===============

1. After the above pre-requisite changes are merged, pull the latest master and create a development branch on your local machine.
2. Make some code changes. 
3. Save your files.
4. Clear and create (or, clear or create) the directory used to store your serialized code archive:

.. code-block:: bash 

   mkdir _pb_output || true
   rm -f _pb_output/*.tar.gz 

.. admonition:: Flytkit < 0.16.0

 Step 4 is not required.
   
5. Assume a role that has the write access to the intermittent directory you'll use to store fast registration code distributions.
6. Using a Python environment with Flytekit installed, fast-register your changes:

.. code-block:: bash 

   pyflyte -c sandbox.config --pkgs recipes serialize --in-container-config-path /root/sandbox.config \
       --local-source-root <path to your code> --image <older/existing docker image here> fast workflows -f _pb_output/

Or, from within your container:

.. code-block:: bash

    pyflyte --config /root/sandbox.config serialize fast workflows -f _pb_output/ 

.. admonition:: Flytkit < 0.16.0

    The command to fast-register the changes is:

    .. code-block:: bash

       flytekit_venv pyflyte -p myproject -d development -v <your-base-image> -c /code/myproject/development.config fast-register  workflows --source-dir /code/myproject/

7. Next, fast-register your serialized files. You'll note the overlap with the existing register command (auth role and output location) but with a new flag pointing to an additional distribution dir. This must be writable from the role you assume and readable from the role your Flytepropeller assumes.

.. code-block:: bash

    flyte-cli fast-register-files -p flytetester -d development --kubernetes-service-account ${FLYTE_AUTH_KUBERNETES_SERVICE_ACCOUNT} \
        --output-location-prefix ${FLYTE_AUTH_RAW_OUTPUT_DATA_PREFIX} -h ${FLYTE_PLATFORM_URL} \
        --additional-distribution-dir ${FLYTE_SDK_FAST_REGISTRATION_DIR} _pb_output/*

.. admonition:: Flytekit < 0.16.0

 Step 7 is not required.

8. Open the Flyte UI and launch the latest version of your workflow (under the domain you fast-registered above). It should run with your new code!

"""
