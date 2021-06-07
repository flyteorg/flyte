Hive
====

Flyte backend can be connected with various hive services. Once enabled it can allow you to query a hive service (e.g. Qubole) and retrieve typed schema (optionally).
This section will provide how to use the Hive Query Plugin using flytekit python

Installation
------------

To use the flytekit hive plugin simply run the following:

.. prompt:: bash

    pip install flytekitplugins-hive

No Need of a dockerfile
------------------------
This plugin is purely a spec and since SQL is completely portable has no need to build a container. Thus this plugin examples do not have any Dockerfile

.. TODO: write a subsection for "Configuring the backend to get hive working"

This is coming soon.
