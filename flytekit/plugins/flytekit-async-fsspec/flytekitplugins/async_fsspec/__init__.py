"""
.. currentmodule:: flytekitplugins.async_fsspec

This package contains things that are useful when extending Flytekit.

.. autosummary::
   :template: custom.rst
   :toctree: generated/

   AsyncS3FileSystem
"""
import fsspec

from .s3fs.s3fs import AsyncS3FileSystem

fsspec.register_implementation("s3", AsyncS3FileSystem)
