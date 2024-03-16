"""
Flytekit File Type
==========================================================
.. currentmodule:: flytekit.types.file

This list also contains a bunch of pre-formatted :py:class:`flytekit.types.file.FlyteFile` types.

.. autosummary::
   :toctree: generated/
   :template: file_types.rst

   FlyteFile
   HDF5EncodedFile
   HTMLPage
   JoblibSerializedFile
   JPEGImageFile
   PDFFile
   PNGImageFile
   PythonPickledFile
   PythonNotebook
   SVGImageFile
"""
import typing

from typing_extensions import Annotated, get_args, get_origin

from .file import FlyteFile


class FileExt:
    """
    Used for annotating file extension types of FlyteFile.
    This is useful for extensions that have periods in them, e.g., "tar.gz".

    Example:
    TAR_GZ = Annotated[str, FileExt("tar.gz")]
    """

    def __init__(self, ext: str):
        self._ext = ext

    def __str__(self):
        return self._ext

    def __repr__(self):
        return self._ext

    @staticmethod
    def check_and_convert_to_str(item: typing.Union[typing.Type, str]) -> str:
        if get_origin(item) is not Annotated:
            return str(item)
        if get_args(item)[0] == str:
            return str(get_args(item)[1])
        raise ValueError("Underlying type of File Extension must be of type <str>")


# The following section provides some predefined aliases for commonly used FlyteFile formats.
# This makes their usage extremely simple for the users. Please keep the list sorted.

hdf5 = Annotated[str, FileExt("hdf5")]
#: This can be used to denote that the returned file is of type hdf5 and can be received by other tasks that
#: accept an hdf5 format. This is usually useful for serializing Tensorflow models
HDF5EncodedFile = FlyteFile[hdf5]

html = Annotated[str, FileExt("html")]
#: Can be used to receive or return an HTMLPage. The underlying type is a FlyteFile type. This is just a
#: decoration and useful for attaching content type information with the file and automatically documenting code.
HTMLPage = FlyteFile[html]

joblib = Annotated[str, FileExt("joblib")]
#: This File represents a file that was serialized using `joblib.dump` method can be loaded back using `joblib.load`.
JoblibSerializedFile = FlyteFile[joblib]

jpeg = Annotated[str, FileExt("jpeg")]
#: Can be used to receive or return an JPEGImage. The underlying type is a FlyteFile type. This is just a
#: decoration and useful for attaching content type information with the file and automatically documenting code.
JPEGImageFile = FlyteFile[jpeg]

pdf = Annotated[str, FileExt("pdf")]
#: Can be used to receive or return an PDFFile. The underlying type is a FlyteFile type. This is just a
#: decoration and useful for attaching content type information with the file and automatically documenting code.
PDFFile = FlyteFile[pdf]

png = Annotated[str, FileExt("png")]
#: Can be used to receive or return an PNGImage. The underlying type is a FlyteFile type. This is just a
#: decoration and useful for attaching content type information with the file and automatically documenting code.
PNGImageFile = FlyteFile[png]

python_pickle = Annotated[str, FileExt("python_pickle")]
#: This type can be used when a serialized Python pickled object is returned and shared between tasks. This only
#: adds metadata to the file in Flyte, but does not really carry any object information.
PythonPickledFile = FlyteFile[python_pickle]

ipynb = Annotated[str, FileExt("ipynb")]
#: This type is used to identify a Python notebook file.
PythonNotebook = FlyteFile[ipynb]

svg = Annotated[str, FileExt("svg")]
#: Can be used to receive or return an SVGImage. The underlying type is a FlyteFile type. This is just a
#: decoration and useful for attaching content type information with the file and automatically documenting code.
SVGImageFile = FlyteFile[svg]

csv = Annotated[str, FileExt("csv")]
#: Can be used to receive or return a CSVFile. The underlying type is a FlyteFile type. This is just a
#: decoration and useful for attaching content type information with the file and automatically documenting code.
CSVFile = FlyteFile[csv]

onnx = Annotated[str, FileExt("onnx")]
#: Can be used to receive or return an ONNXFile. The underlying type is a FlyteFile type. This is just a
#: decoration and useful for attaching content type information with the file and automatically documenting code.
ONNXFile = FlyteFile[onnx]

tfrecords_file = Annotated[str, FileExt("tfrecord")]
#: Can be used to receive or return an TFRecordFile. The underlying type is a FlyteFile type. This is just a
#: decoration and useful for attaching content type information with the file and automatically documenting code.
TFRecordFile = FlyteFile[tfrecords_file]
