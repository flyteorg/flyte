"""
.. currentmodule:: flytekitplugins.airflow

This package contains things that are useful when extending Flytekit.

.. autosummary::
   :template: custom.rst
   :toctree: generated/

   AirflowConfig
   AirflowTask
   AirflowAgent
"""

from .agent import AirflowAgent
from .task import AirflowObj, AirflowTask
