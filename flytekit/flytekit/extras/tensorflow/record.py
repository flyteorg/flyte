import os
from dataclasses import dataclass
from typing import Optional, Tuple, Type, Union

import tensorflow as tf
from dataclasses_json import DataClassJsonMixin
from tensorflow.python.data.ops.readers import TFRecordDatasetV2
from typing_extensions import Annotated, get_args, get_origin

from flytekit.core.context_manager import FlyteContext
from flytekit.core.type_engine import TypeEngine, TypeTransformer, TypeTransformerFailedError
from flytekit.models.core import types as _core_types
from flytekit.models.literals import Blob, BlobMetadata, Literal, Scalar
from flytekit.models.types import LiteralType
from flytekit.types.directory import TFRecordsDirectory
from flytekit.types.file import TFRecordFile


@dataclass
class TFRecordDatasetConfig(DataClassJsonMixin):
    """
    TFRecordDatasetConfig can be used while creating tf.data.TFRecordDataset comprising
    record of one or more TFRecord files.

    Args:
      compression_type: A scalar evaluating to one of "" (no compression), "ZLIB", or "GZIP".
      buffer_size: The number of bytes in the read buffer. If None, a sensible default for both local and remote file systems is used.
      num_parallel_reads: The number of files to read in parallel. If greater than one, the record of files read in parallel are outputted in an interleaved order.
      name: A name for the operation.
    """

    compression_type: Optional[str] = None
    buffer_size: Optional[int] = None
    num_parallel_reads: Optional[int] = None
    name: Optional[str] = None


def extract_metadata_and_uri(
    lv: Literal, t: Type[Union[TFRecordFile, TFRecordsDirectory]]
) -> Tuple[Union[TFRecordFile, TFRecordsDirectory], TFRecordDatasetConfig]:
    try:
        uri = lv.scalar.blob.uri
    except AttributeError:
        TypeTransformerFailedError(f"Cannot convert from {lv} to {t}")
    metadata = TFRecordDatasetConfig()
    if get_origin(t) is Annotated:
        _, metadata = get_args(t)
        if isinstance(metadata, TFRecordDatasetConfig):
            return uri, metadata
        else:
            raise TypeTransformerFailedError(f"{t}'s metadata needs to be of type TFRecordDatasetConfig")
    return uri, metadata


class TensorFlowRecordFileTransformer(TypeTransformer[TFRecordFile]):
    """
    TypeTransformer that supports serialising and deserialising to and from TFRecord file.
    https://www.tensorflow.org/tutorials/load_data/tfrecord
    """

    TENSORFLOW_FORMAT = "TensorFlowRecord"

    def __init__(self):
        super().__init__(name="TensorFlow Record File", t=TFRecordFile)

    def get_literal_type(self, t: Type[TFRecordFile]) -> LiteralType:
        return LiteralType(
            blob=_core_types.BlobType(
                format=self.TENSORFLOW_FORMAT,
                dimensionality=_core_types.BlobType.BlobDimensionality.SINGLE,
            )
        )

    def to_literal(
        self, ctx: FlyteContext, python_val: TFRecordFile, python_type: Type[TFRecordFile], expected: LiteralType
    ) -> Literal:
        meta = BlobMetadata(
            type=_core_types.BlobType(
                format=self.TENSORFLOW_FORMAT,
                dimensionality=_core_types.BlobType.BlobDimensionality.SINGLE,
            )
        )
        local_dir = ctx.file_access.get_random_local_directory()
        local_path = os.path.join(local_dir, "0000.tfrecord")
        with tf.io.TFRecordWriter(local_path) as writer:
            writer.write(python_val.SerializeToString())
        remote_path = ctx.file_access.put_raw_data(local_path)
        return Literal(scalar=Scalar(blob=Blob(metadata=meta, uri=remote_path)))

    def to_python_value(
        self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[TFRecordFile]
    ) -> TFRecordDatasetV2:
        uri, metadata = extract_metadata_and_uri(lv, expected_python_type)
        local_path = ctx.file_access.get_random_local_path()
        ctx.file_access.get_data(uri, local_path, is_multipart=False)
        filenames = [local_path]
        return tf.data.TFRecordDataset(
            filenames=filenames,
            compression_type=metadata.compression_type,
            buffer_size=metadata.buffer_size,
            num_parallel_reads=metadata.num_parallel_reads,
            name=metadata.name,
        )

    def guess_python_type(self, literal_type: LiteralType) -> Type[TFRecordFile]:
        if (
            literal_type.blob is not None
            and literal_type.blob.dimensionality == _core_types.BlobType.BlobDimensionality.SINGLE
            and literal_type.blob.format == self.TENSORFLOW_FORMAT
        ):
            return TFRecordFile

        raise ValueError(f"Transformer {self} cannot reverse {literal_type}")


class TensorFlowRecordsDirTransformer(TypeTransformer[TFRecordsDirectory]):
    """
    TypeTransformer that supports serialising and deserialising to and from TFRecord directory.
    https://www.tensorflow.org/tutorials/load_data/tfrecord
    """

    TENSORFLOW_FORMAT = "TensorFlowRecord"

    def __init__(self):
        super().__init__(name="TensorFlow Record Directory", t=TFRecordsDirectory)

    def get_literal_type(self, t: Type[TFRecordsDirectory]) -> LiteralType:
        return LiteralType(
            blob=_core_types.BlobType(
                format=self.TENSORFLOW_FORMAT,
                dimensionality=_core_types.BlobType.BlobDimensionality.MULTIPART,
            )
        )

    def to_literal(
        self,
        ctx: FlyteContext,
        python_val: TFRecordsDirectory,
        python_type: Type[TFRecordsDirectory],
        expected: LiteralType,
    ) -> Literal:
        meta = BlobMetadata(
            type=_core_types.BlobType(
                format=self.TENSORFLOW_FORMAT,
                dimensionality=_core_types.BlobType.BlobDimensionality.MULTIPART,
            )
        )
        local_dir = ctx.file_access.get_random_local_directory()
        for i, val in enumerate(python_val):
            local_path = f"{local_dir}/part_{i}.tfrecord"
            with tf.io.TFRecordWriter(local_path) as writer:
                writer.write(val.SerializeToString())
        remote_path = ctx.file_access.put_raw_data(local_dir)
        return Literal(scalar=Scalar(blob=Blob(metadata=meta, uri=remote_path)))

    def to_python_value(
        self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[TFRecordsDirectory]
    ) -> TFRecordDatasetV2:
        uri, metadata = extract_metadata_and_uri(lv, expected_python_type)
        local_dir = ctx.file_access.get_random_local_directory()
        ctx.file_access.get_data(uri, local_dir, is_multipart=True)
        files = os.scandir(local_dir)
        filenames = [os.path.join(local_dir, f.name) for f in files]
        return tf.data.TFRecordDataset(
            filenames=filenames,
            compression_type=metadata.compression_type,
            buffer_size=metadata.buffer_size,
            num_parallel_reads=metadata.num_parallel_reads,
            name=metadata.name,
        )

    def guess_python_type(self, literal_type: LiteralType) -> Type[TFRecordsDirectory]:
        if (
            literal_type.blob is not None
            and literal_type.blob.dimensionality == _core_types.BlobType.BlobDimensionality.MULTIPART
            and literal_type.blob.format == self.TENSORFLOW_FORMAT
        ):
            return TFRecordsDirectory

        raise ValueError(f"Transformer {self} cannot reverse {literal_type}")


TypeEngine.register(TensorFlowRecordsDirTransformer())
TypeEngine.register(TensorFlowRecordFileTransformer())
