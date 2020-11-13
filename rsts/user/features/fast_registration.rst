.. _fast_registration:


*****************
Fast Registration
*****************


Are you frustrated by having to wait for an image build in order to test out simple code changes to your Flyte workflows? If you're interested in reducing to your iteration cycle to mere seconds, read on below.

Caveats
#######

Fast-registration only works when you're testing out code changes. If you need to update your container, say by installing a dependency or modifying a Dockerfile, you **must** use the conventional method of committing your changes and rebuilding a container image.

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

