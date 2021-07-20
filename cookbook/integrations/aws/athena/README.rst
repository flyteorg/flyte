######
Athena
######

Executing Athena Queries
=======================
Flyte backend can be connected with athena. Once enabled, it allows you to query AWS Athena service (Presto + ANSI SQL Support) and retrieve typed schema (optionally).

This section provides a guide on how to use the AWS Athena Plugin using flytekit python.

Installation
------------

To use the flytekit athena plugin simply run the following:

.. prompt:: bash

    pip install flytekitplugins-athena

No Need for a dockerfile
------------------------
This plugin is purely a spec and since SQL is completely portable, it has no need to build a container. Thus this plugin examples do not have any Dockerfiles.

.. TODO: write a subsection for "Configuring the backend to get athena working"
=======
    Coming soon 🛠
