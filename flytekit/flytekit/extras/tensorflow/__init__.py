"""
.. currentmodule:: flytekit.extras.tensorflow

.. autosummary::
   :template: custom.rst
   :toctree: generated/

   TensorFlowRecordFileTransformer
   TensorFlowRecordsDirTransformer
"""

from flytekit.loggers import logger

# TODO: abstract this out so that there's an established pattern for registering plugins
# that have soft dependencies
try:
    # isolate the exception to the tensorflow import
    import tensorflow

    _tensorflow_installed = True
except TypeError as e:
    logger.warn(f"Unsupported version of tensorflow installed. Error message from protobuf library: {e}")
    _tensorflow_installed = False
except (ImportError, OSError):
    _tensorflow_installed = False


if _tensorflow_installed:
    from .model import TensorFlowModelTransformer
    from .record import TensorFlowRecordFileTransformer, TensorFlowRecordsDirTransformer
else:
    logger.info(
        "We won't register TensorFlowRecordFileTransformer, TensorFlowRecordsDirTransformer and TensorFlowModelTransformer"
        "because tensorflow is not installed."
    )
