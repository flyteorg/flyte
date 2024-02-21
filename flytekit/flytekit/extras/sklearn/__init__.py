"""
.. currentmodule:: flytekit.extras.sklearn

.. autosummary::
   :template: custom.rst
   :toctree: generated/

   SklearnEstimatorTransformer
"""
from flytekit.loggers import logger

# TODO: abstract this out so that there's an established pattern for registering plugins
# that have soft dependencies
try:
    # isolate the exception to the sklearn import
    import sklearn

    _sklearn_installed = True
except (ImportError, OSError):
    _sklearn_installed = False


if _sklearn_installed:
    from .native import SklearnEstimatorTransformer
else:
    logger.info("We won't register SklearnEstimatorTransformer because scikit-learn is not installed.")
