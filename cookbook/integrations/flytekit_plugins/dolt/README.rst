Dolt
====

.. tags:: Integration, Data, SQL, Intermediate

The ``DoltTable`` plugin is a wrapper that uses `Dolt <https://github.com/dolthub/dolt>`__ to move data between
``pandas.DataFrame``'s at execution time and database tables at rest.

Installation
------------
The dolt plugin and dolt command line tool are required to run these examples:

.. code:: bash

   pip install flytekitplugins.dolt
   sudo bash -c 'curl -L https://github.com/dolthub/dolt/releases/latest/download/install.sh | sudo bash'

Dolt requires a user configuration to run ``init``:

.. code:: bash

   dolt config --global --add user.email <email>
   dolt config --global --add user.name <name>

These demos assume a ``foo`` database has been created locally:

.. code:: bash

   mkdir foo
   cd foo
   dolt init

.. panels::
    :header: text-center
    :column: col-lg-12 p-2

    .. link-button:: https://blog.flyte.org/upleveling-flyte-data-lineage-using-dolt
       :type: url
       :text: Blog Post
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    An article detailing Dolt and Flyte integration.

.. toctree::
    :maxdepth: -1
    :caption: Contents
    :hidden:

    Blog Post <https://blog.flyte.org/upleveling-flyte-data-lineage-using-dolt>
