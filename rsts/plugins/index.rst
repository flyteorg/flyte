.. _plugins:

################
Extending Flyte
################

.. _plugins_howto:

Flyte as platform was designed with extensibility as a core primitive. Flyte is essentially an integration framework and hence extensibility is possible through-out the system.
The following sections will guide you through writing your own extensions - either private or public (contribute back to the community).

.. toctree::
   :maxdepth: 1
   :name: howtoextendtoc

   extend/intro
   extend/flytekit_python
   extend/flyte_backend


====================
Available Extensions
====================
The following is a list of maintained plugins for Flyte and guides on how to install / use them.

.. toctree::
   :maxdepth: 1
   :name: pluginstoc

   spark_k8s
   pod
   sqlite3
   pandera
   papermill
   hive
   aws/index
   kubeflow/index
