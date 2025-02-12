(tensorflow_type)=

# TensorFlow types

```{eval-rst}
.. tags:: MachineLearning, Basic
```

This document outlines the TensorFlow types available in Flyte, which facilitate the integration of TensorFlow models and datasets in Flyte workflows.

### Import necessary libraries and modules
```{literalinclude} /examples/data_types_and_io/data_types_and_io/tensorflow_type.py
:caption: data_types_and_io/tensorflow_type.py
:lines: 3-12
```

## Tensorflow model
Flyte supports the TensorFlow SavedModel format for serializing and deserializing `tf.keras.Model` instances. The `TensorFlowModelTransformer` is responsible for handling these transformations.

### Transformer
- **Name:** TensorFlow Model
- **Class:** `TensorFlowModelTransformer`
- **Python Type:** `tf.keras.Model`
- **Blob Format:** `TensorFlowModel`
- **Dimensionality:** `MULTIPART`

### Usage
The `TensorFlowModelTransformer` allows you to save a TensorFlow model to a remote location and retrieve it later in your Flyte workflows.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```
```{literalinclude} /examples/data_types_and_io/data_types_and_io/tensorflow_type.py
:caption: data_types_and_io/tensorflow_type.py
:lines: 16-34
```

## TFRecord files
Flyte supports TFRecord files through the TFRecordFile type, which can handle serialized TensorFlow records. The TensorFlowRecordFileTransformer manages the conversion of TFRecord files to and from Flyte literals.

### Transformer
- **Name:** TensorFlow Record File
- **Class:** `TensorFlowRecordFileTransformer`
- **Blob Format:** `TensorFlowRecord`
- **Dimensionality:** `SINGLE`

### Usage
The `TensorFlowRecordFileTransformer` enables you to work with single TFRecord files, making it easy to read and write data in TensorFlow's TFRecord format.

```{literalinclude} /examples/data_types_and_io/data_types_and_io/tensorflow_type.py
:caption: data_types_and_io/tensorflow_type.py
:lines: 38-48
```

## TFRecord directories
Flyte supports directories containing multiple TFRecord files through the `TFRecordsDirectory type`. The `TensorFlowRecordsDirTransformer` manages the conversion of TFRecord directories to and from Flyte literals.

### Transformer
- **Name:** TensorFlow Record Directory
- **Class:** `TensorFlowRecordsDirTransformer`
- **Python Type:** `TFRecordsDirectory`
- **Blob Format:** `TensorFlowRecord`
- **Dimensionality:** `MULTIPART`

### Usage
The `TensorFlowRecordsDirTransformer` allows you to work with directories of TFRecord files, which is useful for handling large datasets that are split across multiple files.

#### Example
```{literalinclude} /examples/data_types_and_io/data_types_and_io/tensorflow_type.py
:caption: data_types_and_io/tensorflow_type.py
:lines: 52-62
```

## Configuration class: `TFRecordDatasetConfig`
The `TFRecordDatasetConfig` class is a data structure used to configure the parameters for creating a `tf.data.TFRecordDataset`, which allows for efficient reading of TFRecord files. This class uses the `DataClassJsonMixin` for easy JSON serialization.

### Attributes
- **compression_type**: (Optional) Specifies the compression method used for the TFRecord files. Possible values include an empty string (no compression), "ZLIB", or "GZIP".
- **buffer_size**: (Optional) Defines the size of the read buffer in bytes. If not set, defaults will be used based on the local or remote file system.
- **num_parallel_reads**: (Optional) Determines the number of files to read in parallel. A value greater than one outputs records in an interleaved order.
- **name**: (Optional) Assigns a name to the operation for easier identification in the pipeline.

This configuration is crucial for optimizing the reading process of TFRecord datasets, especially when dealing with large datasets or when specific performance tuning is required.
