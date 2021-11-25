#############
Flyte Console
#############
|Current Release| |Build Status| |License| |CodeCoverage| |Slack|
    .. |Current Release| image:: https://img.shields.io/github/release/lyft/flyteconsole.svg
        :target: https://github.com/lyft/flyteconsole/releases/latest
        
    .. |Build Status| image:: https://travis-ci.org/lyft/flyteconsole.svg?branch=master
        :target: https://travis-ci.org/lyft/flyteconsole

    .. |License| image:: https://img.shields.io/badge/LICENSE-Apache2.0-ff69b4.svg
        :target: http://www.apache.org/licenses/LICENSE-2.0.html

    .. |CodeCoverage| image:: https://img.shields.io/codecov/c/github/lyft/flyteconsole.svg
        :target: https://codecov.io/gh/lyft/flyteconsole
   
    .. |Slack| image:: https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social
        :target: https://slack.flyte.org

This is the web UI for the Flyte platform.

*********************
Running flyteconsole
*********************

=====================
Install Dependencies
=====================
Running flyteconsole locally requires `NodeJS <https://nodejs.org>`_ and
`yarn <https://yarnpkg.com>`_. Once these are installed, all of the dependencies
can be installed by running ``yarn`` in the project directory.

====================
Quickstart
====================

#. Follow `Start a Local flyte backend <https://docs.flyte.org/en/latest/getting_started/first_run.html>`_, like

  .. code:: bash

     docker run --rm --privileged -p 30081:30081 -p 30082:30082 -p 30084:30084 cr.flyte.org/flyteorg/flyte-sandbox

#. Now Export these 2 env variables

   .. code:: bash

       export ADMIN_API_URL=http://localhost:30081
       export DISABLE_AUTH=1

   .. note::

     You can persist these environment variables in a ``.env`` file at the root
     of the repository. This will persist the settings across multiple terminal
     sessions

#. Start the server (uses localhost:3000)

   .. code:: bash

      yarn start

#. Explore your local copy: `http://localhost:3000`

======================
Environment variables
======================
Before we can run the server, we need to set up an environment variable or two.
Environment variables can be set either in the current shell or persisted in
``.env`` file stored under the root of the repository.

``ADMIN_API_URL`` (default: `window.location.origin <https://developer.mozilla.org/en-US/docs/Web/API/Window/location>`_)

The Flyte console displays information fetched from the Flyte Admin API. This
environment variable specifies the host prefix used in constructing API requests.

*Note*: this is only the host portion of the API endpoint, consisting of the
protocol, domain, and port (if not using the standard 80/443).

This value will be combined with a suffix (such as ``/api/v1``) to construct the
final URL used in an API request.

*Default Behavior*

In most cases, ``flyteconsole`` will be hosted in the same cluster as the Admin
API, meaning that the domain used to access the console is the same value used to
access the API. For this reason, if no value is set for ``ADMIN_API_URL``, the
default behavior is to use the value of `window.location.origin`.


``BASE_URL`` (default: ``undefined``)

This allows running the console at a prefix on the target host. This is
necessary when hosting the API and console on the same domain (with prefixes of
``/api/v1`` and ``/console`` for example). For local development, this is
usually not needed, so the default behavior is to run without a prefix.


``CORS_PROXY_PREFIX`` (default: ``/cors_proxy``)

Sets the local endpoint for `CORS request proxying <cors-proxying_>`_.

===============
Run the server
===============

To start the local development server, run ``yarn start``. This will spin up a
Webpack development server, compile all of the code into bundles, and start the
NodeJS server on the default port (3000). All requests to the NodeJS server will
be stalled until the bundles have finished. The application will be accessible
at http://localhost:3000 (if using the default port).

************
Development
************

==========
Storybook
==========

This project has support for `Storybook <https://storybook.js.org/>`_.
Component stories live next to the components they test, in a ``__stories__``
directory, with the filename pattern ``{Component}.stories.tsx``.

You can run storybook with ``yarn run storybook``, and view the stories at http://localhost:9001.

=============================
Protobuf and the Network tab
=============================

Communication with the Flyte Admin API is done using Protobuf as the
request/response format. Protobuf is a binary format, which means looking at
responses in the Network tab won't be very helpful. To make debugging easier,
each network request is logged to the console with it's URL followed by the
decoded Protobuf payload. You must have debug output enabled (on by default in
development) to see these messages.

============
Debug Output
============

This application makes use of the `debug <https://github.com/visionmedia/debug>`_
libary to provide namespaced debug output in the browser console. In
development, all debug output is enabled. For other environments, the debug
output must be enabled manually. You can do this by setting a flag in
localStorage using the console: ``localStorage.debug = 'flyte:*'``. Each module in
the application sets its own namespace. So if you'd like to only view output for
a single module, you can specify that one specifically
(ex. ``localStorage.debug = 'flyte:adminEntity'`` to only see decoded Flyte
Admin API requests).

.. _cors-proxying:

================================
CORS Proxying: Recommended setup
================================

In the common hosting arrangement, all API requests will be to the same origin
serving the client application, making CORS unnecessary. However, if you would like
to setup your local dev enviornment to target a FlyteAdmin service running on a different
domain you will need to configure your enviornment support CORS. One example would be
hosting the Admin API on a different domain than the console. Another example is
when fetching execution data from external storage such as S3.

The fastest (recommended) way to setup a CORS solution is to do so within the browser. 
If you would like to handle this at the Node level you will need to disable authentication
(see below)

   .. note:: Do not configure for both browser and Node solutions. 

These instructions require using Google Chrome. You will also need to identify the 
URL of your target FlyteAdmin API instance. These instructions will use
`https://different.admin.service.com` as an example.


#. Set `ADMIN_API_URL` and `ADMIN_API_USE_SSL`
   
   .. code:: bash

      export ADMIN_API_URL=https://different.admin.service.com
      export ADMIN_API_USE_SSL="https"

   .. note:: Hint
      Add these to your local profile (eg, `./profile`) to prevent having to do this step each time

#. Generate SSL certificate

   Run the following command from your `flyteconsole` directory

   .. code:: bash

      make generate_ssl


#. Add new record to hosts file

   .. code:: bash
      
      sudo vim /etc/hosts

   Add the following record
   
   .. code:: bash
   
      127.0.0.1 localhost.different.admin.service.com

#. Install Chrome plugin: `Allow CORS: Access-Control-Allow-Origin <https://chrome.google.com/webstore/detail/allow-cors-access-control/lhobafahddgcelffkeicbaginigeejlf>`_

      .. note:: Activate plugin (toggle to "on")

#. Start `flyteconsole`

   .. code:: bash

      yarn start

   Your new localhost is `localhost.different.admin.service.com <http://localhost.different.admin.service.com>`_

.. note:: Hint

   Ensure you don't have `ADMIN_API_URL` or `DISABLE_AUTH` set (eg, in your `/.profile`.)

=======
Release
=======
To release, you have to annotate the PR message to include either #minor,
#patch or #major.
