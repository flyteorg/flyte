##########
AWS Athena
##########

Executing Athena Queries
========================
Flyte backend can be connected with Athena. Once enabled, it allows you to query AWS Athena service (Presto + ANSI SQL Support) and retrieve typed schema (optionally).
This plugin is purely a spec and since SQL is completely portable, it has no need to build a container. Thus this plugin example does not have any Dockerfile.

Installation
------------

To use the flytekit Athena plugin, simply run the following:

.. prompt:: bash

    pip install flytekitplugins-athena

Now let's dive into the code.
