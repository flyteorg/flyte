"""
.. currentmodule:: flytekitplugins.papermill

This package contains things that are useful when extending Flytekit.

.. autosummary::
   :template: custom.rst
   :toctree: generated/

   NotebookTask
   record_outputs
"""

from .task import NotebookTask, load_flytedirectory, load_flytefile, load_structureddataset, record_outputs
