"""
Flytekit Directory Type
==========================================================
.. currentmodule:: flytekit.types.directory

Similar to :py:class:`flytekit.types.file.FlyteFile` there are some 'preformatted' directory types.

.. autosummary::
   :toctree: generated/
   :template: file_types.rst

   FlyteDirectory
   TensorboardLogs
   TFRecordsDirectory
"""

import typing

from .types import FlyteDirectory

# The following section provides some predefined aliases for commonly used FlyteDirectory formats.

tensorboard = typing.TypeVar("tensorboard")
TensorboardLogs = FlyteDirectory[tensorboard]
"""
    This type can be used to denote that the output is a folder that contains logs that can be loaded in TensorBoard.
    This is usually the SummaryWriter output in PyTorch or Keras callbacks which record the history readable by
    TensorBoard.
"""

tfrecords_dir = typing.TypeVar("tfrecords_dir")
TFRecordsDirectory = FlyteDirectory[tfrecords_dir]
"""
    This type can be used to denote that the output is a folder that contains tensorflow record files.
    This is usually the TFRecordWriter output in Tensorflow which writes serialised tf.train.Example
    message (or protobuf) to tfrecord files
"""
