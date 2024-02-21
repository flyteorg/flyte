from typing import Dict, Tuple

import numpy as np
import tensorflow as tf
from tensorflow.python.data.ops.readers import TFRecordDatasetV2
from typing_extensions import Annotated

from flytekit import task, workflow
from flytekit.extras.tensorflow.record import TFRecordDatasetConfig
from flytekit.types.directory import TFRecordsDirectory
from flytekit.types.file import TFRecordFile

a = tf.train.Feature(bytes_list=tf.train.BytesList(value=[b"foo", b"bar"]))
b = tf.train.Feature(float_list=tf.train.FloatList(value=[1.0, 2.0]))
c = tf.train.Feature(int64_list=tf.train.Int64List(value=[3, 4]))
d = tf.train.Feature(bytes_list=tf.train.BytesList(value=[b"ham", b"spam"]))
e = tf.train.Feature(float_list=tf.train.FloatList(value=[8.0, 9.0]))
f = tf.train.Feature(int64_list=tf.train.Int64List(value=[22, 23]))
features1 = tf.train.Features(feature=dict(a=a, b=b, c=c))
features2 = tf.train.Features(feature=dict(a=d, b=e, c=f))


def decode_fn(dataset: TFRecordDatasetV2) -> Dict[str, np.ndarray]:
    examples_list = []
    # parse serialised tensors https://www.tensorflow.org/tutorials/load_data/tfrecord#reading_a_tfrecord_file_2
    for batch in list(dataset.as_numpy_iterator()):
        example = tf.train.Example()
        example.ParseFromString(batch)
        examples_list.append(example)
    result = {}
    for a in examples_list:
        # convert example to dict of numpy arrays
        # https://www.tensorflow.org/tutorials/load_data/tfrecord#reading_a_tfrecord_file_2
        for key, feature in a.features.feature.items():
            kind = feature.WhichOneof("kind")
            val = np.array(getattr(feature, kind).value)
            if key not in result.keys():
                result[key] = val
            else:
                result.update({key: np.concatenate((result[key], val))})
    return result


@task
def generate_tf_record_file() -> TFRecordFile:
    return tf.train.Example(features=features1)


@task
def generate_tf_record_dir() -> TFRecordsDirectory:
    return [tf.train.Example(features=features1), tf.train.Example(features=features2)]


@task
def t1(
    dataset: Annotated[
        TFRecordFile,
        TFRecordDatasetConfig(buffer_size=1024, num_parallel_reads=3, compression_type="GZIP"),
    ],
):
    assert isinstance(dataset, TFRecordDatasetV2)
    assert dataset._compression_type == "GZIP"
    assert dataset._buffer_size == 1024
    assert dataset._num_parallel_reads == 3


@task
def t2(dataset: TFRecordFile):
    # if not annotated with TFRecordDatasetConfig, all attributes should default to None
    assert isinstance(dataset, TFRecordDatasetV2)
    assert dataset._compression_type is None
    assert dataset._buffer_size is None
    assert dataset._num_parallel_reads is None


@task
def t3(dataset: TFRecordsDirectory):
    # if not annotated with TFRecordDatasetConfig, all attributes should default to None
    assert isinstance(dataset, TFRecordDatasetV2)
    assert dataset._compression_type is None
    assert dataset._buffer_size is None
    assert dataset._num_parallel_reads is None


@task
def t4(dataset: Annotated[TFRecordFile, TFRecordDatasetConfig(buffer_size=1024)]) -> Dict[str, np.ndarray]:
    return decode_fn(dataset)


@task
def t5(dataset: Annotated[TFRecordsDirectory, TFRecordDatasetConfig(buffer_size=1024)]) -> Dict[str, np.ndarray]:
    return decode_fn(dataset)


@workflow
def wf() -> Tuple[Dict[str, np.ndarray], Dict[str, np.ndarray]]:
    file = generate_tf_record_file()
    files = generate_tf_record_dir()
    t1(dataset=file)
    t2(dataset=file)
    t3(dataset=files)
    files_res = t4(dataset=file)
    dir_res = t5(dataset=files)
    return files_res, dir_res


def test_wf():
    file_res, dir_res = wf()
    assert np.array_equal(file_res["a"], np.array([b"foo", b"bar"]))
    assert np.array_equal(file_res["b"], np.array([1.0, 2.0]))
    assert np.array_equal(file_res["c"], np.array([3, 4]))

    assert np.array_equal(np.sort(dir_res["a"]), np.array([b"bar", b"foo", b"ham", b"spam"]))
    assert np.array_equal(np.sort(dir_res["b"]), np.array([1.0, 2.0, 8.0, 9.0]))
    assert np.array_equal(np.sort(dir_res["c"]), np.array([3, 4, 22, 23]))
