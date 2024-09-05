=====================
Unit Testing
=====================

.. currentmodule:: flytekit.testing

The imports exposed in this package will help you unit test your Flyte tasks. These are particularly helpful when
testing workflows that contain tasks that cannot run locally (a Hive task for instance).

.. autosummary::
   :toctree: generated/

   patch - A decorator similar to the regular one you're probably used to
   task_mock - Non-decorative function
