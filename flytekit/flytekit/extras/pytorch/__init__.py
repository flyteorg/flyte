"""
.. currentmodule:: flytekit.extras.pytorch

.. autosummary::
   :template: custom.rst
   :toctree: generated/

    PyTorchCheckpoint
    PyTorchCheckpointTransformer
    PyTorchModuleTransformer
    PyTorchTensorTransformer
"""
from flytekit.loggers import logger

# TODO: abstract this out so that there's an established pattern for registering plugins
# that have soft dependencies
try:
    # isolate the exception to the torch import
    import torch

    _torch_installed = True
except (ImportError, OSError):
    _torch_installed = False


if _torch_installed:
    from .checkpoint import PyTorchCheckpoint, PyTorchCheckpointTransformer
    from .native import PyTorchModuleTransformer, PyTorchTensorTransformer
else:
    logger.info(
        "We won't register PyTorchCheckpointTransformer, PyTorchTensorTransformer, and PyTorchModuleTransformer because torch is not installed."
    )
