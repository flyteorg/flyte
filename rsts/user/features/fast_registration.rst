.. _fast_registration:


*****************
Fast Registration
*****************


Are you frustrated by having to wait for an image build in order to test out simple code changes to your Flyte workflows? If you're interested in reducing to your iteration cycle to mere seconds, read on below.

Caveats
=======

Fast-registration only works when you're testing out code changes. If you need to update your container, say by installing a dependency or modifying a Dockerfile, you **must** use the conventional method of committing your changes and rebuilding a container image.

Prerequisites
=============

* Upgrade your flytekit dependency to ``>=0.16.0`` and re-run piptools if necessary


You'll need to build a base image with these changes incorporated before you can use fast registration.


Fast-registering
================

How-to:

* After the above prerequisite changes are merged, pull the latest master and create a development branch on your local machine.
* Make some code changes. Save your files.
* Clear and/or create the directory used to store your serialized code archive:

.. code-block:: text 

   mkdir _pb_output || true
   rm -f _pb_output/*.tar.gz 

* Using a python environment with flytekit installed fast-register your changes:

.. code-block:: python 

   pyflyte -c sandbox.config --pkgs recipes serialize --in-container-config-path /root/sandbox.config \
       --local-source-root <path to your code> --image <older/existing docker image here> fast workflows -f _pb_output/
 

Or, from within your container:

 .. code-block:: python 

    pyflyte --config /root/sandbox.config serialize fast workflows -f _pb_output/ 

* Assume a role that has write access to the intermittent directory you'll use to store fast registration code distributions .
* Fast-register your serialized files. You'll note the overlap with the existing register command (auth role and output location)
   but with an new flag pointing to an additional distribution dir. This must be writable from the role you assume and readable from
   the role your flytepropeller assumes:

 .. code-block:: python 

    flyte-cli fast-register-files -p flytetester -d development --kubernetes-service-account ${FLYTE_AUTH_KUBERNETES_SERVICE_ACCOUNT} \
        --output-location-prefix ${FLYTE_AUTH_RAW_OUTPUT_DATA_PREFIX} -h ${FLYTE_PLATFORM_URL} \
        --additional-distribution-dir ${FLYTE_SDK_FAST_REGISTRATION_DIR} _pb_output/*
 

* Open the Flyte UI and launch the latest version of your workflow (under the domain you fast-registered above). It should run with your new code!


Older flytekit instructions (flytekit <0.16.0)
==============================================

Flytekit releases prior to the introduction of native typing have a slightly modified workflow for fast-registering.

Prerequisites
#############

* Upgrade your flytekit dependency to ``>=0.15.0``.

* Update your development flyte config and add a new required parameter to the sdk block specifying an intermittent directory for code distributions. Whichever role you use in the [auth] block must have read access to
this directory. For example:

.. code-block:: text

   [sdk]
   fast_registration_dir=s3://my-s3-bucket/distributions/<myproject>

You'll need to build a base image with these changes incorporated before you can use fast registration.

Fast-registering
################

How-to:

#. After the above prerequisite changes are merged, pull the latest master and create a development branch on your local machine.
#. Make some code changes. Save your files.
#. Assume a role that has write access to the intermittent directory you'll use to store fast registration code distributions (specified in your flytekit config above).
#. Using a python environment with flytekit installed fast-register your changes: ``# flytekit_venv pyflyte -p myproject -d development -v <your-base-image> -c /code/myproject/development.config fast-register  workflows --source-dir /code/myproject/``
#. Open the Flyte UI and launch the latest version of your workflow (under the domain you fast-registered above). It should run with your new code!

