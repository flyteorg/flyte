"""
.. currentmodule:: flytekitplugins.kfpytorch

This package contains things that are useful when extending Flytekit.

.. autosummary::
   :template: custom.rst
   :toctree: generated/

   PyTorch
   Elastic
"""

from .task import CleanPodPolicy, Elastic, Master, PyTorch, RestartPolicy, RunPolicy, Worker
