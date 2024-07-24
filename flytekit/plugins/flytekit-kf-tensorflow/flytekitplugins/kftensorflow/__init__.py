"""
.. currentmodule:: flytekitplugins.kftensorflow

This package contains things that are useful when extending Flytekit.

.. autosummary::
   :template: custom.rst
   :toctree: generated/

   TfJob
"""

from .task import PS, Chief, CleanPodPolicy, Evaluator, RestartPolicy, RunPolicy, TfJob, Worker
